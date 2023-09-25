package fdk

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// Runner defines the runtime that executes the request/response handler lifecycle.
type Runner interface {
	Run(ctx context.Context, logger *slog.Logger, h Handler)
}

// RegisterRunner registers a runner.
func RegisterRunner(runnerType string, r Runner) {
	if _, ok := runners[runnerType]; ok {
		panic(fmt.Sprintf("runner type already exists: %q", runnerType))
	}

	runners[runnerType] = r
}

func run(ctx context.Context, logger *slog.Logger, h Handler) {
	rt := os.Getenv("RUNNER_TYPE")
	if rt == "" {
		rt = "http"
	}

	r := runners[rt]
	if r == nil {
		panic(fmt.Sprintf("invalid RUNNER_TYPE provided: %q", rt))
	}

	r.Run(ctx, logger, h)
}

var runners = map[string]Runner{
	"http": new(runnerHTTP),
}
