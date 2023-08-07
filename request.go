package fdk

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// Request is a default fdk request
type Request struct {
	Body        json.RawMessage `json:"body,omitempty"`
	Params      *Params         `json:"params,omitempty"`
	URL         string          `json:"url"`
	Method      string          `json:"method"`
	Context     json.RawMessage `json:"context,omitempty"`
	AccessToken string          `json:"access_token,omitempty"`
}

// Fields returns an ordered slice of fields from a workflow.
func (r *Request) Fields() []*Field {
	fields := make([]*Field, 0)

	fieldCtx := struct {
		Fields []struct {
			Name    string      `json:"name"`
			Display string      `json:"display"`
			Kind    string      `json:"kind"`
			Value   interface{} `json:"value"`
		} `json:"fields"`
	}{}

	if err := json.Unmarshal(r.Context, &fieldCtx); err != nil {
		// If we cannot parse the context return an empty slice.
		return fields
	}

	for _, val := range fieldCtx.Fields {
		if val.Name == "" || val.Kind == "" || val.Display == "" || val.Value == nil {
			// Required information is missing so do not append and continue.
			continue
		}

		fields = append(fields, &Field{
			Name:    val.Name,
			Display: val.Display,
			Kind:    val.Kind,
			Value:   val.Value,
		})
	}

	return fields
}

// Field is the metadata and value of a field from a workflow.
type Field struct {
	Name    string      `json:"name"`
	Display string      `json:"display"`
	Kind    string      `json:"kind"`
	Value   interface{} `json:"value"`
}

// Params represents the request params
type Params struct {
	Header http.Header `json:"header,omitempty"`
	Query  url.Values  `json:"query,omitempty"`
}

// Config is a simple config map
type Config map[string]interface{}

// ExecuteFunctionV1 holds request information to execute a function
type ExecuteFunctionV1 struct {
	ID       string  `json:"id" description:"ID of the function to execute, in the format 'function_id:version' or 'function_id'."`
	ConfigID string  `json:"config_id" description:"ConfigID for the function to use."`
	Request  Request `json:"request" description:"The request params/body to execute the command. The data model for this section is determined at runtime via the request schema."`
}

// ExecuteOperationV1 holds request information to execute an operation
type ExecuteOperationV1 struct {
	ID       string  `json:"id" description:"ID of the operation to execute, in the format 'operation_id:version' or 'operation_id'."`
	ConfigID string  `json:"config_id" description:"ConfigID for the function to use."`
	Request  Request `json:"request" description:"The request params/body to execute the command. The data model for this section is determined at runtime via the request schema."`
}

// ExecuteCommandRequestV1 stores information for execute command requests
type ExecuteCommandRequestV1 struct {
	Resources []ExecuteFunctionV1 `json:"resources" description:"List of functions to execute"`
}
