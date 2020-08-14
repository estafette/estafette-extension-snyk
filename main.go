package main

import (
	"context"
	"encoding/json"
	"runtime"

	"github.com/alecthomas/kingpin"
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
	gitName   = kingpin.Flag("git-name", "The owner plus repository name.").Envar("ESTAFETTE_GIT_NAME").Required().String()

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
	snykAPIToken := ""
	if *snykAPITokenJSON != "" {
		log.Info().Msg("Unmarshalling injected snyk api token credentials")
		var credentials []APITokenCredentials
		err := json.Unmarshal([]byte(*snykAPITokenJSON), &credentials)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed unmarshalling injected snyk api token credentials")
		}
		if len(credentials) == 0 {
			log.Fatal().Msg("No snyk api token credentials have been injected")
		}
		snykAPIToken = credentials[0].AdditionalProperties.Token
	}

	if snykAPIToken != "" {
		log.Debug().Msg("Extracted snyk API token from credentials")
	} else {
		log.Fatal().Msg("Failed to extract snyk API token from credentials")
	}

	// todo use token to communicate with snyk api
	apiClient := NewApiClient(snykAPIToken)

	status, err := apiClient.GetStatus(ctx, *gitSource, *gitOwner, *gitName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed retrieving status from Snyk API")
	}

	log.Info().Msgf("Snyk API returned status: %v", status)
}
