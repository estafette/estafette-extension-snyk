package main

import (
	"context"
	"runtime"

	"github.com/alecthomas/kingpin"
	"github.com/estafette/estafette-extension-snyk/clients/credentials"
	"github.com/estafette/estafette-extension-snyk/clients/snykapi"
	"github.com/estafette/estafette-extension-snyk/services/extension"
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
	goVersion = runtime.Version()
)

var (
	// flags
	gitSource = kingpin.Flag("git-source", "The source of the repository.").Envar("ESTAFETTE_GIT_SOURCE").Required().String()
	gitOwner  = kingpin.Flag("git-owner", "The owner of the repository.").Envar("ESTAFETTE_GIT_OWNER").Required().String()
	gitName   = kingpin.Flag("git-name", "The repository name.").Envar("ESTAFETTE_GIT_NAME").Required().String()
	gitBranch = kingpin.Flag("git-branch", "The branch.").Envar("ESTAFETTE_GIT_BRANCH").Required().String()

	minValue = kingpin.Flag("min-value", "The minimum value to get from snyk to pass the check.").Default("0").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_MIN_VALUE").Int()

	snykAPITokenJSON = kingpin.Flag("snyk-api-token", "Snyk api token credentials configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_SNYK_API_TOKEN").Required().String()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// init log format from envvar ESTAFETTE_LOG_FORMAT
	foundation.InitLoggingFromEnv(foundation.NewApplicationInfo(appgroup, app, version, branch, revision, buildDate))

	// create context to cancel commands on sigterm
	ctx := foundation.InitCancellationContext(context.Background())

	// get api token from injected credentials
	credentialsClient := credentials.NewClient()
	token, err := credentialsClient.GetToken(ctx, *snykAPITokenJSON)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed getting snyk api token from injected credentials")
	}

	snykapiClient := snykapi.NewClient(token)
	extensionService := extension.NewService(snykapiClient)

	err = extensionService.Run(ctx, *gitSource, *gitOwner, *gitName, *gitBranch, *minValue)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed running status check")
	}
}
