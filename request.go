package fdk

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type (
	// Request defines a request structure that is given to the runner. The Body is set to
	// json.RawMessage, to enable decoration/middleware.
	Request RequestOf[json.RawMessage]

	// RequestOf provides a generic body we can target our unmarshaling into.
	RequestOf[T any] struct {
		FnID      string
		FnVersion int

		Body T

		Context json.RawMessage
		Params  struct {
			Header http.Header
			Query  url.Values
		}
		URL         string
		Method      string
		AccessToken string
		TraceID     string
	}
)
