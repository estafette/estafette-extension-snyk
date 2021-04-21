package snykcli

import (
	"context"

	foundation "github.com/estafette/estafette-foundation"
)

type Client interface {
	Auth(ctx context.Context) (err error)
	Test(ctx context.Context, severityThreshold, failOn, file string) (err error)
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

func (c *client) Test(ctx context.Context, severityThreshold, failOn, file string) (err error) {
	// snyk auth (https://support.snyk.io/hc/en-us/articles/360003812578-CLI-reference)
	command := "snyk test"
	if severityThreshold != "" {
		command += " --severity-threshold=" + severityThreshold
	}
	if failOn != "" {
		command += " --fail-on=" + failOn
	}
	if file != "" {
		command += " --file" + file
	}

	err = foundation.RunCommandExtended(ctx, command)
	if err != nil {
		return
	}

	return
}
