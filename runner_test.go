package fdk_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func TestRun_httprunner(t *testing.T) {
	t.Run("when executing provided handler with successful startup", func(t *testing.T) {
		type (
			inputs struct {
				body        []byte
				config      string
				headers     http.Header
				method      string
				path        string
				queryParams url.Values
			}

			wantFn func(t *testing.T, resp *http.Response)
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
				want: func(t *testing.T, resp *http.Response) {
					equalVals(t, 201, resp.StatusCode)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, resp.Header)

					var got respBody
					decodeBody(t, resp.Body, &got)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)

					if len(echo.Req.Body) > 0 {
						t.Errorf("invalid request body received\n\t\tgot: %s", string(echo.Req.Body))
					}
					if len(echo.Req.Context) > 0 {
						t.Errorf("invalid request context received\n\t\tgot: %s", string(echo.Req.Context))
					}
					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "DELETE", echo.Req.Method)
					containsHeaders(t, wantHeaders, echo.Req.Headers)
					equalVals(t, "id1", echo.Req.Queries.Get("ids"))
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
				want: func(t *testing.T, resp *http.Response) {
					equalVals(t, 201, resp.StatusCode)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, resp.Header)

					var got respBody
					decodeBody(t, resp.Body, &got)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)

					equalVals(t, "GET", echo.Req.Method)
					equalVals(t, "/path", echo.Req.Path)
					containsHeaders(t, wantHeaders, echo.Req.Headers)
					equalVals(t, "baz", echo.Req.Queries.Get("bar"))
				},
			},
			{
				name: "simple POST request should pass",
				inputs: inputs{
					body:   []byte(`{"dodgers":"stink","context":{"kings":"stink_too"}}`),
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
					m.Post("/path", newJSONBodyHandler(cfg))
					return m
				},
				want: func(t *testing.T, resp *http.Response) {
					equalVals(t, 201, resp.StatusCode)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, resp.Header)

					var got respBody
					decodeBody(t, resp.Body, &got)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					equalVals(t, `{"dodgers":"stink","context":{"kings":"stink_too"}}`, string(echo.Req.Body))
					equalVals(t, `{"kings":"stink_too"}`, string(echo.Req.Context))
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)
					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "POST", echo.Req.Method)
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
				want: func(t *testing.T, resp *http.Response) {
					equalVals(t, 201, resp.StatusCode)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, resp.Header)

					var got respBody
					decodeBody(t, resp.Body, &got)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					equalVals(t, `{"dodgers":"still stink"}`, string(echo.Req.Body))
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)
					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "PUT", echo.Req.Method)
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
				want: func(t *testing.T, resp *http.Response) {
					equalVals(t, 201, resp.StatusCode)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, resp.Header)

					var got respBody
					decodeBody(t, resp.Body, &got)

					if len(got.Errs) > 0 {
						t.Errorf("received unexpected errors\n\t\tgot:\t%+v", got.Errs)
					}

					echo := got.Req

					equalVals(t, `{"dodgers":"stink"}`, string(echo.Req.Body))
					equalVals(t, config{Str: "val", Int: 1}, echo.Config)
					equalVals(t, "/path", echo.Req.Path)
					equalVals(t, "POST", echo.Req.Method)
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
				want: func(t *testing.T, resp *http.Response) {
					equalVals(t, 404, resp.StatusCode)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, resp.Header)

					var got respBody
					decodeBody(t, resp.Body, &got)

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
				want: func(t *testing.T, resp *http.Response) {
					equalVals(t, 405, resp.StatusCode)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, resp.Header)

					var got respBody
					decodeBody(t, resp.Body, &got)

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
				want: func(t *testing.T, resp *http.Response) {
					equalVals(t, 400, resp.StatusCode)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, resp.Header)

					var got struct {
						Errs []fdk.APIError `json:"errors"`
					}
					decodeBody(t, resp.Body, &got)

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
				want: func(t *testing.T, resp *http.Response) {
					equalVals(t, http.StatusInternalServerError, resp.StatusCode)

					wantHeaders := make(http.Header)
					wantHeaders.Set("X-Cs-Origin", "fooorigin")
					wantHeaders.Set("X-Cs-Executionid", "exec_id")
					wantHeaders.Set("X-Cs-Traceid", "trace_id")
					containsHeaders(t, wantHeaders, resp.Header)

					var got struct {
						Errs []fdk.APIError `json:"errors"`
					}
					decodeBody(t, resp.Body, &got)

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
		}

		for _, tt := range tests {
			tt := tt
			fn := func(t *testing.T) {
				if tt.inputs.config != "" {
					tmp := filepath.Join(t.TempDir(), "config.json")
					t.Setenv("CS_FN_CONFIG_PATH", tmp)

					err := os.WriteFile(tmp, []byte(tt.inputs.config), 0666)
					mustNoErr(t, err)
				}

				readyChan := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				done := make(chan struct{})
				go func() {
					defer close(done)
					fdk.Run(ctx, func(cfg config) fdk.Handler {
						h := tt.newHandlerFn(ctx, cfg)
						close(readyChan)
						return h
					})
				}()

				select {
				case <-readyChan:
				case <-time.After(50 * time.Millisecond):
				}

				var reqBody io.Reader
				if len(tt.inputs.body) > 0 {
					reqBody = bytes.NewBuffer(tt.inputs.body)
				}

				req, err := http.NewRequestWithContext(
					ctx,
					tt.inputs.method,
					"http://localhost:8081"+path.Join("/", tt.inputs.path),
					reqBody,
				)
				mustNoErr(t, err)

				for h, vals := range tt.inputs.headers {
					for _, v := range vals {
						req.Header.Add(h, v)
					}
				}

				q := req.URL.Query()
				for name, vals := range tt.inputs.queryParams {
					for _, v := range vals {
						q.Add(name, v)
					}
				}
				req.URL.RawQuery = q.Encode()

				resp, err := http.DefaultClient.Do(req)
				mustNoErr(t, err)
				cancel()

				tt.want(t, resp)
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
		Errs []fdk.APIError `json:"errors"`
		Req  echoReq        `json:"body"`
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
