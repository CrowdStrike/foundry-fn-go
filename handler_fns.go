package fdk

import (
	"context"
	"encoding/json"
	"net/http"
)

// HandlerFn wraps a function to return a handler. Similar to the http.HandlerFunc.
type HandlerFn func(ctx context.Context, r Request) Response

// Handle is the request/response lifecycle handler.
func (h HandlerFn) Handle(ctx context.Context, r Request) Response {
	return h(ctx, r)
}

// HandleFnOf provides a means to translate the incoming requests to the destination body type.
// This normalizes the sad path and provides the caller with a zero fuss request to work with. Reducing
// json boilerplate for what is essentially the same operation on different types.
func HandleFnOf[T any](fn func(ctx context.Context, r RequestOf[T]) Response) Handler {
	return HandlerFn(func(ctx context.Context, r Request) Response {
		var v T
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			return Response{Errors: []APIError{{Code: http.StatusBadRequest, Message: "failed to unmarshal payload: " + err.Error()}}}
		}

		return fn(ctx, RequestOf[T]{
			FnID:        r.FnID,
			FnVersion:   r.FnVersion,
			Body:        v,
			Context:     r.Context,
			Headers:     r.Headers,
			Queries:     r.Queries,
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
func HandlerFnOfOK[T interface{ OK() []APIError }](fn func(ctx context.Context, r RequestOf[T]) Response) Handler {
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
func HandleWorkflow(fn func(ctx context.Context, r Request, wrkCtx WorkflowCtx) Response) Handler {
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
func HandleWorkflowOf[T any](fn func(ctx context.Context, r RequestOf[T], wrkCtx WorkflowCtx) Response) Handler {
	return HandleWorkflow(func(ctx context.Context, r Request, workflowCtx WorkflowCtx) Response {
		next := HandleFnOf(func(ctx context.Context, r RequestOf[T]) Response {
			return fn(ctx, r, workflowCtx)
		})
		return next.Handle(ctx, r)
	})
}

// ErrHandler creates a new handler to respond with only errors.
func ErrHandler(errs ...APIError) Handler {
	return HandlerFn(func(ctx context.Context, r Request) Response {
		return ErrResp(errs...)
	})
}
