package fdk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
)

// Handler provides a handler for our incoming request.
type Handler interface {
	Handle(ctx context.Context, r Request) Response
}

// HandlerFn wraps a function to return a handler. Similar to the http.HandlerFunc.
type HandlerFn func(ctx context.Context, r Request) Response

// Handle is the request/response lifecycle handler.
func (h HandlerFn) Handle(ctx context.Context, r Request) Response {
	return h(ctx, r)
}

// HandleFnOf provides a means to translate the incoming requests to the destination body type.
// This normalizes the sad path and provides the caller with a zero fuss request to work with. Reducing
// json boilerplate for what is essentially the same operation on different types.
func HandleFnOf[T any](fn func(context.Context, RequestOf[T]) Response) Handler {
	return HandlerFn(func(ctx context.Context, r Request) Response {
		var v T
		if err := json.Unmarshal(r.Body, &v); err != nil {
			return Response{Errors: []APIError{{Code: http.StatusBadRequest, Message: "failed to unmarshal payload: " + err.Error()}}}
		}

		return fn(ctx, RequestOf[T]{
			Body:        v,
			Context:     r.Context,
			Params:      r.Params,
			URL:         r.URL,
			Method:      r.Method,
			AccessToken: r.AccessToken,
			TraceID:     r.TraceID,
		})
	})
}

// HandlerFnOfOK provides a means to translate the incoming requests to the destination body type
// and execute validation on that type. This normalizes the sad path for both the unmarshalling of
// the request body and the validation of that request type using its OK() method.
func HandlerFnOfOK[T interface{ OK() []APIError }](fn func(context.Context, RequestOf[T]) Response) Handler {
	return HandleFnOf(func(ctx context.Context, r RequestOf[T]) Response {
		if errs := r.Body.OK(); len(errs) > 0 {
			return ErrResp(errs...)
		}
		return fn(ctx, r)
	})
}

// WorkflowCtx is the Request.Context field when integrating a function with Falcon Fusion workflow.
type WorkflowCtx struct {
	ActivityExecID    string `json:"activity_execution_id"`
	AppID             string `json:"app_id"`
	CID               string `json:"cid"`
	OwnerCID          string `json:"owner_cid"`
	DefinitionID      string `json:"definition_id,omitempty"`
	DefinitionVersion int    `json:"definition_version,omitempty"`
	ExecutionID       string `json:"execution_id,omitempty"`
	Activity          struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		NodeID string `json:"node_id"`
	} `json:"activity"`
	Trigger struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"trigger"`
}

// HandleWorkflow provides a means to create a handler with workflow integration. This function
// does not have an opinion on the request body but does expect a workflow integration. Typically,
// this is useful for DELETE/GET handlers.
func HandleWorkflow(fn func(context.Context, Request, WorkflowCtx) Response) Handler {
	return HandlerFn(func(ctx context.Context, r Request) Response {
		var w WorkflowCtx
		if err := json.Unmarshal(r.Context, &w); err != nil {
			return Response{Errors: []APIError{{Code: http.StatusBadRequest, Message: "failed to unmarshal workflow context: " + err.Error()}}}
		}

		return fn(ctx, r, w)
	})
}

// HandleWorkflowOf provides a means to create a handler with Workflow integration. This
// function is useful when you expect a request body and have workflow integrations. Typically, this
// is with PATCH/POST/PUT handlers.
func HandleWorkflowOf[T any](fn func(context.Context, RequestOf[T], WorkflowCtx) Response) Handler {
	return HandleWorkflow(func(ctx context.Context, r Request, workflowCtx WorkflowCtx) Response {
		next := HandleFnOf(func(ctx context.Context, r RequestOf[T]) Response {
			return fn(ctx, r, workflowCtx)
		})
		return next.Handle(ctx, r)
	})
}

type (
	// Request defines a request structure that is given to the runner. The Body is set to
	// json.RawMessage, to enable decoration/middleware.
	Request RequestOf[json.RawMessage]

	// RequestOf provides a generic body we can target our unmarshaling into.
	RequestOf[T any] struct {
		FnID      string
		FnVersion int
		Body      T
		Context   json.RawMessage
		Params    struct {
			Header http.Header
			Query  url.Values
		}
		URL         string
		Method      string
		AccessToken string
		TraceID     string
	}
)

// APIError defines a error that is shared back to the caller.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error provides a human readable error message.
func (a APIError) Error() string {
	return fmt.Sprintf("[%d] %s", a.Code, a.Message)
}

// Response is the domain type for the response.
type Response struct {
	Body   json.Marshaler
	Code   int
	Errors []APIError
	Header http.Header
}

// StatusCode returns the response status code. When a Response.Code is not
// set and errors exist, then the highest code on the errors is returned.
func (r Response) StatusCode() int {
	code := r.Code
	if code == 0 && len(r.Errors) > 0 {
		for _, e := range r.Errors {
			if e.Code > code {
				code = e.Code
			}
		}
	}
	if code == 0 {
		code = http.StatusOK
	}
	return code
}

// Run is the meat and potatoes. This is the entrypoint for everything.
func Run[T Cfg](ctx context.Context, newHandlerFn func(_ context.Context, cfg T) Handler) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	var runFn Handler = HandlerFn(func(ctx context.Context, r Request) Response {
		cfg, loadErr := readCfg[T](ctx)
		if loadErr != nil {
			if loadErr.err != nil {
				logger.Error("failed to load config", "err", loadErr.err)
			}
			return ErrResp(loadErr.apiErr)
		}

		h := newHandlerFn(ctx, cfg)

		return h.Handle(ctx, r)
	})
	runFn = recoverer(logger)(runFn)

	run(ctx, logger, runFn)
}

// JSON jsonifies the input to valid json upon request marshaling.
func JSON(v any) json.Marshaler {
	return jsoned{v: v}
}

type jsoned struct {
	v any
}

func (j jsoned) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.v)
}

// ErrHandler creates a new handler to respond with only errors.
func ErrHandler(errs ...APIError) Handler {
	return HandlerFn(func(ctx context.Context, r Request) Response {
		return ErrResp(errs...)
	})
}

// ErrResp creates a sad path errors only response.
//
// Note: the highest status code from the errors will be used for the response
// status if no status code is set on the response.
func ErrResp(errs ...APIError) Response {
	resp := Response{Errors: errs}
	resp.Code = resp.StatusCode()
	return resp
}

func recoverer(logger *slog.Logger) func(Handler) Handler {
	return func(h Handler) Handler {
		return HandlerFn(func(ctx context.Context, r Request) (resp Response) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic caught", "stack_trace", string(debug.Stack()))
					resp = ErrResp(APIError{Code: http.StatusServiceUnavailable, Message: "encountered unexpected error"})
				}
			}()

			return h.Handle(ctx, r)
		})
	}
}
