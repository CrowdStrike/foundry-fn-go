package fdk

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

// File represents a response that is a response body. The runner is in charge
// of getting the contents to the destination. The metadata will be received.
type File struct {
	ContentType string        `json:"content_type"`
	Encoding    string        `json:"encoding"`
	Filename    string        `json:"filename"`
	Contents    io.ReadCloser `json:"-"`
}

// MarshalJSON marshals the file metadata.
func (f File) MarshalJSON() ([]byte, error) {
	type alias File
	return json.Marshal(alias(f))
}

// CompressGzip compresses a files contents with gzip compression.
func CompressGzip(file File) File {
	encoding := "gzip"
	if file.Encoding != "" {
		encoding = file.Encoding + ", " + encoding
	}
	file.Encoding = encoding
	file.Contents = newCompressorGzip(file.Contents)
	return file
}

type compressorGzip struct {
	rc io.ReadCloser
	pr *io.PipeReader

	pwClosed bool
	pw       *io.PipeWriter
	gwClosed bool
	gw       *gzip.Writer

	mu       sync.Mutex
	started  atomic.Int32
	closeErr error
	copyErr  error
}

func newCompressorGzip(rc io.ReadCloser) *compressorGzip {
	pr, pw := io.Pipe()
	return &compressorGzip{
		rc: rc,
		pw: pw,
		pr: pr,
		gw: gzip.NewWriter(pw),
	}
}

func (c *compressorGzip) Read(p []byte) (int, error) {
	if c.started.CompareAndSwap(0, 1) {
		go c.compressInput()
	}
	return c.pr.Read(p)
}

func (c *compressorGzip) compressInput() {
	defer func() {
		c.gwClosed, c.pwClosed = true, true
		err := c.pw.CloseWithError(c.gw.Close())
		c.mu.Lock()
		c.closeErr = err
		c.mu.Unlock()
	}()
	if _, err := io.Copy(c.gw, c.rc); err != nil && !errors.Is(err, io.EOF) {
		c.mu.Lock()
		c.copyErr = err
		c.mu.Unlock()
	}
}

func (c *compressorGzip) Close() error {
	c.mu.Lock()
	errs := []error{c.closeErr, c.copyErr}
	c.mu.Unlock()
	if !c.gwClosed {
		errs = append(errs, c.gw.Close())
	}
	if !c.pwClosed {
		errs = append(errs, c.pw.Close())
	}

	errs = append(errs, c.rc.Close(), c.pr.Close())

	return errors.Join(errs...)
}
