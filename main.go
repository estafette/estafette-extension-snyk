package main

import (
	"context"

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

	mavenMirrorUrl = kingpin.Flag("maven-mirror-url", "Maven mirror to use for fetching packages.").Envar("ESTAFETTE_EXTENSION_MAVEN_MIRROR_URL").String()
	mavenUsername  = kingpin.Flag("maven-user", "Maven mirror username.").Envar("ESTAFETTE_EXTENSION_MAVEN_USERNAME").String()
	mavenPassword  = kingpin.Flag("maven-password", "Maven mirror password.").Envar("ESTAFETTE_EXTENSION_MAVEN_PASSWORD").String()

	// injected credentials
	snykAPITokenPath = kingpin.Flag("snyk-api-token-path", "Snyk api token credentials configured at the CI server, passed in to this trusted extension.").Default("/credentials/snyk_api_token.json").String()

	versionGo     = kingpin.Flag("version-go", "Version info.").Envar("EXT_GO_VERSION").String()
	versionNode   = kingpin.Flag("version-node", "Version info.").Envar("EXT_NODE_VERSION").String()
	versionNpm    = kingpin.Flag("version-npm", "Version info.").Envar("EXT_NPM_VERSION").String()
	versionJava   = kingpin.Flag("version-java", "Version info.").Envar("EXT_JAVA_VERSION").String()
	versionMaven  = kingpin.Flag("version-maven", "Version info.").Envar("EXT_MAVEN_VERSION").String()
	versionDotnet = kingpin.Flag("version-dotnet", "Version info.").Envar("EXT_DOTNET_VERSION").String()
	versionPython = kingpin.Flag("version-python", "Version info.").Envar("EXT_PYTHON_VERSION").String()
	versionPip    = kingpin.Flag("version-pip", "Version info.").Envar("EXT_PIP_VERSION").String()
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

	flags := api.SnykFlags{
		Language:          api.LanguageUnknown,
		FailOn:            *failOn,
		File:              *file,
		PackagesFolder:    *packagesFolder,
		SeverityThreshold: *severityThreshold,

		MavenMirrorUrl: *mavenMirrorUrl,
		MavenUsername:  *mavenUsername,
		MavenPassword:  *mavenPassword,
	}

	toolVersions := api.ToolVersions{
		Go:     *versionGo,
		Node:   *versionNode,
		Npm:    *versionNpm,
		Java:   *versionJava,
		Maven:  *versionMaven,
		Dotnet: *versionDotnet,
		Python: *versionPython,
		Pip:    *versionPip,
	}

	flags, err = extensionService.AugmentFlags(ctx, flags, toolVersions)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed augmenting flags")
	}

	err = extensionService.Run(ctx, flags)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed running status check")
	}
}
