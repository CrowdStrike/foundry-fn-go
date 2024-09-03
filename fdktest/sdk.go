package fdktest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/xeipuuv/gojsonschema"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

// SchemaOK validates that the provided schema conforms to JSON Schema.
func SchemaOK(schema string) error {
	schemaLoader := gojsonschema.NewSchemaLoader()
	_, err := schemaLoader.Compile(gojsonschema.NewStringLoader(schema))
	return err
}

// HandlerSchemaOK validates the handler and schema integrations.
func HandlerSchemaOK(h fdk.Handler, r fdk.Request, reqSchema, respSchema string) error {
	if reqSchema != "" {
		if err := validateSchema(reqSchema, r.Body); err != nil {
			return fmt.Errorf("failed request schema validation: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp := h.Handle(ctx, r)

	if respSchema != "" {
		b, err := resp.Body.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal response payload: %w", err)
		}

		err = validateSchema(respSchema, bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("failed response schema validation: %w", err)
		}
	}

	return nil
}

func validateSchema(schema string, body io.Reader) error {
	payload, _ := io.ReadAll(body)
	result, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(schema),
		gojsonschema.NewBytesLoader(payload),
	)
	if err != nil {
		return fmt.Errorf("failed to validate document against schema: %w", err)
	}

	var errs []error
	for _, resErr := range result.Errors() {
		errMsg := resErr.String()

		// sometimes the library prefixes the string message with (root): which is confusing and is best to defer
		// to the description message in this case
		id := resErr.Field()
		if len(resErr.Details()) > 0 && resErr.Details()["property"] != nil {
			id = fmt.Sprintf("%s.%s", id, resErr.Details()["property"])
		}
		if resErr.Field() == "(root)" {
			errMsg = resErr.Description()
			if prop, ok := resErr.Details()["property"]; ok {
				id = prop.(string)
			}
		}
		errs = append(errs, errors.New(errMsg))
	}

	return errors.Join(errs...)

}
