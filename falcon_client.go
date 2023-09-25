package fdk

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/crowdstrike/gofalcon/falcon"
	"github.com/crowdstrike/gofalcon/falcon/client"
)

const (
	// version marks the version of the fdk for use in the falcon client. This should be
	// provided via an LDFlag on build.
	version = "development"
)

// FalconClient returns a new instance of the GoFalcon client.
// If the client cannot be created or if there is no access token in the request,
// an error is returned.
func FalconClient(ctx context.Context, r Request) (*client.CrowdStrikeAPISpecification, error) {
	token := strings.TrimSpace(r.AccessToken)
	if token == "" {
		return nil, errors.New("falcon client requires an access token")
	}

	c := strings.ToLower(os.Getenv("CS_CLOUD"))
	c = strings.ReplaceAll(c, "-", "")
	c = strings.TrimSpace(c)
	if c == "" {
		c = "us-1"
	}
	cloud := falcon.Cloud(c)

	return falcon.NewClient(&falcon.ApiConfig{
		AccessToken:       token,
		Cloud:             cloud,
		Context:           ctx,
		UserAgentOverride: fmt.Sprintf("foundry-fn/%s", version),
	})
}
