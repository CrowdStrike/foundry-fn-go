package fdk

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// Runner defines the runtime that executes the request/response handler lifecycle.
type Runner func(ctx context.Context, newHandlerFn func(context.Context, *slog.Logger) Handler)

// RegisterRunner registers a runner.
func RegisterRunner(runnerType string, r Runner) {
	if _, ok := runners[runnerType]; ok {
		panic(fmt.Sprintf("runner type already exists: %q", runnerType))
	}

	runners[runnerType] = r
}

func run(ctx context.Context, newHandlerFn func(context.Context, *slog.Logger) Handler) {
	rt := os.Getenv("CS_RUNNER_TYPE")
	if rt == "" {
		rt = "http"
	}

	r := runners[rt]
	if r == nil {
		panic(fmt.Sprintf("invalid RUNNER_TYPE provided: %q", rt))
	}

	r(ctx, newHandlerFn)
}

var runners = map[string]func(ctx context.Context, newHandlerFn func(context.Context, *slog.Logger) Handler){
	"http": runHTTP,
}
