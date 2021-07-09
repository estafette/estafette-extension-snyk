package snykcli

import (
	"context"

	"github.com/estafette/estafette-extension-snyk/pkg/api"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/rs/zerolog/log"
)

type Client interface {
	Auth(ctx context.Context) (err error)
	Monitor(ctx context.Context, flags api.SnykFlags) (err error)
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

func (c *client) Monitor(ctx context.Context, flags api.SnykFlags) (err error) {
	// snyk auth (https://support.snyk.io/hc/en-us/articles/360003812578-CLI-reference)
	command := "snyk monitor"
	if flags.ProjectName != "" {
		command += " --project-name=" + flags.ProjectName
	}
	if flags.Debug {
		command += " -d"
	}

	err = foundation.RunCommandExtended(ctx, command)
	if err != nil {
		return
	}

	return
}

func (c *client) Test(ctx context.Context, flags api.SnykFlags) (err error) {
	for pm, paths := range flags.SubProjects {
		for _, path := range paths {
			err = c.testCore(ctx, flags, pm, path)
			if err != nil {
				return
			}
		}
	}

	return nil
}

func (c *client) testCore(ctx context.Context, flags api.SnykFlags, packageManager api.PackageManager, path string) (err error) {
	// snyk auth (https://support.snyk.io/hc/en-us/articles/360003812578-CLI-reference)
	command := "snyk test"
	if packageManager == api.PackageManagerDocker {
		command = "snyk container test"
	}
	command += " --file=" + path
	if flags.ProjectName != "" {
		command += " --project-name=" + flags.ProjectName
	}
	if flags.FailOn != "" {
		command += " --fail-on=" + flags.FailOn
	}
	if flags.PackagesFolder != "" {
		command += " --packages-folder=" + flags.PackagesFolder
	}
	if flags.SeverityThreshold != "" {
		command += " --severity-threshold=" + flags.SeverityThreshold
	}
	if flags.Debug {
		command += " -d"
	}

	err = foundation.RunCommandExtended(ctx, command)
	if err != nil {
		if packageManager.IgnoreErrors() {
			log.Warn().Err(err).Msgf("Failed testing %v application, ignoring until package manager is fully supported...", packageManager)
			return nil
		}
		return
	}

	return
}
