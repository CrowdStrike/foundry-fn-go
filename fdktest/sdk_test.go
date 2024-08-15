package fdktest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
	"github.com/CrowdStrike/foundry-fn-go/fdktest"
)

func TestHandlerSchemaOK(t *testing.T) {
	type (
		inputs struct {
			handler    fdk.Handler
			req        fdk.Request
			reqSchema  string
			respSchema string
		}

		wantFn func(t *testing.T, err error)
	)

	tests := []struct {
		name  string
		input inputs
		want  wantFn
	}{
		{
			name: "with valid req and resp schema and compliant req and resp bodies should pass",
			input: inputs{
				handler: fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
					return fdk.Response{Body: fdk.JSON(map[string]string{"foo": "bar"})}
				}),
				req: fdk.Request{
					URL:    "/",
					Method: http.MethodPost,
					Body:   bytes.NewReader(json.RawMessage(`{"postalCode": "55755"}`)),
				},
				reqSchema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "postalCode": {
      "type": "string",
      "description": "The person's first name.",
      "pattern": "\\d{5}"
    },
    "optional": {
      "type": "string",
      "description": "The person's last name."
    }
  },
  "required": [
    "postalCode"
  ]
}`,
				respSchema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "foo": {
      "type": "string",
      "description": "The person's first name.",
      "enum": ["bar"]
    }
  },
  "required": [
    "foo"
  ]
}`,
			},
			want: mustNoErr,
		},
		{
			name: "with valid req schema and invalid request body should fail",
			input: inputs{
				handler: fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
					return fdk.Response{Body: fdk.JSON(map[string]string{"foo": "bar"})}
				}),
				req: fdk.Request{
					URL:    "/",
					Method: http.MethodPost,
					Body:   bytes.NewReader(json.RawMessage(`{"postalCode": "5"}`)),
				},
				reqSchema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "postalCode": {
      "type": "string",
      "description": "The person's first name.",
      "pattern": "\\d{5}"
    }
  },
  "required": [
    "postalCode"
  ]
}`,
			},
			want: func(t *testing.T, err error) {
				errMsg := "failed request schema validation: postalCode: Does not match pattern '\\d{5}'"
				if err == nil || !strings.HasSuffix(err.Error(), errMsg) {
					t.Fatal("did not get expected error: ", err)
				}
			},
		},
		{
			name: "with valid resp schema and invalid response body should fail",
			input: inputs{
				handler: fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
					return fdk.Response{Body: fdk.JSON(map[string]string{"foo": "NOT BAR"})}
				}),
				req: fdk.Request{
					URL:    "/",
					Method: http.MethodPost,
				},
				respSchema: `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "foo": {
      "type": "string",
      "description": "The person's first name.",
      "enum": ["bar"]
    }
  },
  "required": [
    "foo"
  ]
}`,
			},
			want: func(t *testing.T, err error) {
				errMsg := "failed response schema validation: foo: foo must be one of the following: \"bar\""
				if err == nil || !strings.HasSuffix(err.Error(), errMsg) {
					t.Fatal("did not get expected error: ", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fdktest.HandlerSchemaOK(tt.input.handler, tt.input.req, tt.input.reqSchema, tt.input.respSchema)
			tt.want(t, err)
		})
	}
}

func TestSchemaOK(t *testing.T) {
	schema := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "postalCode": {
      "type": "string",
      "description": "The person's first name.",
      "pattern": "\\d{5}"
    },
    "optional": {
      "type": "string",
      "description": "The person's last name."
    }
  },
  "required": [
    "postalCode"
  ]
}`

	t.Run("with valid schema should pass", func(t *testing.T) {
		err := fdktest.SchemaOK(schema)
		mustNoErr(t, err)
	})

	t.Run("with invalid shcema should fail", func(t *testing.T) {
		invalidScheam := schema[15:]
		err := fdktest.SchemaOK(invalidScheam)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})
}

func mustNoErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal("unexpected error: " + err.Error())
	}
}
