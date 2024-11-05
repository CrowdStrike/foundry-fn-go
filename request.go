package fdk

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type (
	// Request defines a request structure that is given to the runner. The Body is set to
	// io.Reader, to enable decoration/middleware.
	Request RequestOf[io.Reader]

	// RequestOf provides a generic body we can target our unmarshaling into.
	RequestOf[T any] struct {
		FnID      string
		FnVersion int

		Body T

		Context     json.RawMessage
		Headers     http.Header
		Queries     url.Values
		URL         string
		Method      string
		AccessToken string
		TraceID     string
	}
)
