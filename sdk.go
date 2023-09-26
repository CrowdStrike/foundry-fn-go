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

type HandlerFn func(ctx context.Context, r Request) Response

func (h HandlerFn) Handle(ctx context.Context, r Request) Response {
	return h(ctx, r)
}

// HandleFnOf provides a means to translate the incoming requests to the destination body type.
// This normalizes the sad path and provides the caller with a zero fuss request to work with. Reducing
// json boilerplate for what is essentially the same operation on different types.
//
// TODO(berg): name could be HandlerOf perhaps?
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
	// Request a word for Request. We can treat it as an internal impl details, and provide the
	// user an http.Request, built from this internal impl. They can then build up
	// the usual http.ServeHTTP, the same as they are likely accustomed too. We too
	// can take advantage of this and add middleware,etc to the runner as needed. When testing
	// locally, they just use curl, and they provide it the way they are use too. For our lambda
	// impl, we convert it to the expected request body the http server expects, then its treated
	// the same as our standard impl.
	Request RequestOf[json.RawMessage]

	// RequestOf provides a generic body we can target our unmarshaling into. We don't have to have this
	// as some users may be happy with the OG Request. However, we can express the former with the latter
	// here.
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

	Response struct {
		Body   json.Marshaler
		Code   int
		Errors []APIError
		Header http.Header
	}

	APIError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

// Run is the meat and potatoes. This is the entrypoint for everything.
func Run[T Cfg](ctx context.Context, newHandlerFn func(_ context.Context, cfg T) Handler) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic caught", "stack_trace", string(debug.Stack()))
			run(ctx, logger, HandlerFn(func(ctx context.Context, r Request) Response {
				return Response{
					Errors: []APIError{{
						Code:    http.StatusServiceUnavailable,
						Message: "encountered unexpected error",
					}},
				}
			}))
		}
	}()

	cfg, apiErr := readCfg[T](ctx, logger)
	if apiErr != nil {
		run(ctx, logger, HandlerFn(func(ctx context.Context, r Request) Response {
			return Response{Errors: []APIError{*apiErr}}
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
