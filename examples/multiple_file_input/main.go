package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func main() {
	fdk.Run(context.Background(), newHandler)
}

// newHandler showcases how to provide multiple inputs to a function which happen to be files.
// In this case, it simply echoes back the contents of those files concatenated together with spaces.
// This could easily be changed to something more advanced to work with arbitrary binary.
func newHandler(_ context.Context, log *slog.Logger, _ fdk.SkipCfg) fdk.Handler {
	mux := fdk.NewMux()
	mux.Post("/my-endpoint", fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		s, err := readFiles(r.Body)
		if err != nil {
			log.With("error", err).Error("failed to read payload")
			return fdk.ErrResp(fdk.APIError{
				Code:    500,
				Message: fmt.Sprintf("failed to read payload with error: %s", err),
			})
		}
		return fdk.Response{
			Body:   fdk.JSON(map[string]string{"allText": s}),
			Code:   200,
			Header: http.Header{"X-Fn-Method": []string{r.Method}},
		}
	}))
	return mux
}

func readFiles(body io.Reader) (string, error) {
	if c, ok := body.(*fdk.ComplexPayload); ok {
		// Multiple files are presented in a map of string -> io.Reader.
		// This mapping consists of the filename and a link to the uploaded file.
		contents := make([]string, 0)
		for _, r := range c.Files {
			s, err := readString(r)
			if err != nil {
				return "", err
			}
			contents = append(contents, s)
		}
		return strings.Join(contents, " "), nil
	}

	return readString(body)
}

func readString(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
