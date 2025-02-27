package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func main() {
	fdk.Run(context.Background(), newHandler)
}

// newHandler showcases how to provide a single input to a function which happens to be a file.
// In this case, it simply echos back the contents of a text file, though this could easily
// be changed to do something more advanced or to work with arbitrary binary data.
func newHandler(_ context.Context, log *slog.Logger, _ fdk.SkipCfg) fdk.Handler {
	mux := fdk.NewMux()
	mux.Post("/my-endpoint", fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.With("error", err).Error("failed to read payload")
			return fdk.ErrResp(fdk.APIError{
				Code:    500,
				Message: fmt.Sprintf("failed to read payload with error: %s", err),
			})
		}

		return fdk.Response{
			Body:   fdk.JSON(map[string]string{"fileContents": string(b)}),
			Code:   200,
			Header: http.Header{"X-Fn-Method": []string{r.Method}},
		}
	}))
	return mux
}
