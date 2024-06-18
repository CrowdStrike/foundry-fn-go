package fdk

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
)

// TestRun executes the handler in order to test it through the HTTP runner.
func TestRun[T Cfg](ctx context.Context, cfg T, newHandlerFn func(_ context.Context, cfg T) Handler) (testHandlerFn func(context.Context, Request) Response) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	mux := newHTTPRunMux(ctx, logger, newHandlerFn(ctx, cfg))

	return func(ctx context.Context, r Request) Response {
		b, err := json.Marshal(httpPayload{
			Body:        r.Body,
			Context:     r.Context,
			AccessToken: r.AccessToken,
			Method:      r.Method,
			Params: struct {
				Header http.Header `json:"header"`
				Query  url.Values  `json:"query"`
			}{Header: r.Params.Header, Query: r.Params.Query},
			URL:     r.URL,
			TraceID: r.TraceID,
		})
		if err != nil {
			return ErrResp(APIError{Code: http.StatusInternalServerError, Message: "unable to create http request body: " + err.Error()})
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewReader(b))
		if err != nil {
			return ErrResp(APIError{Code: http.StatusInternalServerError, Message: "unable to create http request: " + err.Error()})
		}

		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		var respBody struct {
			Body    json.RawMessage `json:"body,omitempty"`
			Code    int             `json:"code"`
			Errs    []APIError      `json:"errors"`
			Headers http.Header     `json:"headers"`
		}
		err = json.Unmarshal(rec.Body.Bytes(), &respBody)
		if err != nil {
			return ErrResp(APIError{Code: http.StatusInternalServerError, Message: "unable to unmarshal response body: " + err.Error()})
		}

		return Response{
			Body:   respBody.Body,
			Code:   respBody.Code,
			Errors: respBody.Errs,
			Header: respBody.Headers,
		}
	}
}
