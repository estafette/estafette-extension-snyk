package snykcli

import (
	"context"
	"os/exec"
	"strings"

	"github.com/estafette/estafette-extension-snyk/pkg/api"
	foundation "github.com/estafette/estafette-foundation"
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
	err = c.monitorCore(ctx, flags, "snyk monitor", true)
	if err != nil {
		return
	}

	// err = c.monitorCore(ctx, flags, "snyk container monitor", false)
	// if err != nil {
	// 	return
	// }

	// err = c.monitorCore(ctx, flags, "snyk iac monitor", false)
	// if err != nil {
	// 	return
	// }

	return nil
}

func (c *client) Test(ctx context.Context, flags api.SnykFlags) (err error) {
	err = c.testCore(ctx, flags, "snyk test", true)
	if err != nil {
		return
	}

	// err = c.testCore(ctx, flags, "snyk container test", false)
	// if err != nil {
	// 	return
	// }

	// err = c.testCore(ctx, flags, "snyk iac test", false)
	// if err != nil {
	// 	return
	// }

	return nil
}

func (c *client) monitorCore(ctx context.Context, flags api.SnykFlags, command string, allProjects bool) (err error) {
	// snyk auth (https://support.snyk.io/hc/en-us/articles/360003812578-CLI-reference)
	if flags.GroupName != "" {
		command += " --remote-repo-url=" + flags.GroupName
	}
	if allProjects {
		command += " --all-projects"
	}
	if allProjects && len(flags.ExcludeDirectories) > 0 {
		command += " --exclude=" + strings.Join(flags.ExcludeDirectories, ",")
	}
	if flags.Debug {
		command += " -d"
	}

	err = foundation.RunCommandExtended(ctx, command)
	if err != nil {
		// EXIT CODES
		// Possible exit codes and their meaning:

		// 0: success, no vulns found
		// 1: action_needed, vulns found
		// 2: failure, try to re-run command
		// 3: failure, no supported projects detected
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 3 {
			return nil
		}

		return
	}

	return
}

func (c *client) testCore(ctx context.Context, flags api.SnykFlags, command string, allProjects bool) (err error) {
	// snyk auth (https://support.snyk.io/hc/en-us/articles/360003812578-CLI-reference)
	if flags.GroupName != "" {
		command += " --remote-repo-url=" + flags.GroupName
	}
	if allProjects {
		command += " --all-projects"
	}
	if allProjects && len(flags.ExcludeDirectories) > 0 {
		command += " --exclude=" + strings.Join(flags.ExcludeDirectories, ",")
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
		// EXIT CODES
		// Possible exit codes and their meaning:

		// 0: success, no vulns found
		// 1: action_needed, vulns found
		// 2: failure, try to re-run command
		// 3: failure, no supported projects detected
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 3 {
			return nil
		}

		return
	}

	return
}
