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
	"time"
)

const (
	envPort = "PORT"

	mb = 1 << 20
)

type runnerHTTP struct{}

func (r *runnerHTTP) Run(ctx context.Context, logger *slog.Logger, h Handler) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r, err := toRequest(req)
		if err != nil {
			logger.Error("failed to create request", "err", err)
			writeErr := writeResponse(logger, w, Response{
				Errors: []APIError{{Code: http.StatusInternalServerError, Message: "unable to process incoming request"}},
			})
			if writeErr != nil {
				logger.Error("failed to write failed request response", "err", writeErr)
			}
			return
		}

		resp := h.Handle(ctx, r)

		err = writeResponse(logger, w, resp)
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
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			logger.Info("shutting down HTTP server...")
			if err := s.Shutdown(shutdownCtx); err != nil {
				logger.Error("failed to shutdown server", "err", err)
			}
		}
	}()

	logger.Info("serving HTTP server on port " + strconv.Itoa(port()))
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("unexpected shutdown of server", "err", err)
	}
}

func toRequest(req *http.Request) (Request, error) {
	var r struct {
		Body        json.RawMessage `json:"body"`
		Context     json.RawMessage `json:"context"`
		AccessToken string          `json:"access_token"`
		Method      string          `json:"method"`
		Params      struct {
			Header http.Header `json:"header"`
			Query  url.Values  `json:"query"`
		} `json:"params"`
		URL string `json:"url"`
	}
	payload, err := io.ReadAll(io.LimitReader(req.Body, 5*mb))
	if err != nil {
		return Request{}, fmt.Errorf("failed to read request body: %s", err)
	}

	if err = json.Unmarshal(payload, &r); err != nil {
		return Request{}, fmt.Errorf("failed to unmarshal request body: %s", err)
	}

	// Ensure headers are canonically formatted else header.Get("my-key") won't necessarily work.
	hCanon := make(http.Header)
	for k, v := range r.Params.Header {
		for _, s := range v {
			hCanon.Add(k, s)
		}
	}
	r.Params.Header = hCanon

	out := Request{
		Body:    r.Body,
		Context: r.Context,
		Params: struct {
			Header http.Header
			Query  url.Values
		}{Header: r.Params.Header, Query: r.Params.Query},
		URL:         r.URL,
		Method:      r.Method,
		AccessToken: r.AccessToken,
	}
	return out, nil
}

func writeResponse(logger *slog.Logger, w http.ResponseWriter, resp Response) error {
	b, err := json.Marshal(struct {
		Body    json.Marshaler `json:"body,omitempty"`
		Code    int            `json:"code,omitempty"`
		Errors  []APIError     `json:"errors"`
		Headers http.Header    `json:"headers,omitempty"`
	}{
		Body:    resp.Body,
		Code:    resp.StatusCode(),
		Errors:  resp.Errors,
		Headers: resp.Header,
	})
	if err != nil {
		logger.Error("failed to marshal json payload with body", "err", err)
	}
	if len(b) == 0 {
		return nil
	}

	if code := resp.StatusCode(); code != 0 {
		w.WriteHeader(code)
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
