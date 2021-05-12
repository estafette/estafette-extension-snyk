package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"

	"github.com/alecthomas/kingpin"
	"github.com/estafette/estafette-extension-snyk/api"
	"github.com/estafette/estafette-extension-snyk/clients/credentials"
	"github.com/estafette/estafette-extension-snyk/clients/snykcli"
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
)

var (
	// parameters
	failOn            = kingpin.Flag("fail-on", "Fail on all|upgradable|patchable.").Default("all").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_FAIL_ON").Enum("all", "upgradable", "patchable")
	file              = kingpin.Flag("file", "Path to file to run analysis for.").Envar("ESTAFETTE_EXTENSION_FILE").String()
	packagesFolder    = kingpin.Flag("packages-folder", "This is the folder in which your dependencies are installed.").Envar("ESTAFETTE_EXTENSION_PACKAGES_FOLDER").String()
	severityThreshold = kingpin.Flag("severity-threshold", "The minimum severity to fail on.").Default("high").OverrideDefaultFromEnvar("ESTAFETTE_EXTENSION_SEVERITY_THRESHOLD").Enum("low", "medium", "high")

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

	// get api token from injected credentials
	credentialsClient := credentials.NewClient()
	token, err := credentialsClient.GetToken(ctx, *snykAPITokenPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed getting snyk api token from injected credentials")
	}

	snykcliClient := snykcli.NewClient(token)
	extensionService := extension.NewService(snykcliClient)

	// check if there's a single sln file and set file argument
	if *file == "" {
		files, err := checkExt(".sln")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed getting sln files")
		}
		if len(files) == 1 {
			*file = files[0]
			log.Info().Msgf("Autodetected file %v and using it as 'file' parameter", files[0])
			// restoring first, otherwise it fails (and we can't inject a stage in the mi)
			foundation.RunCommand(ctx, "dotnet restore --packages .nuget/packages")
		}
	}

	flags := api.SnykFlags{
		FailOn:            *failOn,
		File:              *file,
		PackagesFolder:    *packagesFolder,
		SeverityThreshold: *severityThreshold,
	}

	err = extensionService.Run(ctx, flags)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed running status check")
	}
}

func checkExt(ext string) ([]string, error) {
	var files []string
	pathS, err := os.Getwd()
	if err != nil {
		return files, err
	}
	filepath.Walk(pathS, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			r, err := regexp.MatchString(ext, f.Name())
			if err == nil && r {
				files = append(files, f.Name())
			}
		}
		return nil
	})
	return files, nil
}
