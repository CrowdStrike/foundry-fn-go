package fdk_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func TestAPIError(t *testing.T) {
	errs := []fdk.APIError{
		{Code: http.StatusInternalServerError, Message: "some internal error"},
		{Code: http.StatusBadRequest, Message: "user dorked it up"},
		{Message: "missing code will print a zero"},
	}
	for _, err := range errs {
		want := fmt.Sprintf("[%d] %s", err.Code, err.Message)
		fdk.EqualVals(t, want, err.Error())
	}
}

func TestHandlerFnOfOK(t *testing.T) {
	h := fdk.HandlerFnOfOK(func(ctx context.Context, r fdk.RequestOf[okReq]) fdk.Response {
		return fdk.Response{Code: http.StatusOK}
	})

	b, err := json.Marshal(okReq{
		Shoulds: []int{
			http.StatusOK,                  // should not cause an error
			http.StatusBadRequest,          // should return a 400 error
			http.StatusInternalServerError, // should return a 500 error
		},
	})
	mustNoErr(t, err)

	resp := h.Handle(context.TODO(), fdk.Request{Body: b})
	fdk.EqualVals(t, http.StatusInternalServerError, resp.Code)

	wantErrs := []fdk.APIError{
		{Code: http.StatusBadRequest, Message: http.StatusText(http.StatusBadRequest)},
		{Code: http.StatusInternalServerError, Message: http.StatusText(http.StatusInternalServerError)},
	}
	if !fdk.EqualVals(t, len(wantErrs), len(resp.Errors)) {
		return
	}
	fdk.EqualVals(t, wantErrs[0], resp.Errors[0])
	fdk.EqualVals(t, wantErrs[1], resp.Errors[1])
}

type okReq struct {
	Shoulds []int `json:"shoulds"`
}

func (r okReq) OK() []fdk.APIError {
	var out []fdk.APIError
	for _, sh := range r.Shoulds {
		if sh < 400 {
			continue
		}
		out = append(out, fdk.APIError{Code: sh, Message: http.StatusText(sh)})
	}
	return out
}
