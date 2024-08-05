package fdk

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
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
			writeErr := writeResponse(logger, w, ErrResp(APIError{Code: http.StatusInternalServerError, Message: "unable to process incoming request"}))
			if writeErr != nil {
				logger.Error("failed to write failed request response", "err", writeErr)
			}
			return
		}

		const ctxKeyTraceID = "_traceid"
		ctx = context.WithValue(ctx, ctxKeyTraceID, r.TraceID)

		resp := h.Handle(ctx, r)

		if f, ok := resp.Body.(File); ok {
			f = NormalizeFile(f)
			sha256Hash, size, err := writeFile(logger, f.Contents, f.Filename)
			if err != nil {
				resp.Errors = append(resp.Errors, APIError{Code: http.StatusInternalServerError, Message: err.Error()})
				writeErr := writeResponse(logger, w, resp)
				if writeErr != nil {
					logger.Error("failed to write failed request response", "write_err", writeErr, "err", err.Error())
				}
				return
			}
			// the sha and size will be left to the runner to determine. This removes the chicken and egg
			// problem where you need hte size and sha but want to work with the stream only. This isn't
			// possible without having the runner do it. We can maintain streaming semantics while also
			// obtaining our sha/size by extending the writer to support this when we're moving the
			// contents to disk (or w/e sync we use)
			respBody := struct {
				ContentType string `json:"content_type"`
				Encoding    string `json:"encoding"`
				Filename    string `json:"filename"`
				SHA256      string `json:"sha256"`
				Size        int    `json:"size,string"`
			}{
				ContentType: f.ContentType,
				Encoding:    f.Encoding,
				Filename:    f.Filename,
				SHA256:      sha256Hash,
				Size:        size,
			}
			resp.Body = JSON(respBody)
		}

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
		FnID        string          `json:"fn_id"`
		FnVersion   int             `json:"fn_version"`
		Body        json.RawMessage `json:"body"`
		Context     json.RawMessage `json:"context"`
		AccessToken string          `json:"access_token"`
		Method      string          `json:"method"`
		Params      struct {
			Header http.Header `json:"header"`
			Query  url.Values  `json:"query"`
		} `json:"params"`
		URL     string `json:"url"`
		TraceID string `json:"trace_id"`
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
		FnID:      r.FnID,
		FnVersion: r.FnVersion,
		Body:      r.Body,
		Context:   r.Context,
		Params: struct {
			Header http.Header
			Query  url.Values
		}{Header: r.Params.Header, Query: r.Params.Query},
		Method:      r.Method,
		URL:         r.URL,
		TraceID:     r.TraceID,
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

func writeFile(logger *slog.Logger, r io.ReadCloser, filename string) (string, int, error) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		// just in case
		_ = f.Close()
		_ = r.Close()
	}()

	sizer, h := &sizeRecorder{w: f}, sha256.New()
	mw := io.MultiWriter(sizer, h)

	_, err = io.Copy(mw, r)
	if err != nil {
		return "", 0, fmt.Errorf("failed to write contents to file: %w", err)
	}

	err = r.Close()
	if err != nil {
		// we swallow the error here, there's nothing we can do about it...
		logger.Error("failed to close file contents", "err", err)
	}

	err = f.Close()
	if err != nil {
		return "", 0, fmt.Errorf("failed to close file: %w", err)
	}

	sha256Hash := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return sha256Hash, sizer.n, nil
}

func port() int {
	if v, _ := strconv.Atoi(os.Getenv(envPort)); v > 0 {
		return v
	}
	return 8081
}

type sizeRecorder struct {
	w io.Writer
	n int
}

func (s *sizeRecorder) Write(p []byte) (int, error) {
	n, err := s.w.Write(p)
	s.n += n
	return n, err
}
