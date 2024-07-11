package fdk_test

import (
	"fmt"
	"net/http"
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func TestFn(t *testing.T) {
	type (
		inputs struct {
			fnID string
		}

		wantFn func(t *testing.T, gotFnID string)
	)

	tests := []struct {
		name   string
		inputs inputs
		wants  wantFn
	}{
		{
			name: "fn-id set with version 1",
			inputs: inputs{
				fnID: "fn-id",
			},
			wants: func(t *testing.T, gotFnID string) {
				fdk.EqualVals(t, "fn-id", gotFnID)
			},
		},
		{
			name: "fn-id set without version",
			inputs: inputs{
				fnID: "fn-id",
			},
			wants: func(t *testing.T, gotFnID string) {
				fdk.EqualVals(t, "fn-id", gotFnID)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CS_FN_ID", tt.inputs.fnID)

			fn := fdk.Fn()
			tt.wants(t, fn.ID)
		})
	}

}

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
