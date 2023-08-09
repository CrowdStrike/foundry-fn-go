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

Add the SDK to your project by following the [installation](#installation) instructions above, then create your `main.go`:

```go
package main

import (
    /* omitting other imports */
    fdk "github.com/CrowdStrike/foundry-fn-go"
)

var (
    cfg *fdk.Config /*** (1) ***/
)

/*** (2) ***/
func RunHandler(_ context.Context, request fdk.Request /*** (3) ***/) (fdk.Response, error) {
    b := bytes.NewBuffer(nil)
    err := json.NewEncoder(b).Encode(request)
    return fdk.Response{Body: b.Bytes(), Code: 200}, err /*** (4) ***/
}

func main() { /*** (5) ***/
    cfg = &fdk.Config{}
    err := fdk.LoadConfig(cfg) /*** (6) ***/
    if err != nil && err != fdk.ErrNoConfig {
        os.Exit(1)
    }
    fdk.Start(RunHandler) /*** (7) ***/
}
```

1. `cfg`: A global variable which holds any loaded configuration. Should be initialized exactly once within `main()`.
2. `RunHandler()`: Called once on each inbound request. This is where the business logic of the function should
   exist.
3. `request`: Request payload and metadata. At the time of this writing, the `Request` struct consists of:
    1. `Body`: The raw request payload as given in the Function Gateway `body` payload field.
    2. `Params`: Contains request headers and query parameters.
    3. `URL`: The request path relative to the function as a string.
    4. `Method`: The request HTTP method or verb.
    5. `Context`: Caller-supplied raw context.
    6. `AccessToken`: Caller-supplied access token.
4. Return from `RunHandler()`: Returns two values - a `Response` and an `error`.
    1. The `Response` contains fields `Body` (the payload of the response), `Code` (an HTTP status code),
       `Errors` (a slice of `APIError`s), and `Headers` (a map of any special HTTP headers which should be present on
       the response).
    2. The `error` is an indication that there was a transient error on the request.
       In production, returning a non-nil `error` will result in a retry of the request.
       There is a limited number of reties before the request will be dead-lettered.
       ***Do not return an `error` unless you actually want the request to be retried.***
5. `main()`: Initialization and bootstrap logic. Called exactly once at the startup of the runtime.
   Any initialization logic should exist prior to calling `fdk.Start()`.
6. `LoadConfig(cfg)`: Loads any configuration deployed along with function into the `*fdk.Config`.
   Will return an `ErrNoConfig` if no configuration was available to be loaded.
   Will return another `error` if there was some other error loading the configuration.
7. `fdk.Start(RunHandler)`: Binds the handler function to the runtime and starts listening to inbound requests.

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

Foundry Function Go ships with [gofalcon](https://github.com/CrowdStrike/gofalcon) pre-integrated and a convenience constructor.
While it is not strictly necessary to use convenience function, it is recommended.

**Important:** Create a new instance of the `gofalcon` client on each request.

```go
import (
    /* omitting other imports */
    fdk "github.com/crowdstrike/foundry-fn-go"
)

func RunHandler(ctx context.Context, request fdk.Request) (fdk.Response, error) {
    /* ... omitting other code ... */
    
    // !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
    // !!! create a new client instance on each request !!!
    // !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
    
    client, err := fdk.FalconClient(ctx, request)
    if err != nil {
        if err == fdk.ErrFalconNoToken {
            // not a processable request
            return fdk.Response{ /* snip */ }, nil
        }
        // some other error - see gofalcon documentation
    }
    
    /* ... omitting other code ... */
}
```

---


<p align="center"><img src="https://raw.githubusercontent.com/CrowdStrike/falconpy/main/docs/asset/cs-logo-footer.png"><BR/><img width="250px" src="https://raw.githubusercontent.com/CrowdStrike/falconpy/main/docs/asset/adversary-red-eyes.png"></P>
<h3><P align="center">WE STOP BREACHES</P></h3>
