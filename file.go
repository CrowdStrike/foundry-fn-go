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

// FileGZip writes the
func FileGZip(filename, contentType string, contents io.ReadCloser) File {
	return File{
		ContentType: contentType,
		Encoding:    "gzip",
		Filename:    filename,
		Contents:    newCompressorGZIP(contents),
	}
}

type compressorGZIP struct {
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

func newCompressorGZIP(rc io.ReadCloser) *compressorGZIP {
	pr, pw := io.Pipe()
	return &compressorGZIP{
		rc: rc,
		pw: pw,
		pr: pr,
		gw: gzip.NewWriter(pw),
	}
}

func (c *compressorGZIP) Read(p []byte) (int, error) {
	if c.started.CompareAndSwap(0, 1) {
		go c.compressInput()
	}
	return c.pr.Read(p)
}

func (c *compressorGZIP) compressInput() {
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

func (c *compressorGZIP) Close() error {
	var closers []io.Closer
	if !c.gwClosed {
		closers = append(closers, c.gw)
	}
	if !c.pwClosed {
		closers = append(closers, c.pw)
	}
	closers = append(closers, c.rc, c.pr)

	c.mu.Lock()
	errs := []error{c.closeErr, c.copyErr}
	c.mu.Unlock()
	for _, cl := range closers {
		errs = append(errs, cl.Close())
	}

	return errors.Join(errs...)
}
