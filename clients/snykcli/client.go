package snykcli

import (
	"context"

	"github.com/estafette/estafette-extension-snyk/api"
	foundation "github.com/estafette/estafette-foundation"
)

type Client interface {
	Auth(ctx context.Context) (err error)
	Test(ctx context.Context, flags api.SnykFlags) (err error)
}

// NewClient returns a new snykapi.Client
func NewClient(apiToken string) Client {
	return &client{
		apiToken: apiToken,
	}
}

type client struct {
	apiToken string
}

func (c *client) Auth(ctx context.Context) (err error) {
	// snyk auth (https://support.snyk.io/hc/en-us/articles/360003812578-CLI-reference)
	err = foundation.RunCommandExtended(ctx, "snyk auth %v", c.apiToken)
	if err != nil {
		return
	}

	return
}

func (c *client) Test(ctx context.Context, flags api.SnykFlags) (err error) {
	// snyk auth (https://support.snyk.io/hc/en-us/articles/360003812578-CLI-reference)
	command := "snyk test"
	if flags.FailOn != "" {
		command += " --fail-on=" + flags.FailOn
	}
	if flags.File != "" {
		command += " --file=" + flags.File
	}
	if flags.PackagesFolder != "" {
		command += " --packages-folder=" + flags.PackagesFolder
	}
	if flags.SeverityThreshold != "" {
		command += " --severity-threshold=" + flags.SeverityThreshold
	}

	err = foundation.RunCommandExtended(ctx, command)
	if err != nil {
		return
	}

	return
}
