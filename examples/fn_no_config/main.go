package main

import (
	"context"
	"encoding/json"
	"net/http"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func main() {
	fdk.Run(context.Background(), newHandler)
}

// newHandlerWithCfg here is showcasing a handler that does not utilize a config, so
// it provides the SkipCfg as the config so no config load is attempted. This is the
// minority of functions.
func newHandler(context.Context, fdk.SkipCfg) fdk.Handler {
	mux := fdk.NewMux()
	mux.Get("/", fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		return fdk.Response{
			Body:   json.RawMessage(`{"foo":"val"}`),
			Code:   200,
			Header: http.Header{"X-Fn-Method": []string{r.Method}},
		}
	}))
	return mux
}
