package fdk

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var initMime = func() func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			_ = mime.AddExtensionType(".xml", "application/xml")
			_ = mime.AddExtensionType(".jsonld", "application/ld+json")
			_ = mime.AddExtensionType(".jsonnet", "application/jsonnet")
			_ = mime.AddExtensionType(".yaml", "text/yaml")
			_ = mime.AddExtensionType(".yml", "text/yaml")
		})
	}
}()

const contentTypeOctetStream = "application/octet-stream"

var nowFn = time.Now

// File represents a response that is a response body. The runner is in charge
// of getting the contents to the destination. The metadata will be received. One
// note, we call NormalizeFile on the File's in the Runner's that execute the handler.
// Testing through the Run function will illustrate this.
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

// NormalizeFile normalizes a file so that all fields are set with sane defaults.
func NormalizeFile(f File) File {
	if f.ContentType == "" {
		f.ContentType = normalizeContentType(f.Filename)
	}
	if f.Encoding == "" {
		f.Encoding = normalizeEncoding(f.Filename)
	}
	if f.Filename == "" {
		f.Filename = normalizeFilename(f.ContentType, f.Encoding, nowFn())
	}
	return f
}

// CompressGzip compresses a files contents with gzip compression.
func CompressGzip(file File) File {
	switch {
	case file.Encoding == "":
		file.Encoding = "gzip"
	case file.Encoding != "" && !strings.Contains(file.Encoding, "gzip"):
		file.Encoding += ", gzip"
	}
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

func normalizeContentType(filename string) string {
	initMime()
	parts := strings.Split(filepath.Base(filename), ".")
	if len(parts) < 2 {
		return contentTypeOctetStream
	}
	for _, ext := range parts[1:] {
		if ct := mime.TypeByExtension("." + ext); ct != "" {
			return ct
		}
	}
	return contentTypeOctetStream
}

func normalizeEncoding(filename string) string {
	parts := strings.SplitN(filepath.Base(filename), ".", 2)
	if len(parts) == 1 {
		return ""
	}

	mapping := map[string]string{
		"br":  "brotli",
		"gz":  "gzip",
		"zst": "zstd",
	}
	var out []string
	for _, part := range strings.Split(parts[1], ".") {
		if encoding, ok := mapping[part]; ok {
			out = append(out, encoding)
		}
	}
	return strings.Join(out, ", ")
}

var compressionToExt = map[string]string{
	"brotli": "br",
	"gzip":   "gz",
	"zstd":   "zst",
}

func normalizeFilename(contentType, encoding string, now time.Time) string {
	filename := "upload_" + now.Format(time.RFC3339)
	if encoding != "" {
		var converted []string
		for _, enc := range strings.Split(strings.ReplaceAll(encoding, " ", ""), ",") {
			if ext := compressionToExt[enc]; ext != "" {
				converted = append(converted, ext)
			}
		}
		if len(converted) > 0 {
			encoding = "." + strings.Join(converted, ".")
		}
	}

	var ctExt string
	if contentType != contentTypeOctetStream {
		initMime()
		exts, _ := mime.ExtensionsByType(contentType)
		if len(exts) > 0 {
			ctExt = exts[0]
		}
	}
	return filename + ctExt + encoding
}
