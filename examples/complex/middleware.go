package main

import (
	"context"
	"log/slog"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func loggingMW(logger *slog.Logger) func(next fdk.Handler) fdk.Handler {
	return func(next fdk.Handler) fdk.Handler {
		return fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
			logger.Info("import logging here",
				"url_path", r.URL,
				/* trim additional fields  */
			)

			return next.Handle(ctx, r)
		})
	}
}
