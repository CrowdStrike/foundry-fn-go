package fdk

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/crowdstrike/gofalcon/falcon"
	"github.com/crowdstrike/gofalcon/falcon/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_convertRequest(t *testing.T) {

	payload := bytes.NewBuffer([]byte(`
{
	"access_token": "ad669dff-36df-4713-9c61-3e6840f3d675",
	"body": {
		"e5eed0ce688b": "36cc4e7e2d2a"
	},
	"context": {
		"fields": [
			{
				"name": "a",
				"display": "A",
				"kind": "foo",
				"value": "baz"
			}
		]
	},
	"method": "POST",
	"params": {
		"header": {
			"cc7f08e30972": ["203a5cfeabb4"]
		},
		"query": {
			"7226f87a4afa": ["ed368691e4cc"]
		}
	},
	"url": "/1c6f79942cba"
}`))
	want := Request{
		AccessToken: "ad669dff-36df-4713-9c61-3e6840f3d675",
		Body: []byte(`{
		"e5eed0ce688b": "36cc4e7e2d2a"
	}`),
		Context: json.RawMessage(`{
		"fields": [
			{
				"name": "a",
				"display": "A",
				"kind": "foo",
				"value": "baz"
			}
		]
	}`),
		Method: "POST",
		Params: &Params{
			Header: http.Header{
				"cc7f08e30972": []string{"203a5cfeabb4"},
			},
			Query: url.Values{
				"7226f87a4afa": []string{"ed368691e4cc"},
			},
		},
		URL: "/1c6f79942cba",
	}

	req, err := http.NewRequest("GET", "/", payload)
	require.NoError(t, err)

	got, err := convertRequest(req)
	if err != nil {
		t.Errorf("convertRequest() error = %v, wantErr nil", err)
		return
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("convertRequest() got = %+v, want = %+v", got, want)
	}
}

func TestLoadConfig(t *testing.T) {
	type sample struct {
		Hello string `json:"hello"`
	}
	type testCase struct {
		name        string
		setEnvFn    func(t *testing.T)
		wantsErr    bool
		wantsObject *sample
	}
	testCases := []testCase{
		{
			name: "reading an existing json file populates object",
			setEnvFn: func(t *testing.T) {
				t.Setenv(envFnConfigPath, "test_data/load_config/success.json")
			},
			wantsErr:    false,
			wantsObject: &sample{Hello: "world"},
		},
		{
			name: "reading a config from env var should populate object",
			setEnvFn: func(t *testing.T) {
				t.Setenv("FN_CONFIG", base64Encode(t, `{"hello":"blueskys"}`))
			},
			wantsErr:    false,
			wantsObject: &sample{Hello: "blueskys"},
		},
		{
			name: "reading an existing file which is not json returns error",
			setEnvFn: func(t *testing.T) {
				t.Setenv(envFnConfigPath, "test_data/load_config/raw_text.txt")
			},
			wantsErr:    true,
			wantsObject: nil,
		},
		{
			name: "reading a non-existing file returns error",
			setEnvFn: func(t *testing.T) {
				t.Setenv(envFnConfigPath, "test_data/load_config/abc.json")
			},
			wantsErr:    true,
			wantsObject: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setEnvFn(t)
			actualObject := &sample{}
			err := LoadConfig(actualObject)
			if !((tc.wantsErr && err != nil) || (!tc.wantsErr && err == nil)) {
				t.Errorf("expected err %t but got error as %v", tc.wantsErr, err)
			}

			if tc.wantsObject == nil && actualObject.Hello != "" {
				t.Errorf("expected object to not be populated but got %v", actualObject)
			}
			if tc.wantsObject != nil && actualObject.Hello != tc.wantsObject.Hello {
				t.Errorf("expected object %v but got %v", tc.wantsObject, actualObject)
			}
		})
	}
}

func base64Encode(t *testing.T, v string) string {
	t.Helper()
	return base64.StdEncoding.EncodeToString([]byte(v))
}

func TestFalconClient(t *testing.T) {
	var accessToken string
	var cloud falcon.CloudType

	factory := func(cfg *falcon.ApiConfig) (*client.CrowdStrikeAPISpecification, error) {
		accessToken = cfg.AccessToken
		cloud = cfg.Cloud
		return &client.CrowdStrikeAPISpecification{}, nil
	}

	tests := []struct {
		name     string
		exec     func(t *testing.T) (*client.CrowdStrikeAPISpecification, error)
		setup    func(t *testing.T)
		tearDown func(t *testing.T)
		wants    func(t *testing.T, c *client.CrowdStrikeAPISpecification, err error)
	}{
		{
			name: "given no valid cloud and no access token, return an error",
			setup: func(t *testing.T) {
				accessToken = ""
				cloudProvider = func() string { return "" }
				fcFactory = factory
			},
			tearDown: func(t *testing.T) {
				cloudProvider = defaultCloudProvider
				fcFactory = defaultFalconClientFactory
			},
			exec: func(t *testing.T) (*client.CrowdStrikeAPISpecification, error) {
				t.Helper()
				return FalconClient(context.Background(), Request{AccessToken: ""})
			},
			wants: func(t *testing.T, c *client.CrowdStrikeAPISpecification, err error) {
				t.Helper()
				assert.Error(t, err)
				assert.Equal(t, ErrFalconNoToken, err)
				assert.Nil(t, c)
			},
		},
		{
			name: "given a valid cloud and access token, construct a valid client",
			setup: func(t *testing.T) {
				accessToken = "abc"
				cloudProvider = func() string { return "us-2" }
				fcFactory = factory
			},
			tearDown: func(t *testing.T) {
				cloudProvider = defaultCloudProvider
				fcFactory = defaultFalconClientFactory
			},
			exec: func(t *testing.T) (*client.CrowdStrikeAPISpecification, error) {
				t.Helper()
				return FalconClient(context.Background(), Request{AccessToken: "abc"})
			},
			wants: func(t *testing.T, c *client.CrowdStrikeAPISpecification, err error) {
				t.Helper()
				assert.NoError(t, err)
				assert.NotNil(t, c)
				assert.Equal(t, "abc", accessToken)
				assert.Equal(t, falcon.CloudType(falcon.CloudUs2), cloud)
			},
		},
		{
			name: "given an empty cloud and valid access token, construct a valid client",
			setup: func(t *testing.T) {
				accessToken = "abc"
				cloudProvider = func() string { return "" }
				fcFactory = factory
			},
			tearDown: func(t *testing.T) {
				cloudProvider = defaultCloudProvider
				fcFactory = defaultFalconClientFactory
			},
			exec: func(t *testing.T) (*client.CrowdStrikeAPISpecification, error) {
				t.Helper()
				return FalconClient(context.Background(), Request{AccessToken: "abc"})
			},
			wants: func(t *testing.T, c *client.CrowdStrikeAPISpecification, err error) {
				t.Helper()
				assert.NoError(t, err)
				assert.NotNil(t, c)
				assert.Equal(t, "abc", accessToken)
				assert.Equal(t, falcon.CloudType(falcon.CloudUs1), cloud)
			},
		},
	}

	for _, tt := range tests {
		tt.setup(t)
		t.Run(tt.name, func(t *testing.T) {
			c, err := tt.exec(t)
			tt.wants(t, c, err)
		})
		tt.tearDown(t)
	}
}
