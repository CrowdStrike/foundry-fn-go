package fdk_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

type testBody struct {
	Name string `json:"name"`
}

func (t testBody) OK() []fdk.APIError {
	if t.Name == "fail" {
		return []fdk.APIError{{Code: http.StatusBadRequest, Message: "got a fail"}}
	}
	return nil
}

func TestHandlers(t *testing.T) {
	mux := fdk.NewMux()
	mux.Post("/handler-fn-of", fdk.HandleFnOf(func(ctx context.Context, r fdk.RequestOf[testBody]) fdk.Response {
		return fdk.Response{Code: 200, Body: fdk.JSON(r)}
	}))
	mux.Post("/handle-fn-of-ok", fdk.HandlerFnOfOK(func(ctx context.Context, r fdk.RequestOf[testBody]) fdk.Response {
		return fdk.Response{Code: 200, Body: fdk.JSON(r)}
	}))
	mux.Get("/handle-workflow", fdk.HandleWorkflow(func(ctx context.Context, r fdk.Request, wrkCtx fdk.WorkflowCtx) fdk.Response {
		return fdk.Response{Code: 200, Body: fdk.JSON(r)}
	}))
	mux.Put("/handle-workflow-of", fdk.HandleWorkflowOf(func(ctx context.Context, r fdk.RequestOf[testBody], wrkCtx fdk.WorkflowCtx) fdk.Response {
		return fdk.Response{Code: 200, Body: fdk.JSON(r)}
	}))

	params := struct {
		Header http.Header
		Query  url.Values
	}{
		Header: http.Header{"X-Cs-Foo": []string{"header"}},
		Query:  url.Values{"key": []string{"value"}},
	}

	wrkCtxVal, err := json.Marshal(fdk.WorkflowCtx{
		ActivityExecID:    "act-exec-id",
		AppID:             "app-id",
		CID:               "cid",
		OwnerCID:          "owner-cid",
		DefinitionID:      "def-id",
		DefinitionVersion: 9000,
		ExecutionID:       "exec-id",
	})
	mustNoErr(t, err)

	t.Run("HandleFnOf", func(t *testing.T) {
		resp := mux.Handle(context.TODO(), fdk.Request{
			FnID:        "id1",
			FnVersion:   1,
			Body:        strings.NewReader(`{"name":"frodo"}`),
			Context:     json.RawMessage(`{"some":"ctx"}`),
			URL:         "/handler-fn-of",
			Params:      params,
			Method:      "POST",
			AccessToken: "access",
			TraceID:     "trace-id",
		})
		gotStatusOK(t, resp)

		b, err := resp.Body.MarshalJSON()
		mustNoErr(t, err)

		var got fdk.RequestOf[testBody]
		mustNoErr(t, json.Unmarshal(b, &got))

		fdk.EqualVals(t, "id1", got.FnID)
		fdk.EqualVals(t, 1, got.FnVersion)
		fdk.EqualVals(t, "/handler-fn-of", got.URL)
		fdk.EqualVals(t, "POST", got.Method)
		fdk.EqualVals(t, `{"some":"ctx"}`, string(got.Context))
		fdk.EqualVals(t, "header", got.Params.Header.Get("X-Cs-Foo"))
		fdk.EqualVals(t, "value", got.Params.Query.Get("key"))
		fdk.EqualVals(t, "access", got.AccessToken)
		fdk.EqualVals(t, "trace-id", got.TraceID)

		wantFoo := testBody{Name: "frodo"}
		fdk.EqualVals(t, wantFoo, got.Body)
	})

	t.Run("HandlerFnOfOK", func(t *testing.T) {
		t.Run("with valid name", func(t *testing.T) {
			resp := mux.Handle(context.TODO(), fdk.Request{
				FnID:        "id1",
				FnVersion:   1,
				Body:        strings.NewReader(`{"name":"frodo"}`),
				Context:     json.RawMessage(`{"some":"ctx"}`),
				URL:         "/handle-fn-of-ok",
				Params:      params,
				Method:      "POST",
				AccessToken: "access",
				TraceID:     "trace-id",
			})
			gotStatusOK(t, resp)

			b, err := resp.Body.MarshalJSON()
			mustNoErr(t, err)

			var got fdk.RequestOf[testBody]
			mustNoErr(t, json.Unmarshal(b, &got))

			fdk.EqualVals(t, "id1", got.FnID)
			fdk.EqualVals(t, 1, got.FnVersion)
			fdk.EqualVals(t, "/handle-fn-of-ok", got.URL)
			fdk.EqualVals(t, "POST", got.Method)
			fdk.EqualVals(t, `{"some":"ctx"}`, string(got.Context))
			fdk.EqualVals(t, "header", got.Params.Header.Get("X-Cs-Foo"))
			fdk.EqualVals(t, "value", got.Params.Query.Get("key"))
			fdk.EqualVals(t, "access", got.AccessToken)
			fdk.EqualVals(t, "trace-id", got.TraceID)

			wantFoo := testBody{Name: "frodo"}
			fdk.EqualVals(t, wantFoo, got.Body)
		})

		t.Run("with invalid name", func(t *testing.T) {
			resp := mux.Handle(context.TODO(), fdk.Request{
				FnID:      "id1",
				FnVersion: 1,
				Body:      strings.NewReader(`{"name":"fail"}`),
				URL:       "/handle-fn-of-ok",
				Method:    "POST",
			})
			fdk.EqualVals(t, http.StatusBadRequest, resp.Code)
			fdk.EqualVals(t, 1, len(resp.Errors), "got invalid errors: %s", resp.Errors)
			fdk.EqualVals(t, fdk.APIError{Code: http.StatusBadRequest, Message: "got a fail"}, resp.Errors[0])
		})
	})

	t.Run("HandleWorkflow", func(t *testing.T) {
		resp := mux.Handle(context.TODO(), fdk.Request{
			FnID:        "id1",
			FnVersion:   1,
			Context:     wrkCtxVal,
			URL:         "/handle-workflow",
			Params:      params,
			Method:      "GET",
			AccessToken: "access",
			TraceID:     "trace-id",
		})
		gotStatusOK(t, resp)

		b, err := resp.Body.MarshalJSON()
		mustNoErr(t, err)

		var got fdk.Request
		mustNoErr(t, json.Unmarshal(b, &got))

		fdk.EqualVals(t, "id1", got.FnID)
		fdk.EqualVals(t, 1, got.FnVersion)
		fdk.EqualVals(t, "/handle-workflow", got.URL)
		fdk.EqualVals(t, "GET", got.Method)
		fdk.EqualVals(t, string(wrkCtxVal), string(got.Context))
		fdk.EqualVals(t, "header", got.Params.Header.Get("X-Cs-Foo"))
		fdk.EqualVals(t, "value", got.Params.Query.Get("key"))
		fdk.EqualVals(t, "access", got.AccessToken)
		fdk.EqualVals(t, "trace-id", got.TraceID)
	})

	t.Run("HandleWorkflowOf", func(t *testing.T) {
		resp := mux.Handle(context.TODO(), fdk.Request{
			FnID:        "id1",
			FnVersion:   1,
			Body:        strings.NewReader(`{"name":"frodo"}`),
			Context:     wrkCtxVal,
			URL:         "/handle-workflow-of",
			Params:      params,
			Method:      "PUT",
			AccessToken: "access",
			TraceID:     "trace-id",
		})
		gotStatusOK(t, resp)

		b, err := resp.Body.MarshalJSON()
		mustNoErr(t, err)

		var got fdk.RequestOf[testBody]
		mustNoErr(t, json.Unmarshal(b, &got))

		fdk.EqualVals(t, "id1", got.FnID)
		fdk.EqualVals(t, 1, got.FnVersion)
		fdk.EqualVals(t, "/handle-workflow-of", got.URL)
		fdk.EqualVals(t, "PUT", got.Method)
		fdk.EqualVals(t, string(wrkCtxVal), string(got.Context))
		fdk.EqualVals(t, "header", got.Params.Header.Get("X-Cs-Foo"))
		fdk.EqualVals(t, "value", got.Params.Query.Get("key"))
		fdk.EqualVals(t, "access", got.AccessToken)
		fdk.EqualVals(t, "trace-id", got.TraceID)

		wantFoo := testBody{Name: "frodo"}
		fdk.EqualVals(t, wantFoo, got.Body)
	})
}

func gotStatusOK(t *testing.T, resp fdk.Response) {
	t.Helper()

	b, _ := json.MarshalIndent(resp.Errors, "", "  ")
	equals := fdk.EqualVals(t, 0, len(resp.Errors), "errors encountered:\n"+string(b))
	equals = fdk.EqualVals(t, http.StatusOK, resp.Code, "status code is invalid") && equals

	if !equals {
		t.FailNow()
	}
}
