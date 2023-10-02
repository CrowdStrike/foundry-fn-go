package fdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	envPort = "PORT"

	mb = 1 << 20
)

const (
	headerTraceID = "X-Cs-Traceid"
	headerOrigin  = "X-Cs-Origin"
	headerExecID  = "X-Cs-Executionid"
)

type runnerHTTP struct{}

func (r *runnerHTTP) Run(ctx context.Context, logger *slog.Logger, h Handler) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r, err := toRequest(req)
		if err != nil {
			logger.Error("failed to create request", "err", err)
			writeErr := writeResponse(logger, req, w, Response{
				Errors: []APIError{{Code: http.StatusInternalServerError, Message: "unable to process incoming request"}},
			})
			if writeErr != nil {
				logger.Error("failed to write failed request response", "err", writeErr)
			}
			return
		}

		resp := h.Handle(ctx, r)

		err = writeResponse(logger, req, w, resp)
		if err != nil {
			logger.Error("failed to write response", "err", err)
		}
	}))

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port()),
		Handler:        mux,
		MaxHeaderBytes: mb,
	}
	go func() {
		select {
		case <-ctx.Done():
			if err := s.Shutdown(ctx); err != nil {
				logger.Error("failed to shutdown server", "err", err)
			}
		}
	}()

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("unexpected shutdown of server", "err", err)
	}
}

func toRequest(req *http.Request) (Request, error) {
	r := Request{
		Params: struct {
			Header http.Header
			Query  url.Values
		}{
			Header: req.Header,
			Query:  req.URL.Query(),
		},
		URL:    req.URL.Path,
		Method: req.Method,
	}
	if req.Body != nil {
		body, err := io.ReadAll(io.LimitReader(req.Body, 5*mb))
		if err != nil {
			return Request{}, err
		}
		if len(body) > 0 {
			r.Body = body

			var ctxPayload struct {
				Context json.RawMessage `json:"context"`
			}
			_ = json.Unmarshal(body, &ctxPayload)
			r.Context = ctxPayload.Context
		}
	}

	return r, nil
}

func addReqHeaders(reqH, respH http.Header) http.Header {
	if respH == nil {
		respH = make(http.Header, 3)
	}
	takeHeaders(reqH, respH, headerExecID, headerOrigin, headerTraceID)
	return respH
}

func takeHeaders(reqH, respH http.Header, headers ...string) {
	for _, header := range headers {
		if v := reqH.Get(header); v != "" {
			respH.Set(header, v)
		}
	}
}

func writeResponse(logger *slog.Logger, req *http.Request, w http.ResponseWriter, resp Response) error {
	for name, vals := range addReqHeaders(req.Header, resp.Header) {
		for _, v := range vals {
			w.Header().Set(name, v)
		}
	}

	b, err := json.Marshal(struct {
		Errors []APIError     `json:"errors"`
		Body   json.Marshaler `json:"body"`
	}{
		Errors: resp.Errors,
		Body:   resp.Body,
	})
	if err != nil {
		logger.Error("failed to marshal json payload with body", "err", err)
		b, err = json.Marshal(struct {
			Errors []APIError `json:"errors"`
		}{
			Errors: append(resp.Errors, APIError{Code: http.StatusInternalServerError, Message: err.Error()}),
		})
		if err != nil {
			logger.Error("failed to marshal json error payload", "err", err)
		}
	}

	if resp.StatusCode() > 0 {
		w.WriteHeader(resp.StatusCode())
	}

	_, err = w.Write(b)
	return err
}

func port() int {
	if v, _ := strconv.Atoi(os.Getenv(envPort)); v > 0 {
		return v
	}
	return 8081
}
