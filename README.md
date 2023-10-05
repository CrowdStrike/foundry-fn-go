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
	"encoding/json"
	"errors"
	"net/http"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func main() {
	fdk.Run(context.Background(), newHandler)
}

// newHandler here is showing how a config is integrated. It is using generics,
// so we can unmarshal the config into a concrete type and then validate it. The
// OK method is run to validate the contents of the config.
func newHandler(_ context.Context, cfg config) fdk.Handler {
	mux := fdk.NewMux()
	mux.Post("/echo", fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		return fdk.Response{
			Body:   r.Body,
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
2. `Request`: Request payload and metadata. At the time of this writing, the `Request` struct consists of:
    1. `Body`: The raw request payload as given in the Function Gateway `body` payload field.
    2. `Params`: Contains request headers and query parameters.
    3. `URL`: The request path relative to the function as a string.
    4. `Method`: The request HTTP method or verb.
    5. `Context`: Caller-supplied raw context.
    6. `AccessToken`: Caller-supplied access token.
3. `Response`
   1. The `Response` contains fields `Body` (the payload of the response), `Code` (an HTTP status code),
      `Errors` (a slice of `APIError`s), and `Headers` (a map of any special HTTP headers which should be present on
      the response).
4. `main()`: Initialization and bootstrap logic all contained with fdk.Run and handler constructor.

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

# run the executable
./run_me
```

Requests can now be made against the executable.

```shell
curl -X POST http://localhost:8081/echo \
  -H "Content-Type: application/json" \
  --data '{"foo": "bar"}'
```

## Convenience Functionality 🧰

### `gofalcon`

Foundry Function Go ships with [gofalcon](https://github.com/CrowdStrike/gofalcon) pre-integrated and a convenience
constructor.
While it is not strictly necessary to use convenience function, it is recommended.

**Important:** Create a new instance of the `gofalcon` client on each request.

```go
package main

import (
	"context"
	"net/http"

	/* omitting other imports */
	fdk "github.com/crowdstrike/foundry-fn-go"
)

func newHandler(_ context.Context, cfg config) fdk.Handler {
	mux := fdk.NewMux()
	mux.Post("/echo", fdk.HandlerFn(func(ctx context.Context, r fdk.Request) fdk.Response {
		client, err := fdk.FalconClient(ctx, request)
		if err != nil {
			if err == fdk.ErrFalconNoToken {
				// not a processable request
				return fdk.Response{ /* snip */ }
			}
			// some other error - see gofalcon documentation
		}

		// trim rest
	}))
	return mux
}

// omitting rest of implementation
```

---


<p align="center"><img src="https://raw.githubusercontent.com/CrowdStrike/falconpy/main/docs/asset/cs-logo-footer.png"><BR/><img width="250px" src="https://raw.githubusercontent.com/CrowdStrike/falconpy/main/docs/asset/adversary-red-eyes.png"></P>
<h3><P align="center">WE STOP BREACHES</P></h3>
