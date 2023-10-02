package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func main() {
	fdk.Run(context.Background(), newHandler)
}

// newHandler here is showing how a config is integrated. It is using generics,
// so we can unmarshal the config into a concrete type and then validate it. The
// OK method is run to validate the contents of the config.
func newHandler(_ context.Context, cfg config) fdk.Handler {
	mux := fdk.NewMux()
	mux.Get("/foo", fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		return fdk.Response{
			Body:   json.RawMessage(`{"foo":"val"}`),
			Code:   200,
			Header: http.Header{"X-Fn-Method": []string{r.Method}},
		}
	}))
	mux.Post("/foo", fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		return fdk.Response{
			Body:   r.Body,
			Code:   201,
			Header: http.Header{"X-Fn-Method": []string{r.Method}},
		}
	}))
	return mux
}

type config struct {
	Int int    `json:"integer"`
	Str string `json:"string"`
}

func (c config) OK() error {
	var errs []error
	if c.Int < 1 {
		errs = append(errs, errors.New("integer must be greater than 0"))
	}
	if c.Str == "" {
		errs = append(errs, errors.New("non empty string must be provided"))
	}
	return errors.Join(errs...)
}
