package fdktest

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

// RequestOf converts a RequestOf into its fdk.Request equivalent. I.e.
// creates the json payload.
func RequestOf[T any](t *testing.T, r fdk.RequestOf[T]) fdk.Request {
	t.Helper()

	b, err := json.Marshal(r.Body)
	mustNoErr(t, "", err)

	return fdk.Request{
		FnID:        r.FnID,
		FnVersion:   r.FnVersion,
		Body:        bytes.NewReader(b),
		Context:     r.Context,
		Headers:     r.Headers,
		Queries:     r.Queries,
		URL:         r.URL,
		Method:      r.Method,
		AccessToken: r.AccessToken,
		TraceID:     r.TraceID,
	}
}

// GZIPReader is a helper for gzipping the contents and returning the reader.
func GZIPReader(t *testing.T, v string) io.Reader {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	defer func() { _ = gw.Close() }()

	_, err := gw.Write([]byte(v))
	if err != nil {
		t.Fatal("failed to write to gzip writer: " + err.Error())
	}

	err = gw.Close()
	if err != nil {
		t.Fatal("failed to close gzip writer: " + err.Error())
	}

	return &buf
}

// NewLogger creates a new logger that integrates with the testing.T logging.
func NewLogger(t *testing.T) *slog.Logger {
	return slog.New(slog.NewJSONHandler(&testLogger{t: t}, nil))
}

type testLogger struct {
	t *testing.T
}

func (t *testLogger) Write(p []byte) (int, error) {
	t.t.Log(string(p))
	return len(p), nil
}
