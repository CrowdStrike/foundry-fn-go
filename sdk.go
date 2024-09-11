package fdk

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Handler provides a handler for our incoming request.
type Handler interface {
	Handle(ctx context.Context, r Request) Response
}

// Run is the meat and potatoes. This is the entrypoint for everything.
func Run[T Cfg](ctx context.Context, newHandlerFn func(context.Context, *slog.Logger, T) Handler) {
	run(ctx, func(ctx context.Context, logger *slog.Logger) Handler {
		var runFn Handler = HandlerFn(func(ctx context.Context, r Request) Response {
			cfg, loadErr := readCfg[T](ctx)
			if loadErr != nil {
				if loadErr.err != nil {
					logger.Error("failed to load config", "err", loadErr.err)
				}
				return ErrResp(loadErr.apiErr)
			}

			h := newHandlerFn(ctx, logger, cfg)

			return h.Handle(ctx, r)
		})
		runFn = recoverer(logger)(runFn)

		return runFn
	})
}

func recoverer(logger *slog.Logger) func(Handler) Handler {
	return func(h Handler) Handler {
		return HandlerFn(func(ctx context.Context, r Request) (resp Response) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic caught", "stack_trace", string(debug.Stack()))
					resp = ErrResp(APIError{Code: http.StatusServiceUnavailable, Message: "encountered unexpected error"})
				}
			}()

			return h.Handle(ctx, r)
		})
	}
}
