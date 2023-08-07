package fdk

import "encoding/json"

// Response is a base response
type Response struct {
	Body    json.RawMessage     `json:"body,omitempty"`
	Code    int                 `json:"code,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Errors  []APIError          `json:"errors,omitempty"`
	TraceID string              `json:"trace_id,omitempty"`
}

// Resources reply for effected resources
type Resources struct {
	ResourcesEffected int `json:"resources_affected"`
}

// APIError is a baic api error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BaseResponseV1 is a basic response
type BaseResponseV1 struct {
	Errors []APIError `json:"errors,omitempty"`
}

// ExecuteFunctionResultsV1 is the response body when executing a function
type ExecuteFunctionResultsV1 struct {
	BaseResponseV1
	Resources []ExecuteFunctionResultV1 `json:"resources"`
}

// MetaInfo metadata for MSA reply
type MetaInfo struct {
	QueryTime   float64    `json:"query_time"`
	Paging      *Paging    `json:"pagination,omitempty"`
	Writes      *Resources `json:"writes,omitempty"`
	Attribution string     `json:"powered_by,omitempty"`
	TraceID     string     `json:"trace_id"`
}

// Paging paging meta
type Paging struct {
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
	Total  int64 `json:"total"`
}

// ExecuteFunctionResultV1 is the response body for executing a function
type ExecuteFunctionResultV1 struct {
	ID           string              `json:"id" description:"ID of the function that was executed in the format 'function_id:version.operation_id'."`
	StatusCode   int                 `json:"status_code" description:"The response status code from the partner service."`
	Headers      map[string][]string `json:"headers,omitempty" description:"The response headers from the partner service"`
	ResponseBody json.RawMessage     `json:"response_body,omitempty" description:"The response body from the partner service"`
}
