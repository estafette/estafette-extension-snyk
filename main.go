package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/estafette/estafette-extension-snyk/pkg/api"
	"github.com/estafette/estafette-extension-snyk/pkg/clients/credentials"
	"github.com/estafette/estafette-extension-snyk/pkg/clients/snykcli"
	"github.com/estafette/estafette-extension-snyk/pkg/services/extension"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/rs/zerolog/log"
)

var (
	appgroup  string
	app       string
	version   string
	branch    string
	revision  string
	buildDate string
)

var (
	// parameters
	failOn             = kingpin.Flag("fail-on", "Fail on all|upgradable|patchable.").Default("all").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_FAIL_ON").Enum("all", "upgradable", "patchable")
	packagesFolder     = kingpin.Flag("packages-folder", "This is the folder in which your dependencies are installed.").Envar("ESTAFETTE_EXTENSION_PACKAGES_FOLDER").String()
	severityThreshold  = kingpin.Flag("severity-threshold", "The minimum severity to fail on.").Default("high").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_SEVERITY_THRESHOLD").Enum("low", "medium", "high")
	excludeDirectories = kingpin.Flag("exclude", "Exclude directories from scan.").Default("test").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_EXCLUDE").String()
	debug              = kingpin.Flag("debug", "Print debug information.").Envar("ESTAFETTE_EXTENSION_DEBUG").Bool()
	scan               = kingpin.Flag("scan", "Scan repository for vulnerabilities.").Default("true").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_SCAN").Bool()

	// injected credentials
	snykAPITokenPath = kingpin.Flag("snyk-api-token-path", "Snyk api token credentials configured at the CI server, passed in to this trusted extension.").Default("/credentials/snyk_api_token.json").String()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// init log format from envvar ESTAFETTE_LOG_FORMAT
	foundation.InitLoggingFromEnv(foundation.NewApplicationInfo(appgroup, app, version, branch, revision, buildDate))

	// create context to cancel commands on sigterm
	ctx := foundation.InitCancellationContext(context.Background())

	if !*scan {
		log.Info().Msg("Scanning is disabled, exiting")
		return
	}

	// get api token from injected credentials
	credentialsClient := credentials.NewClient(*snykAPITokenPath)
	token, err := credentialsClient.GetToken(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed getting snyk api token from injected credentials")
	}

	snykcliClient := snykcli.NewClient(token)
	extensionService := extension.NewService(credentialsClient, snykcliClient)

	flags := api.SnykFlags{
		FailOn:             *failOn,
		PackagesFolder:     *packagesFolder,
		GroupName:          fmt.Sprintf("%v/%v/%v", os.Getenv("ESTAFETTE_GIT_SOURCE"), os.Getenv("ESTAFETTE_GIT_OWNER"), os.Getenv("ESTAFETTE_GIT_NAME")),
		SeverityThreshold:  *severityThreshold,
		ExcludeDirectories: strings.Split(*excludeDirectories, ","),
		Debug:              *debug,
	}

	err = extensionService.Run(ctx, flags)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed running status check")
	}
}
