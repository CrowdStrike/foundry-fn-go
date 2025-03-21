package fdk

import (
	"errors"
	"fmt"
	"io"
)

// ComplexPayload holds a mix of file streams and general inputs to a function.
type ComplexPayload struct {
	// Body holds the raw version of any non-file input.
	Body []byte
	// Files maps the file name to the file stream.
	Files map[string]io.Reader
}

var _ io.ReadCloser = (*ComplexPayload)(nil)

// Read is unsupported. ComplexPayload should be treated as a struct.
// This method is only added as a means to satisfy the Request type.
func (c *ComplexPayload) Read([]byte) (int, error) {
	return 0, errors.New("method Read() not supported - treat object as a standard struct")
}

// Close invokes Close() on all of the internal io.Closers.
func (c *ComplexPayload) Close() error {
	if len(c.Files) == 0 {
		return nil
	}

	var errs []error

	for k, v := range c.Files {
		if v == nil {
			continue
		}
		vc, ok := v.(io.Closer)
		if !ok {
			continue
		}
		if err := vc.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close %s: %w", k, err))
		}
	}

	return errors.Join(errs...)
}
