![CrowdStrike Falcon](/docs/asset/cs-logo.png?raw=true)

# Foundry Function as a Service Go SDK

`foundry-fn-go` is a community-driven, open source project designed to enable the authoring of functions.
While not a formal CrowdStrike product, `foundry-fn-go` is maintained by CrowdStrike and supported in partnership
with the open source developer community.

## Installation ⚙️

### Via `go get`

The SDK can be installed or updated via `go get`:

```shell
go get github.com/CrowdStrike/foundry-fn-go
```

### From source

The SDK can be built from source via standard `build`:

```shell
go mod tidy
go build .
```

## Quickstart 💫

### Code

Add the SDK to your project by following the [installation](#installation) instructions above, then create
your `main.go`:

```go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func main() {
	fdk.Run(context.Background(), newHandler)
}

type request struct {
	Name string `json:"name"`
	Val  string `json:"val"`
}

// newHandler here is showing how a config is integrated. It is using generics,
// so we can unmarshal the config into a concrete type and then validate it. The
// OK method is run to validate the contents of the config.
func newHandler(_ context.Context, logger *slog.Logger, cfg config) fdk.Handler {
	mux := fdk.NewMux()
	mux.Get("/name", fdk.HandlerFn(func(_ context.Context, r fdk.Request) fdk.Response {
		return fdk.Response{
			Body: fdk.JSON(map[string]string{"name": r.Params.Query.Get("name")}),
			Code: 200,
		}
	}))
	mux.Post("/echo", fdk.HandlerFnOfOK(func(_ context.Context, r fdk.RequestOf[request]) fdk.Response {
		if r.Body.Name == "kaboom" {
			logger.Error("encountered the kaboom")
		}
		return fdk.Response{
			Body:   fdk.JSON(r.Body),
			Code:   201,
			Header: http.Header{"X-Cs-Method": []string{r.Method}},
		}
	}))
	return mux
}

type config struct {
	Int int    `json:"integer"`
	Str string `json:"string"`
}

func (c config) OK() error {
	var errs []error
	if c.Int < 1 {
		errs = append(errs, errors.New("integer must be greater than 0"))
	}
	if c.Str == "" {
		errs = append(errs, errors.New("non empty string must be provided"))
	}
	return errors.Join(errs...)
}

```

1. `config`: A type the raw json config is unmarshalled into.
2. `logger`: A dedicated logger is provided to capture function logs in all environments (both locally and distributed).
    1. Using a different logger may produce logs in the runtime but won't make it into the logscale infrastructure.
3. `Request`: Request payload and metadata. At the time of this writing, the `Request` struct consists of:
    1. `Body`:  The input io.Reader for the payload as given in the Function Gateway `body` payload field or streamed
       in.
    2. `Params`: Contains request headers and query parameters.
    3. `URL`: The request path relative to the function as a string.
    4. `Method`: The request HTTP method or verb.
    5. `Context`: Caller-supplied raw context.
    6. `AccessToken`: Caller-supplied access token.
4. `RequestOf`: The same as Request only that the Body field is json unmarshalled into the generic type (i.e. `request`
   type above)
5. `Response`
    1. The `Response` contains fields `Body` (the payload of the response), `Code` (an HTTP status code),
       `Errors` (a slice of `APIError`s), and `Headers` (a map of any special HTTP headers which should be present on
       the response).
6. `main()`: Initialization and bootstrap logic all contained with fdk.Run and handler constructor.

more examples can be found at:

* [fn with config](examples/fn_config)
* [fn without config](examples/fn_no_config)
* [more complex/complete example](examples/complex)

### Testing locally

The SDK provides an out-of-the-box runtime for executing the function.
A basic HTTP server will be listening on port 8081.

```shell
# build the project which uses the sdk
cd my-project && go mod tidy && go build -o run_me .

# run the executable. config should be in json format here.
CS_FN_CONFIG_PATH=$PATH_TO_CONFIG_JSON ./run_me
```

Requests can now be made against the executable.

```shell
curl -X POST http://localhost:8081/ \
  -H "Content-Type: application/json" \
  --data '{
    "body": {
        "foo": "bar"
    },
    "method": "POST",
    "url": "/greetings"
}'
```

## Convenience Functionality 🧰

### `gofalcon`

Foundry Function integrates with [gofalcon](https://github.com/CrowdStrike/gofalcon) in a few simple lines.

```go
package main

import (
	"context"
	"log/slog"

	fdk "github.com/CrowdStrike/foundry-fn-go"
	"github.com/CrowdStrike/gofalcon/falcon"
	"github.com/CrowdStrike/gofalcon/falcon/client"
)

func newHandler(_ context.Context, _ *slog.Logger, cfg config) fdk.Handler {
	mux := fdk.NewMux()
	mux.Post("/echo", fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		client, err := newFalconClient(ctx, r.AccessToken)
		if err != nil {
			if err == falcon.ErrFalconNoToken {
				// not a processable request
				return fdk.Response{ /* snip */ }
			}
			// some other error - see gofalcon documentation
		}

		// trim rest
	}))
	return mux
}

func newFalconClient(ctx context.Context, token string) (*client.CrowdStrikeAPISpecification, error) {
	opts := fdk.FalconClientOpts()
	return falcon.NewClient(&falcon.ApiConfig{
		AccessToken:       token,
		Cloud:             falcon.Cloud(opts.Cloud),
		Context:           ctx,
		UserAgentOverride: out.UserAgent,
	})
}

// omitting rest of implementation

```

## Integration with Falcon Fusion workflows

When integrating with a Falcon Fusion workflow, the `Request.Context` can be decoded into
`WorkflowCtx` type. You may json unmarshal into that type. The type provides some additional
context from the workflow. This context is from the execution of the workflow, and may be
dynamic in some usecases. To simplify things further for authors, we have introduced two
handler functions to remove the boilerplate of dealing with a workflow.

```go
package somefn

import (
	"context"
	"log/slog"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

type reqBody struct {
	Foo string `json:"foo"`
}

func New(ctx context.Context, _ *slog.Logger, _ fdk.SkipCfg) fdk.Handler {
	m := fdk.NewMux()

	// for get/delete reqs use HandleWorkflow. The path is just an examples, any payh can be used.
	m.Get("/workflow", fdk.HandleWorkflow(func(ctx context.Context, r fdk.Request, workflowCtx fdk.WorkflowCtx) fdk.Response {
		// ... trim impl
	}))

	// for handlers that expect a request body (i.e. PATCH/POST/PUT)
	m.Post("/workflow", fdk.HandleWorkflowOf(func(ctx context.Context, r fdk.RequestOf[reqBody], workflowCtx fdk.WorkflowCtx) fdk.Response {
		// .. trim imple
	}))

	return m
}

```

## Working with Request and Response Schemas

Within the fdktest pkg, we maintain test funcs for validating a schema and its integration
with a handler. Example:

```go
package somefn_test

import (
	"context"
	"net/http"
	"testing"

	fdk "github.com/CrowdStrike/foundry-fn-go"
	"github.com/CrowdStrike/foundry-fn-go/fdktest"
)

func TestHandlerIntegration(t *testing.T) {
	reqSchema := `{
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

	respSchema := `{
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
}`
	handler := fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		return fdk.Response{Body: fdk.JSON(map[string]string{"foo": "bar"})}
	})

	req := fdk.Request{
		URL:    "/",
		Method: http.MethodPost,
		Body:   json.RawMessage(`{"postalCode": "55755"}`),
	}

	err := fdktest.HandlerSchemaOK(handler, req, reqSchema, respSchema)
	if err != nil {
		t.Fatal("unexpected err: ", err)
	}
}

```

### A note on `os.Exit`

Please refrain from using `os.Exit`. When an error is encountered, we want to return a message
to the caller. Otherwise, it'll `os.Exit` and all stakeholders will have no idea what to make
of it. Instead, use something like the following in `fdk.Run`:

```go
package main

import (
	"context"
	"log/slog"
	"net/http"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func newHandler(_ context.Context, logger *slog.Logger, _ fdk.SkipCfg) fdk.Handler {
	foo, err := newFoo()
	if err != nil {
        // leave yourself/author the nitty-gritty details and return to the end user/caller
        // a valid error that doesn't expose all the implementation details
		logger.Error("failed to create foo", "err", err.Error())
		return fdk.ErrHandler(fdk.APIError{Code: http.StatusInternalServerError, Message: "unexpected error starting function"})
	}

	mux := fdk.NewMux()
	// ...trim rest of setup

	return mux
}

```

---

<p align="center"><img src="https://raw.githubusercontent.com/CrowdStrike/falconpy/main/docs/asset/cs-logo-footer.png"><BR/><img width="250px" src="https://raw.githubusercontent.com/CrowdStrike/falconpy/main/docs/asset/adversary-red-eyes.png"></P>
<h3><P align="center">WE STOP BREACHES</P></h3>
