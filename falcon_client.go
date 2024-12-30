package fdk

import (
	"fmt"
	"os"
	"strings"
)

// version marks the version of the fdk for use in the falcon client. This should be
// provided via an LDFlag on build.
var version = "development"

// FalconClientOpts provides the cloud for use with the falcon client.
// To setup the falcon Client you can follow the following example.
//
//	func newFalconClient(ctx context.Context, token string) (*client.CrowdStrikeAPISpecification, error) {
//		opts := fdk.FalconClientOpts()
//		return falcon.NewClient(&falcon.ApiConfig{
//			AccessToken:       token,
//			Cloud:             falcon.Cloud(opts.Cloud),
//			Context:           ctx,
//			UserAgentOverride: opts.UserAgent,
//		})
//	}
func FalconClientOpts() (out struct {
	Cloud     string
	UserAgent string
}) {
	c := strings.ToLower(os.Getenv("CS_CLOUD"))
	c = strings.TrimSpace(c)
	if c == "" {
		c = "us-1"
	}
	out.Cloud = c
	out.UserAgent = fmt.Sprintf("foundry-fn/%s", version)

	return out
}
