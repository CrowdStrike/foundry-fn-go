package fdktest

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
	"testing"
)

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
