package fdk

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response is the domain type for the response.
type Response struct {
	Body   json.Marshaler
	Code   int
	Errors []APIError
	Header http.Header
}

// StatusCode returns the response status code. When a Response.Code is not
// set and errors exist, then the highest code on the errors is returned.
func (r Response) StatusCode() int {
	code := r.Code
	if code == 0 && len(r.Errors) > 0 {
		for _, e := range r.Errors {
			if e.Code > code {
				code = e.Code
			}
		}
	}
	if code == 0 {
		code = http.StatusOK
	}
	return code
}

// APIError defines a error that is shared back to the caller.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error provides a human readable error message.
func (a APIError) Error() string {
	return fmt.Sprintf("[%d] %s", a.Code, a.Message)
}

// JSON jsonifies the input to valid json upon request marshaling.
func JSON(v any) json.Marshaler {
	return jsoned{v: v}
}

type jsoned struct {
	v any
}

func (j jsoned) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.v)
}

// ErrResp creates a sad path errors only response.
//
// Note: the highest status code from the errors will be used for the response
// status if no status code is set on the response.
func ErrResp(errs ...APIError) Response {
	resp := Response{Errors: errs}
	resp.Code = resp.StatusCode()
	return resp
}
