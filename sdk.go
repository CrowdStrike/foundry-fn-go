package fdk

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
)

// Fn returns the active function id and version.
func Fn() struct {
	ID      string
	Version int
} {
	v, _ := strconv.Atoi(os.Getenv("CS_FN_VERSION"))
	return struct {
		ID      string
		Version int
	}{
		ID:      os.Getenv("CS_FN_ID"),
		Version: v,
	}
}

// Handler provides a handler for our incoming request.
//
// TODO(berg): I'm a little confused why we have a response, with APIErrors, and
// a go type error being returned in the legacy sdks. This creates
// multiple ways to do the same thing. I'd be confused, as I am now
// I suppose, with what goes where. If we remove the error from the
// return tuple, the only place for errors now is in the Response type.
// I think this makes good sense. Lets not create a failure condition
// from something that could be in user space.
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
		})
	})
}

type (
	// Request defines a request structure that is given to the runner. The Body is set to
	// json.RawMessage, to enable decoration/middleware.
	Request RequestOf[json.RawMessage]

	// RequestOf provides a generic body we can target our unmarshaling into.
	RequestOf[T any] struct {
		Body T
		// TODO(berg): can we axe Context? have workflow put details in the body/headers/params instead?
		Context json.RawMessage
		Params  struct {
			Header http.Header
			Query  url.Values
		}
		// TODO(berg): explore changing this field to Path, as URL is misleading. It's never
		// 			   an fqdn, only the path of the url.
		URL         string
		Method      string
		AccessToken string
	}

	// APIError defines a error that is shared back to the caller.
	APIError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

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
			code = e.Code
		}
	}
	return code
}

// Run is the meat and potatoes. This is the entrypoint for everything.
func Run[T Cfg](ctx context.Context, newHandlerFn func(_ context.Context, cfg T) Handler) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	defer func() {
		if err := recover(); err != nil {
			run(ctx, logger, HandlerFn(func(ctx context.Context, r Request) Response {
				logger.Error("panic caught", "stack_trace", string(debug.Stack()))
				return Response{
					Errors: []APIError{{
						Code:    http.StatusServiceUnavailable,
						Message: "encountered unexpected error",
					}},
				}
			}))
		}
	}()

	cfg, loadErr := readCfg[T](ctx)
	if loadErr != nil {
		run(ctx, logger, HandlerFn(func(ctx context.Context, r Request) Response {
			if loadErr.err != nil {
				logger.Error("failed to load config", "err", loadErr.err)
			}
			return Response{Errors: []APIError{loadErr.apiErr}}
		}))
		return
	}

	h := newHandlerFn(ctx, cfg)

	run(ctx, logger, h)

	return
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

// ErrHandler creates a new handler to resopnd with only errors.
func ErrHandler(errs ...APIError) Handler {
	return HandlerFn(func(ctx context.Context, r Request) Response {
		return Response{Errors: errs}
	})
}
