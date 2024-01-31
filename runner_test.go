package fdk_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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
				},
				newHandlerFn: func(ctx context.Context, cfg config) fdk.Handler {
					m := fdk.NewMux()
					m.Delete("/path", newSimpleHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response, got respBody) {
					equalVals(t, 201, resp.StatusCode)
					equalVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)

					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "DELETE", echo.Req.Method)
					equalVals(t, "id1", echo.Req.Queries.Get("ids"))

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
					equalVals(t, 201, resp.StatusCode)
					equalVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)

					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "DELETE", echo.Req.Method)
					equalVals(t, "id1", echo.Req.Queries.Get("ids"))

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
					equalVals(t, 201, resp.StatusCode)
					equalVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)

					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "DELETE", echo.Req.Method)
					equalVals(t, "id1", echo.Req.Queries.Get("ids"))

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
					equalVals(t, 201, resp.StatusCode)
					equalVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)

					equalVals(t, "GET", echo.Req.Method)
					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "baz", echo.Req.Queries.Get("bar"))

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
					equalVals(t, 201, resp.StatusCode)
					equalVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					equalVals(t, `{"dodgers":"stink"}`, string(echo.Req.Body))
					equalVals(t, `{"kings":"stink_too"}`, string(echo.Req.Context))
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)
					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "POST", echo.Req.Method)

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
					equalVals(t, 201, resp.StatusCode)
					equalVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					equalVals(t, `{"dodgers":"still stink"}`, string(echo.Req.Body))
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)
					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "PUT", echo.Req.Method)

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
					equalVals(t, 201, resp.StatusCode)
					equalVals(t, 201, got.Code)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					equalVals(t, `{"dodgers":"stink"}`, string(echo.Req.Body))
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)
					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "POST", echo.Req.Method)

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
					equalVals(t, 404, resp.StatusCode)
					equalVals(t, 404, got.Code)

					if len(got.Errs) != 1 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t1 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErr := fdk.APIError{Code: http.StatusNotFound, Message: "route not found"}
					equalVals(t, wantErr, got.Errs[0])
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
					equalVals(t, 405, resp.StatusCode)
					equalVals(t, 405, got.Code)

					if len(got.Errs) != 1 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t1 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErr := fdk.APIError{Code: http.StatusMethodNotAllowed, Message: "method not allowed"}
					equalVals(t, wantErr, got.Errs[0])
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
					equalVals(t, 400, resp.StatusCode)
					equalVals(t, 400, got.Code)

					if len(got.Errs) != 1 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t1 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErr := fdk.APIError{
						Code:    http.StatusBadRequest,
						Message: "config is invalid: invalid config \"string\" field received: wrong string\ninvalid config \"integer\" field received: 0",
					}
					equalVals(t, wantErr, got.Errs[0])
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
					equalVals(t, http.StatusInternalServerError, resp.StatusCode)
					equalVals(t, http.StatusInternalServerError, got.Code)

					if len(got.Errs) != 1 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t1 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErr := fdk.APIError{
						Code:    http.StatusInternalServerError,
						Message: "gots the error",
					}
					equalVals(t, wantErr, got.Errs[0])
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
					equalVals(t, 501, resp.StatusCode)
					equalVals(t, 501, got.Code)

					if len(got.Errs) != 3 {
						t.Fatalf("did not received expected number of errors\n\t\twant:\t3 error\n\t\tgot:\t%+v", got.Errs)
					}

					wantErrs := []fdk.APIError{
						{Code: 500, Message: "internal server error"},
						{Code: 501, Message: "even higher"},
						{Code: 400, Message: "some user error"},
					}
					equalVals(t, len(wantErrs), len(got.Errs))
					for i, want := range wantErrs {
						equalVals(t, want, got.Errs[i])
					}
				},
			},
		}

		for _, tt := range tests {
			fn := func(t *testing.T) {
				if tt.inputs.config != "" {
					cfgFile := tt.inputs.configFile
					if cfgFile == "" {
						cfgFile = "config.json"
					}
					tmp := filepath.Join(t.TempDir(), cfgFile)
					t.Setenv("CS_FN_CONFIG_PATH", tmp)

					err := os.WriteFile(tmp, []byte(tt.inputs.config), 0666)
					mustNoErr(t, err)
				}
				port := newIP(t)
				t.Setenv("PORT", port)

				readyChan := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				done := make(chan struct{})
				go func() {
					defer close(done)
					fdk.Run(ctx, func(ctx context.Context, cfg config) fdk.Handler {
						h := tt.newHandlerFn(ctx, cfg)
						close(readyChan)
						return h
					})
				}()

				select {
				case <-readyChan:
				case <-time.After(50 * time.Millisecond):
				}

				body := struct {
					AccessToken string          `json:"access_token"`
					Body        json.RawMessage `json:"body"`
					Context     json.RawMessage `json:"context"`
					Method      string          `json:"method"`
					Params      struct {
						Header http.Header `json:"header"`
						Query  url.Values  `json:"query"`
					} `json:"params"`
					URL string `json:"url"`
				}{
					Body: tt.inputs.body,
					Params: struct {
						Header http.Header `json:"header"`
						Query  url.Values  `json:"query"`
					}{
						Header: tt.inputs.headers,
						Query:  tt.inputs.queryParams,
					},
					URL:         tt.inputs.path,
					Method:      tt.inputs.method,
					Context:     tt.inputs.context,
					AccessToken: tt.inputs.accessToken,
				}

				b, err := json.Marshal(body)
				mustNoErr(t, err)

				req, err := http.NewRequestWithContext(
					ctx,
					http.MethodPost,
					"http://localhost:"+port,
					bytes.NewBuffer(b),
				)
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
					equalVals(t, 202, resp.StatusCode)
					equalVals(t, 202, code)

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
					equalVals(t, 202, resp.StatusCode)
					equalVals(t, 202, code)

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
				port := newIP(t)
				t.Setenv("PORT", port)

				readyChan := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				done := make(chan struct{})
				go func() {
					defer close(done)
					fdk.Run(ctx, func(ctx context.Context, cfg fdk.SkipCfg) fdk.Handler {
						h := tt.newHandlerFn(ctx, cfg)
						close(readyChan)
						return h
					})
				}()

				select {
				case <-readyChan:
				case <-time.After(50 * time.Millisecond):
				}

				body := struct {
					Body    json.RawMessage `json:"body"`
					Context json.RawMessage `json:"context"`
					Method  string          `json:"method"`
					URL     string          `json:"url"`
				}{
					Body:    tt.inputs.body,
					URL:     tt.inputs.path,
					Method:  tt.inputs.method,
					Context: tt.inputs.context,
				}

				b, err := json.Marshal(body)
				mustNoErr(t, err)

				req, err := http.NewRequestWithContext(
					ctx,
					http.MethodPost,
					"http://localhost:"+port,
					bytes.NewBuffer(b),
				)
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
		return newEchoResp(cfg, r)
	})
}

func newJSONBodyHandler(cfg config) fdk.Handler {
	if cfg.Err {
		return fdk.ErrHandler(fdk.APIError{
			Code:    http.StatusInternalServerError,
			Message: "gots the error",
		})
	}
	return fdk.HandleFnOf(func(ctx context.Context, r fdk.RequestOf[json.RawMessage]) fdk.Response {
		return newEchoResp(cfg, fdk.Request(r))
	})
}

func newEchoResp(cfg config, r fdk.Request) fdk.Response {
	return fdk.Response{
		Body: fdk.JSON(echoReq{
			Config: cfg,
			Req: echoInputs{
				Body:        r.Body,
				Context:     r.Context,
				Headers:     r.Params.Header,
				Queries:     r.Params.Query,
				Path:        r.URL,
				Method:      r.Method,
				AccessToken: r.AccessToken,
			},
		}),
		Code:   201,
		Header: http.Header{"X-Foo": []string{"foo"}},
	}
}

func equalVals[T comparable](t testing.TB, want, got T) bool {
	t.Helper()

	match := want == got
	if !match {
		t.Errorf("values not equal:\n\t\twant:\t%#v\n\t\tgot:\t%#v", want, got)
	}
	return match
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

	err = json.Unmarshal(b, v)
	if err != nil {
		t.Fatalf("failed to unmarshal json: %s\n\t\tpayload:\t%s", err, string(b))
	}
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
