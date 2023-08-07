package fdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/crowdstrike/gofalcon/falcon"
	"github.com/crowdstrike/gofalcon/falcon/client"
)

const (
	defaultTimeout  = 15 * time.Second
	envCsCloud      = "CS_CLOUD"
	envFnConfigPath = "CS_FN_CONFIG_PATH"

	cloudDefault = "us-1"
)

var (
	acceptingConnections int32

	// ErrConfigPointer is returned when attempting to load config into a non-pointer interface{}
	ErrConfigPointer = errors.New("config must be loaded into a pointer type")

	// ErrNoConfig indicates that no config env variable was provided.
	ErrNoConfig = fmt.Errorf("no value provided for %s", envFnConfigPath)

	// ErrFalconNoToken indicates that no access token was on the request handed to the FalconClient convenience function.
	ErrFalconNoToken = errors.New("falcon client requires an access token")
)

var (
	cloudProvider = defaultCloudProvider
	fcFactory     = defaultFalconClientFactory
)

// Handler is the type of function that the authors define to handle incoming requests.
type Handler func(context.Context, Request) (Response, error)

// Start is called with a handler of various function handler signatures
// It will start a local http server serving your function
func Start(f Handler) {
	readTimeout := parseIntOrDurationValue(os.Getenv("read_timeout"), defaultTimeout)
	writeTimeout := parseIntOrDurationValue(os.Getenv("write_timeout"), defaultTimeout)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8081),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	http.Handle("/", &handlerWrapper{h: f})
	listen(s, writeTimeout)
}

// LoadConfig loads the configuration contents into the passed in interface type
func LoadConfig(config interface{}) error {
	if reflect.ValueOf(config).Kind() != reflect.Ptr {
		return ErrConfigPointer
	}

	var configBytes []byte
	if v2Config := os.Getenv("FN_CONFIG"); v2Config != "" {
		var err error
		configBytes, err = base64.StdEncoding.DecodeString(v2Config)
		if err != nil {
			return fmt.Errorf("unable to decode $FN_CONFIG var: %w", err)
		}
	} else {
		configPath := strings.TrimSpace(os.Getenv(envFnConfigPath))
		if configPath == "" {
			return ErrNoConfig
		}

		var err error
		configBytes, err = os.ReadFile(filepath.Clean(configPath))
		if err != nil {
			return fmt.Errorf("failed to read config file %s with error: %s", configPath, err.Error())
		}
	}

	if err := json.Unmarshal(configBytes, config); err != nil {
		return fmt.Errorf("unmarshal contents as JSON with error: %s", err)
	}
	return nil
}

func convertRequest(req *http.Request) (Request, error) {
	r := Request{
		Body: nil,
		Params: &Params{
			Header: req.Header,
			Query:  req.URL.Query(),
		},
		URL:    req.URL.String(),
		Method: req.Method,
	}
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return r, err
		}
		r.Body = body

		bodyJSON := map[string]json.RawMessage{}

		if errUnmarshal := json.Unmarshal(body, &bodyJSON); errUnmarshal == nil {
			if _, ok := bodyJSON["context"]; ok {
				r.Context = bodyJSON["context"]
			}
		} else {
			log.Println(errUnmarshal)
		}
	}
	return r, nil
}

// handlerWrapper encapsulates the customer provided handler adding functionality for
type handlerWrapper struct {
	h Handler
}

// ServeHTTP is our http Handler for calling the underlying handler function
func (wr *handlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	internalRequest, err := convertRequest(r)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	out, err := wr.h(r.Context(), internalRequest)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(out.Code)
	_, err = w.Write(out.Body)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}

func listen(s *http.Server, shutdownTimeout time.Duration) {
	idleConnsClosed := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM)

		<-sig

		log.Printf("[entrypoint] SIGTERM received.. shutting down server in %s\n", shutdownTimeout.String())

		<-time.After(shutdownTimeout)

		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("[entrypoint] Error in Shutdown: %v", err)
		}

		log.Printf("[entrypoint] No new connections allowed. Exiting in: %s\n", shutdownTimeout.String())

		<-time.After(shutdownTimeout)

		close(idleConnsClosed)
	}()

	// Run the HTTP server in a separate go-routine.
	go func() {
		if err := s.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			log.Printf("[entrypoint] Error ListenAndServe: %v", err)
			close(idleConnsClosed)
		}
	}()

	atomic.StoreInt32(&acceptingConnections, 1)

	<-idleConnsClosed
}

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}
	return duration
}

func defaultFalconClientFactory(cfg *falcon.ApiConfig) (*client.CrowdStrikeAPISpecification, error) {
	return falcon.NewClient(cfg)
}

func defaultCloudProvider() string {
	c := strings.ToLower(os.Getenv(envCsCloud))
	c = strings.ReplaceAll(c, "-", "")
	return strings.TrimSpace(c)
}

// FalconClient returns a new instance of the GoFalcon client.
// If the client cannot be created or if there is no access token in the request,
// an error is returned.
func FalconClient(ctx context.Context, r Request) (*client.CrowdStrikeAPISpecification, error) {
	token := strings.TrimSpace(r.AccessToken)
	if token == "" {
		return nil, ErrFalconNoToken
	}

	c := cloudProvider()
	if c == "" {
		c = cloudDefault
	}
	cloud := falcon.Cloud(c)
	return fcFactory(&falcon.ApiConfig{
		AccessToken: token,
		Cloud:       cloud,
		Context:     ctx,
	})
}
