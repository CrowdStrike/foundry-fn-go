package fdk_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func TestRun_httprunner(t *testing.T) {
	type testReq struct {
		AccessToken string          `json:"access_token"`
		Body        json.RawMessage `json:"body"`
		Context     json.RawMessage `json:"context"`
		FnID        string          `json:"fn_id"`
		FnVersion   int             `json:"fn_version"`
		Method      string          `json:"method"`
		Header      http.Header     `json:"header"`
		Query       url.Values      `json:"query"`
		URL         string          `json:"url"`
		TraceID     string          `json:"trace_id"`
	}

	t.Run("when executing provided handler with successful startup", func(t *testing.T) {
		type (
			inputs struct {
				accessToken string
				body        []byte
				config      string
				configFile  string
				context     []byte
				headers     http.Header
				method      string
				path        string
				queryParams url.Values
				traceID     string
			}

			wantFn func(t *testing.T, resp *http.Response, got respBody)
		)

		tests := []struct {
			name         string
			inputs       inputs
			newHandlerFn func(ctx context.Context, cfg config) fdk.Handler
			want         wantFn
		}{
			{
				name: "simple DELETE request should pass",
				inputs: inputs{
					config: `{"string": "val","integer": 1}`,
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method:      "DELETE",
					path:        "/path",
					queryParams: url.Values{"ids": []string{"id1"}},
					traceID:     "trace1",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Delete("/path", newSimpleHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)
					fdk.EqualVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					fdk.EqualVals(t, config{Str: "val", Int: 1}, echo.Config)

					fdk.EqualVals(t, "/path", echo.Req.Path)
					fdk.EqualVals(t, "DELETE", echo.Req.Method)
					fdk.EqualVals(t, "id1", echo.Req.Queries.Get("ids"))
					fdk.EqualVals(t, "trace1", echo.Req.TraceID)
					fdk.EqualVals(t, "trace1", echo.Req.CtxTraceID)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, echo.Req.Headers)
				},
			},
			{
				name: "simple DELETE request with yaml config should pass",
				inputs: inputs{
					config: `
string: val
integer: 1`,
					configFile: "config.yaml",
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method:      "DELETE",
					path:        "/path",
					queryParams: url.Values{"ids": []string{"id1"}},
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Delete("/path", newSimpleHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)
					fdk.EqualVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					fdk.EqualVals(t, config{Str: "val", Int: 1}, echo.Config)

					fdk.EqualVals(t, "/path", echo.Req.Path)
					fdk.EqualVals(t, "DELETE", echo.Req.Method)
					fdk.EqualVals(t, "id1", echo.Req.Queries.Get("ids"))

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, echo.Req.Headers)
				},
			},
			{
				name: "simple DELETE request with yml config should pass",
				inputs: inputs{
					config: `
string: val
integer: 1`,
					configFile: "config.yml",
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method:      "DELETE",
					path:        "/path",
					queryParams: url.Values{"ids": []string{"id1"}},
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Delete("/path", newSimpleHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)
					fdk.EqualVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					fdk.EqualVals(t, config{Str: "val", Int: 1}, echo.Config)

					fdk.EqualVals(t, "/path", echo.Req.Path)
					fdk.EqualVals(t, "DELETE", echo.Req.Method)
					fdk.EqualVals(t, "id1", echo.Req.Queries.Get("ids"))

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, echo.Req.Headers)
				},
			},
			{
				name: "simple GET request should pass",
				inputs: inputs{
					config: `{"string": "val","integer": 1}`,
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method:      "GET",
					path:        "/path",
					queryParams: url.Values{"bar": []string{"baz"}},
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Get("/path", newSimpleHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)
					fdk.EqualVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					fdk.EqualVals(t, config{Str: "val", Int: 1}, echo.Config)

					fdk.EqualVals(t, "GET", echo.Req.Method)
					fdk.EqualVals(t, "/path", echo.Req.Path)
					fdk.EqualVals(t, "baz", echo.Req.Queries.Get("bar"))

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, echo.Req.Headers)
				},
			},
			{
				name: "simple POST request should pass",
				inputs: inputs{
					body:    []byte(`{"dodgers":"stink"}`),
					context: []byte(`{"kings":"stink_too"}`),
					config:  `{"string": "val","integer": 1}`,
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method: "POST",
					path:   "/path",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Post("/path", newJSONBodyHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)
					fdk.EqualVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					fdk.EqualVals(t, `{"dodgers":"stink"}`, string(echo.Req.Body))
					fdk.EqualVals(t, `{"kings":"stink_too"}`, string(echo.Req.Context))
					fdk.EqualVals(t, config{Str: "val", Int: 1}, echo.Config)
					fdk.EqualVals(t, "/path", echo.Req.Path)
					fdk.EqualVals(t, "POST", echo.Req.Method)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, echo.Req.Headers)
				},
			},
			{
				name: "simple POST request with handler ok should pass",
				inputs: inputs{
					body:    []byte(`{"dodgers":"stink"}`),
					context: []byte(`{"kings":"stink_too"}`),
					config:  `{"string": "val","integer": 1}`,
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method: "POST",
					path:   "/path",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Post("/path", newJSONBodyHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)
					fdk.EqualVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					fdk.EqualVals(t, `{"dodgers":"stink"}`, string(echo.Req.Body))
					fdk.EqualVals(t, `{"kings":"stink_too"}`, string(echo.Req.Context))
					fdk.EqualVals(t, config{Str: "val", Int: 1}, echo.Config)
					fdk.EqualVals(t, "/path", echo.Req.Path)
					fdk.EqualVals(t, "POST", echo.Req.Method)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, echo.Req.Headers)
				},
			},
			{
				name: "simple PUT request should pass",
				inputs: inputs{
					body:   []byte(`{"dodgers":"still stink"}`),
					config: `{"string": "val","integer": 1}`,
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method: "PUT",
					path:   "/path",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Put("/path", newJSONBodyHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)
					fdk.EqualVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					fdk.EqualVals(t, `{"dodgers":"still stink"}`, string(echo.Req.Body))
					fdk.EqualVals(t, config{Str: "val", Int: 1}, echo.Config)
					fdk.EqualVals(t, "/path", echo.Req.Path)
					fdk.EqualVals(t, "PUT", echo.Req.Method)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, echo.Req.Headers)
				},
			},
			{
				name: "POST to a multi path and method handler should pass",
				inputs: inputs{
					body:   []byte(`{"dodgers":"stink"}`),
					config: `{"string": "val","integer": 1}`,
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method: "POST",
					path:   "/path",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Get("/foo", newJSONBodyHandler(cfg))
					m.Delete("/path", newJSONBodyHandler(cfg))
					m.Get("/path", newJSONBodyHandler(cfg))
					m.Post("/path", newJSONBodyHandler(cfg))
					m.Put("/path", newJSONBodyHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)
					fdk.EqualVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					fdk.EqualVals(t, `{"dodgers":"stink"}`, string(echo.Req.Body))
					fdk.EqualVals(t, config{Str: "val", Int: 1}, echo.Config)
					fdk.EqualVals(t, "/path", echo.Req.Path)
					fdk.EqualVals(t, "POST", echo.Req.Method)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, echo.Req.Headers)

				},
			},
			{
				name: "hitting endpoint with no matching route but with matching method should fail with not found",
				inputs: inputs{
					config: `{"string": "val","integer": 1}`,
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method: "DELETE",
					path:   "/missing",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Delete("/found", newJSONBodyHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 404, resp.StatusCode)
					fdk.EqualVals(t, 404, got.Code)

					if len(got.Errs) != 1 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t1 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErr := fdk.APIError{Code: http.StatusNotFound, Message: "route not found"}
					fdk.EqualVals(t, wantErr, got.Errs[0])
				},
			},
			{
				name: "hitting endpoint with matching route and no matching method should fail with method not allowed",
				inputs: inputs{
					config: `{"string": "val","integer": 1}`,
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method: "DELETE",
					path:   "/should-be-get",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Get("/should-be-get", newJSONBodyHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 405, resp.StatusCode)
					fdk.EqualVals(t, 405, got.Code)

					if len(got.Errs) != 1 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t1 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErr := fdk.APIError{Code: http.StatusMethodNotAllowed, Message: "method not allowed"}
					fdk.EqualVals(t, wantErr, got.Errs[0])
				},
			},
			{
				name: "an invalid config should fail",
				inputs: inputs{
					config: `{"string": "wrong string"}`,
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method: "GET",
					path:   "/path",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Get("/path", newJSONBodyHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 400, resp.StatusCode)
					fdk.EqualVals(t, 400, got.Code)

					if len(got.Errs) != 1 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t1 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErr := fdk.APIError{
						Code:    http.StatusBadRequest,
						Message: "config is invalid: invalid config \"string\" field received: wrong string\ninvalid config \"integer\" field received: 0",
					}
					fdk.EqualVals(t, wantErr, got.Errs[0])
				},
			},
			{
				name: "encountering error in new handler",
				inputs: inputs{
					config: `{"string": "val","integer": 1,"err":true}`,
					body:   []byte(`{"should":"ignore"}`),
					headers: http.Header{
						"X-Cs-Origin":      []string{"fooorigin"},
						"X-Cs-Executionid": []string{"exec_id"},
						"X-Cs-Traceid":     []string{"trace_id"},
					},
					method: "POST",
					path:   "/path",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Post("/path", newJSONBodyHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, http.StatusInternalServerError, resp.StatusCode)
					fdk.EqualVals(t, http.StatusInternalServerError, got.Code)

					if len(got.Errs) != 1 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t1 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErr := fdk.APIError{
						Code:    http.StatusInternalServerError,
						Message: "gots the error",
					}
					fdk.EqualVals(t, wantErr, got.Errs[0])
				},
			},
			{
				name: "encountering errors in new handler will return highest error code",
				inputs: inputs{
					config: `{"string": "val","integer": 1,"err":true}`,
					body:   []byte(`{"should":"ignore"}`),
					method: "POST",
					path:   "/path",
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					return fdk.ErrHandler(
						fdk.APIError{Code: 500, Message: "internal server error"},
						fdk.APIError{Code: 501, Message: "even higher"},
						fdk.APIError{Code: 400, Message: "some user error"},
					)
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 501, resp.StatusCode)
					fdk.EqualVals(t, 501, got.Code)

					if len(got.Errs) != 3 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t3 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErrs := []fdk.APIError{
						{Code: 500, Message: "internal server error"},
						{Code: 501, Message: "even higher"},
						{Code: 400, Message: "some user error"},
					}
					fdk.EqualVals(t, len(wantErrs), len(got.Errs))
					for i, want := range wantErrs {
						fdk.EqualVals(t, want, got.Errs[i])
					}
				},
			},
		}

		for _, tt := range tests {
			fn := func(t *testing.T) {
				if tt.inputs.config != "" {
					writeConfigFile(t, tt.inputs.config, tt.inputs.configFile)
				}

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				addr := newServer(ctx, t, func(ctx context.Context, _ *slog.Logger, cfg config) fdk.Handler {
					return tt.newHandlerFn(ctx, cfg)
				})

				body := testReq{
					Body:        tt.inputs.body,
					Header:      tt.inputs.headers,
					Query:       tt.inputs.queryParams,
					URL:         tt.inputs.path,
					Method:      tt.inputs.method,
					Context:     tt.inputs.context,
					AccessToken: tt.inputs.accessToken,
					TraceID:     tt.inputs.traceID,
				}

				b, err := json.Marshal(body)
				mustNoErr(t, err)

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, bytes.NewBuffer(b))
				mustNoErr(t, err)

				resp, err := http.DefaultClient.Do(req)
				mustNoErr(t, err)
				cancel()
				defer func() { _ = resp.Body.Close() }()

				var got respBody
				decodeBody(t, resp.Body, &got)

				tt.want(t, resp, got)
			}
			t.Run(tt.name, fn)
		}
	})

	t.Run("when calling healthz handlers", func(t *testing.T) {
		type (
			inputs struct {
				fnID           string
				fnBuildVersion int
				fnVersion      int
				config         string
			}

			respGeneric struct {
				Code int             `json:"code"`
				Errs []fdk.APIError  `json:"errors"`
				Body json.RawMessage `json:"body"`
			}

			wantFn func(t *testing.T, resp *http.Response, got respGeneric)
		)

		tests := []struct {
			name         string
			inputs       inputs
			newHandlerFn func(ctx context.Context, cfg config) fdk.Handler
			want         wantFn
		}{
			{
				name: "hitting default healthz endpoint should return expected data",
				inputs: inputs{
					fnID:           "id1",
					fnBuildVersion: 1,
					fnVersion:      2,
					config:         `{"string": "val","integer": 1}`,
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					return fdk.NewMux()
				},
				want: func(t *testing.T, resp *http.Response, got respGeneric) {
					fdk.EqualVals(t, 200, resp.StatusCode)
					fdk.EqualVals(t, 200, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					var gotBody map[string]string
					decodeJSON(t, got.Body, &gotBody)

					fdk.EqualVals(t, "ok", gotBody["status"])
					fdk.EqualVals(t, "id1", gotBody["fn_id"])
					fdk.EqualVals(t, "1", gotBody["fn_build_version"])
					fdk.EqualVals(t, "2", gotBody["fn_version"])
				},
			},
			{
				name: "when providing healthz endpoint should use provided healthz endpoint",
				inputs: inputs{
					config: `{"string": "val","integer": 1}`,
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Get("/healthz", fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
						return fdk.Response{
							Code: http.StatusAccepted,
							Body: fdk.JSON("ok"),
						}
					}))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respGeneric) {
					fdk.EqualVals(t, 202, resp.StatusCode)
					fdk.EqualVals(t, 202, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					var gotBody string
					decodeJSON(t, got.Body, &gotBody)

					fdk.EqualVals(t, "ok", gotBody)
				},
			},
			{
				name: "when config is invalid healthz endpoint should return errors",
				inputs: inputs{
					config: `{"string": "","integer": 0}`,
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					return fdk.NewMux()
				},
				want: func(t *testing.T, resp *http.Response, got respGeneric) {
					fdk.EqualVals(t, 400, resp.StatusCode)
					fdk.EqualVals(t, 400, got.Code)

					wantErrs := []fdk.APIError{
						{Code: 400, Message: "config is invalid: invalid config \"string\" field received: \ninvalid config \"integer\" field received: 0"},
					}
					fdk.EqualVals(t, len(wantErrs), len(got.Errs))
					for i, want := range wantErrs {
						fdk.EqualVals(t, want, got.Errs[i])
					}
				},
			},
		}

		for _, tt := range tests {
			fn := func(t *testing.T) {
				if tt.inputs.config != "" {
					writeConfigFile(t, tt.inputs.config, "")
				}

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				addr := newServer(ctx, t, func(ctx context.Context, _ *slog.Logger, cfg config) fdk.Handler {
					return tt.newHandlerFn(ctx, cfg)
				})

				t.Setenv("CS_FN_BUILD_VERSION", strconv.Itoa(tt.inputs.fnBuildVersion))

				b, err := json.Marshal(testReq{
					FnID:      tt.inputs.fnID,
					FnVersion: tt.inputs.fnVersion,
					URL:       "/healthz",
					Method:    http.MethodGet,
				})
				mustNoErr(t, err)

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, bytes.NewBuffer(b))
				mustNoErr(t, err)

				resp, err := http.DefaultClient.Do(req)
				mustNoErr(t, err)
				cancel()
				defer func() { _ = resp.Body.Close() }()

				var got respGeneric
				decodeBody(t, resp.Body, &got)

				tt.want(t, resp, got)
			}
			t.Run(tt.name, fn)
		}
	})

	t.Run("when executing with workflow integration", func(t *testing.T) {
		type (
			inputs struct {
				body    []byte
				context []byte
				method  string
				path    string
			}

			wantFn func(t *testing.T, resp *http.Response, status int, workflowCtx fdk.WorkflowCtx, errs []fdk.APIError)
		)

		tests := []struct {
			name         string
			inputs       inputs
			newHandlerFn func(ctx context.Context, cfg fdk.SkipCfg) fdk.Handler
			want         wantFn
		}{
			{
				name: "with workflow integration and GET request should pass",
				inputs: inputs{
					context: []byte(`{"app_id":"aPp1","cid":"ciD1"}`),
					method:  "GET",
					path:    "/workflow",
				},
				newHandlerFn: func(ctx context.Context, cfg fdk.SkipCfg) fdk.Handler {
					m := fdk.NewMux()
					m.Get("/workflow", fdk.HandleWorkflow(func(ctx context.Context, r fdk.Request, workflowCtx fdk.WorkflowCtx) fdk.Response {
						return fdk.Response{
							Code: 202,
							Body: fdk.JSON(workflowCtx),
						}
					}))
					return m
				},
				want: func(t *testing.T, resp *http.Response, code int, workflowCtx fdk.WorkflowCtx, errs []fdk.APIError) {
					fdk.EqualVals(t, 202, resp.StatusCode)
					fdk.EqualVals(t, 202, code)

					if len(errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", errs)
					}

					want := fdk.WorkflowCtx{AppID: "aPp1", CID: "ciD1"}
					if want != workflowCtx {
						t.Errorf("workflow contexts to not match:\n\t\twant: %#v\n\t\tgot: %#v", want, workflowCtx)
					}
				},
			},
			{
				name: "with workflow integration and POST request should pass",
				inputs: inputs{
					body:    []byte(`{"dodgers":"stink"}`),
					context: []byte(`{"app_id":"aPp1","cid":"ciD1"}`),
					method:  "POST",
					path:    "/workflow",
				},
				newHandlerFn: func(ctx context.Context, cfg fdk.SkipCfg) fdk.Handler {
					m := fdk.NewMux()
					m.Post("/workflow", fdk.HandleWorkflowOf(func(ctx context.Context, r fdk.RequestOf[reqBodyDodgers], workflowCtx fdk.WorkflowCtx) fdk.Response {
						workflowCtx.CID += "-" + r.Body.Dodgers
						return fdk.Response{
							Code: 202,
							Body: fdk.JSON(workflowCtx),
						}
					}))
					return m
				},
				want: func(t *testing.T, resp *http.Response, code int, workflowCtx fdk.WorkflowCtx, errs []fdk.APIError) {
					fdk.EqualVals(t, 202, resp.StatusCode)
					fdk.EqualVals(t, 202, code)

					if len(errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", errs)
					}

					want := fdk.WorkflowCtx{AppID: "aPp1", CID: "ciD1-stink"}
					if want != workflowCtx {
						t.Errorf("workflow contexts to not match:\n\t\twant: %#v\n\t\tgot: %#v", want, workflowCtx)
					}
				},
			},
		}

		for _, tt := range tests {
			fn := func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				addr := newServer(ctx, t, func(ctx context.Context, _ *slog.Logger, cfg fdk.SkipCfg) fdk.Handler {
					return tt.newHandlerFn(ctx, cfg)
				})

				body := testReq{
					Body:    tt.inputs.body,
					URL:     tt.inputs.path,
					Method:  tt.inputs.method,
					Context: tt.inputs.context,
				}

				b, err := json.Marshal(body)
				mustNoErr(t, err)

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, bytes.NewBuffer(b))
				mustNoErr(t, err)

				resp, err := http.DefaultClient.Do(req)
				mustNoErr(t, err)
				cancel()
				defer func() { _ = resp.Body.Close() }()

				var got struct {
					Code        int             `json:"code"`
					Errs        []fdk.APIError  `json:"errors"`
					WorkflowCtx fdk.WorkflowCtx `json:"body"`
				}
				decodeBody(t, resp.Body, &got)

				tt.want(t, resp, got.Code, got.WorkflowCtx, got.Errs)
			}
			t.Run(tt.name, fn)
		}
	})

	t.Run("when executing handler with file response", func(t *testing.T) {
		tmp := t.TempDir()

		newReqBody := func(t *testing.T, r fileInReq) json.RawMessage {
			t.Helper()
			b, err := json.Marshal(r)
			mustNoErr(t, err)
			return b
		}

		type (
			inputs struct {
				body   json.RawMessage
				method string
				path   string
			}

			respBody struct {
				ContentType string `json:"content_type"`
				Encoding    string `json:"encoding"`
				Filename    string `json:"filename"`
				SHA256      string `json:"sha256_checksum"`
				Size        int    `json:"size,string"`
			}

			wantFn func(t *testing.T, resp *http.Response, got respBody)
		)

		tests := []struct {
			name         string
			input        inputs
			newHandlerFn func(ctx context.Context) fdk.Handler
			want         wantFn
		}{
			{
				name: "POST to a endpoint that returns a sdk.File should pass",
				input: inputs{
					body: newReqBody(t, fileInReq{
						ContentType:  "application/json",
						DestFilename: filepath.Join(tmp, "first_file.json"),
						V:            `{"some":"json"}`,
					}),
					method: "POST",
					path:   "/file",
				},
				newHandlerFn: func(ctx context.Context) fdk.Handler {
					m := fdk.NewMux()
					m.Post("/file", fdk.HandleFnOf(newFileHandler))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)

					fdk.EqualVals(t, "application/json", got.ContentType)
					fdk.EqualVals(t, "", got.Encoding)
					fdk.EqualVals(t, "SqgS0EPNEPmkm4NrB9osqbE/bBoalfO9wJFqf3t7FI0=", got.SHA256)
					fdk.EqualVals(t, 15, got.Size)

					wantFilename := filepath.Join(tmp, "first_file.json")
					fdk.EqualVals(t, wantFilename, got.Filename)
					equalFiles(t, got.Filename, `{"some":"json"}`)
				},
			},
			{
				name: "POST to an endpoint that returns a sdk.File with encoding should pass",
				input: inputs{
					body: newReqBody(t, fileInReq{
						ContentType: "application/json",
						// requires file handler to implement the gzip compression, not enforced
						// TODO(@berg): might make sense to add a middleware for compressing files that authors can utilize
						Encoding:     "gzip",
						DestFilename: filepath.Join(tmp, "second_file.json"),
						V:            `{"dodgers":"stink"}`,
					}),
					method: "POST",
					path:   "/file",
				},
				newHandlerFn: func(ctx context.Context) fdk.Handler {
					m := fdk.NewMux()
					m.Post("/file", fdk.HandleFnOf(newFileHandler))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)

					fdk.EqualVals(t, "application/json", got.ContentType)
					fdk.EqualVals(t, "gzip", got.Encoding)
					fdk.EqualVals(t, "AwAr4comqRgfR15a6F1PwfnjGmDMbdOgPe336J+puA4=", got.SHA256)
					fdk.EqualVals(t, 19, got.Size)

					wantFilename := filepath.Join(tmp, "second_file.json")
					fdk.EqualVals(t, wantFilename, got.Filename)
					equalFiles(t, got.Filename, `{"dodgers":"stink"}`)
				},
			},
			{
				name: "POST to an endpoint that returns a gzip compressed sdk.File should pass",
				input: inputs{
					body: newReqBody(t, fileInReq{
						ContentType:  "application/json",
						DestFilename: filepath.Join(tmp, "third_file.json"),
						V:            `{"dodgers":"reallystank"}`,
					}),
					method: "POST",
					path:   "/compress-file",
				},
				newHandlerFn: func(ctx context.Context) fdk.Handler {
					m := fdk.NewMux()
					m.Post("/compress-file", fdk.HandleFnOf(newGzippedFileHandler))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					fdk.EqualVals(t, 201, resp.StatusCode)

					fdk.EqualVals(t, "application/json", got.ContentType)
					fdk.EqualVals(t, "gzip", got.Encoding)
					fdk.EqualVals(t, "H/NpL40Xq6xIVeD5ZOizqzXJzqRRYD4/a9cRG+0dAr0=", got.SHA256)
					fdk.EqualVals(t, 49, got.Size)

					wantFilename := filepath.Join(tmp, "third_file.json")
					fdk.EqualVals(t, wantFilename, got.Filename)
					equalGzipFiles(t, got.Filename, `{"dodgers":"reallystank"}`)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				addr := newServer(ctx, t, func(ctx context.Context, _ *slog.Logger, _ fdk.SkipCfg) fdk.Handler {
					return tt.newHandlerFn(ctx)
				})

				reqBody := testReq{
					Body:   tt.input.body,
					Method: tt.input.method,
					URL:    tt.input.path,
				}

				b, err := json.Marshal(reqBody)
				mustNoErr(t, err)

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, bytes.NewBuffer(b))
				mustNoErr(t, err)

				resp, err := http.DefaultClient.Do(req)
				mustNoErr(t, err)
				cancel()
				defer func() { _ = resp.Body.Close() }()

				b, err = io.ReadAll(resp.Body)
				mustNoErr(t, err)

				var got struct {
					File respBody `json:"body"`
				}
				mustNoErr(t, json.Unmarshal(b, &got))

				tt.want(t, resp, got.File)
			})
		}
	})
}

type config struct {
	Err bool   `json:"err"`
	Str string `json:"string"`
	Int int    `json:"integer"`
}

func (c config) OK() error {
	var errs []error
	if c.Str != "val" {
		errs = append(errs, errors.New(`invalid config "string" field received: `+c.Str))
	}
	if c.Int != 1 {
		errs = append(errs, errors.New(`invalid config "integer" field received: `+strconv.Itoa(c.Int)))
	}

	return errors.Join(errs...)
}

type (
	respBody struct {
		Code    int            `json:"code"`
		Errs    []fdk.APIError `json:"errors"`
		Headers http.Header    `json:"headers"`
		Req     echoReq        `json:"body"`
	}

	echoReq struct {
		Config config     `json:"config"`
		Req    echoInputs `json:"request"`
	}

	echoInputs struct {
		Body        json.RawMessage `json:"body,omitempty"`
		Context     json.RawMessage `json:"context,omitempty"`
		Headers     http.Header     `json:"header"`
		Queries     url.Values      `json:"query"`
		Path        string          `json:"path"`
		Method      string          `json:"method"`
		AccessToken string          `json:"access_token"`
		TraceID     string          `json:"trace_id"`
		CtxTraceID  string          `json:"ctx_trace_id"`
	}

	reqBodyDodgers struct {
		Dodgers string `json:"dodgers"`
	}
)

func newSimpleHandler(cfg config) fdk.Handler {
	if cfg.Err {
		return fdk.ErrHandler(fdk.APIError{
			Code:    http.StatusInternalServerError,
			Message: "gots the error",
		})
	}
	return fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		return newEchoResp(ctx, cfg, r)
	})
}

func newJSONBodyHandler(cfg config) fdk.Handler {
	if cfg.Err {
		return fdk.ErrHandler(fdk.APIError{
			Code:    http.StatusInternalServerError,
			Message: "gots the error",
		})
	}
	return fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		return newEchoResp(ctx, cfg, r)
	})
}

func newEchoResp(ctx context.Context, cfg config, r fdk.Request) fdk.Response {
	traceID, _ := ctx.Value("_traceid").(string)
	bodyB, _ := io.ReadAll(r.Body)
	return fdk.Response{

		Body: fdk.JSON(echoReq{
			Config: cfg,
			Req: echoInputs{
				Body:        bodyB,
				Context:     r.Context,
				Headers:     r.Headers,
				Queries:     r.Queries,
				Path:        r.URL,
				Method:      r.Method,
				AccessToken: r.AccessToken,
				TraceID:     r.TraceID,
				CtxTraceID:  traceID,
			},
		}),
		Code:   201,
		Header: http.Header{"X-Foo": []string{"foo"}},
	}
}

type fileInReq struct {
	ContentType  string `json:"content_type"`
	DestFilename string `json:"destination_filename"`
	Encoding     string `json:"encoding"`
	V            string `json:"value"`
}

func newFileHandler(_ context.Context, r fdk.RequestOf[fileInReq]) fdk.Response {
	return fdk.Response{
		Code: 201,
		Body: fdk.File{
			ContentType: r.Body.ContentType,
			Encoding:    r.Body.Encoding,
			Filename:    r.Body.DestFilename,
			Contents:    io.NopCloser(strings.NewReader(r.Body.V)),
		},
	}
}

func newGzippedFileHandler(_ context.Context, r fdk.RequestOf[fileInReq]) fdk.Response {
	return fdk.Response{
		Code: 201,
		Body: fdk.CompressGzip(fdk.File{
			ContentType: r.Body.ContentType,
			Encoding:    r.Body.Encoding,
			Filename:    r.Body.DestFilename,
			Contents:    io.NopCloser(strings.NewReader(r.Body.V)),
		}),
	}
}

func equalFiles(t testing.TB, filename string, want string) {
	t.Helper()

	f, err := os.Open(filename)
	mustNoErr(t, err)
	defer func() { _ = f.Close() }()

	equalReader(t, want, f)

	err = f.Close()
	if err != nil {
		t.Errorf("failed to close file: " + err.Error())
	}
}

func equalGzipFiles(t testing.TB, filename string, want string) {
	t.Helper()

	f, err := os.Open(filename)
	mustNoErr(t, err)
	defer func() { _ = f.Close() }()

	gr, err := gzip.NewReader(f)
	mustNoErr(t, err)

	equalReader(t, want, gr)
}

func equalReader(t testing.TB, want string, got io.Reader) {
	t.Helper()

	b, err := io.ReadAll(got)
	mustNoErr(t, err)

	fdk.EqualVals(t, want, string(b))
}

func mustNoErr(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("received unexpected err:\n\t\tgot:\t%s", err)
	}
}

func containsHeaders(t testing.TB, want http.Header, got http.Header) {
	t.Helper()

	failed := false
	defer func() {
		if failed {
			t.Logf("received headers:\t%v", got)
		}
	}()

	for h, vals := range want {
		if gotV := got[h]; !reflect.DeepEqual(vals, gotV) {
			t.Errorf("header values don't match:\n\t\theader name:\t%s\n\t\twant:\t%v\n\t\tgot:\t%v", h, vals, gotV)
			failed = true
		}
	}
}

func decodeBody(t testing.TB, r io.Reader, v any) {
	t.Helper()

	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal("failed to read: " + err.Error())
	}

	decodeJSON(t, b, v)
}

func decodeJSON(t testing.TB, b []byte, v any) {
	t.Helper()

	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("failed to unmarshal json: %s\n\t\tpayload:\t%s", err, string(b))
	}
}

func newServer[CFG fdk.Cfg](ctx context.Context, t *testing.T, newHandlerFn func(context.Context, *slog.Logger, CFG) fdk.Handler) string {
	t.Helper()

	port := newIP(t)
	t.Setenv("PORT", port)

	readyChan := make(chan struct{})

	done := make(chan struct{})
	go func() {
		defer close(done)
		fdk.Run(ctx, func(ctx context.Context, logger *slog.Logger, cfg CFG) fdk.Handler {
			h := newHandlerFn(ctx, logger, cfg)
			close(readyChan)
			return h
		})
	}()

	select {
	case <-readyChan:
	case <-time.After(200 * time.Millisecond):
	}

	return "http://localhost:" + port
}

func newIP(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:")
	mustNoErr(t, err)
	defer func() {
		mustNoErr(t, l.Close())
	}()

	parts := strings.Split(l.Addr().String(), ":")
	if len(parts) == 1 {
		t.Fatal("invalid parts returned: ", parts)
	}
	return parts[len(parts)-1]
}

func writeConfigFile(t *testing.T, config, cfgFile string) {
	t.Helper()

	if cfgFile == "" {
		cfgFile = "config.json"
	}
	tmp := filepath.Join(t.TempDir(), cfgFile)
	t.Setenv("CS_FN_CONFIG_PATH", tmp)

	err := os.WriteFile(tmp, []byte(config), 0666)
	mustNoErr(t, err)
}
