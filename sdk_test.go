package fdk_test

import (
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
