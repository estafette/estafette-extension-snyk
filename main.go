package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/picatz/snyk"
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
	// gitSource = kingpin.Flag("git-source", "The source of the repository.").Envar("ESTAFETTE_GIT_SOURCE").Required().String()
	// gitOwner  = kingpin.Flag("git-owner", "The owner of the repository.").Envar("ESTAFETTE_GIT_OWNER").Required().String()
	// gitName   = kingpin.Flag("git-name", "The owner plus repository name.").Envar("ESTAFETTE_GIT_NAME").Required().String()

	projectName      = kingpin.Flag("project-name", "Project name configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_PROJECT_NAME").Required().String()
	targetScore      = kingpin.Flag("target-score", "Target score configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_TARGET_SCORE").Required().String()
	organizationID   = kingpin.Flag("organization-id", "Organization ID configured at the CI server, passed in to this trusted extension.").Envar("ESTAFETTE_ORGANIZATION_ID").Required().String()
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
	client := NewSnykAPIClient(os.Getenv("ESTAFETTE_CREDENTIALS_SNYK_API_TOKEN"))

	// The Organization ID will be fixed
	orgID := os.Getenv("ESTAFETTE_ORGANIZATION_ID")
	projectName := os.Getenv("ESTAFETTE_PROJECT_NAME")
	targetScoreStr := os.Getenv("ESTAFETTE_TARGET_SCORE")
	targetScore, err := strconv.ParseFloat(targetScoreStr, 64)

	if err != nil {
		log.Fatal().Err(err).Msg("Target Score provided is not a float variable")
	}

	projects, err := client.OrganizationProjects(ctx, orgID)

	if err != nil {
		log.Fatal().Msgf(err.Error())
	}

	var vulnerabilities snyk.Vulnerabilities

	for _, project := range projects {
		if strings.HasPrefix(project.Name, projectName) {
			vulnerabilityList, err := client.ProjectVulnerabilities(ctx, orgID, project.ID, nil)
			if err != nil {
				log.Fatal().Msgf(err.Error())
			}
			vulnerabilities = append(vulnerabilities, vulnerabilityList...)
		}
	}

	var scores []float64
	for _, vulnerability := range vulnerabilities {
		fmt.Println(vulnerability.CvssScore)
		scores = append(scores, vulnerability.CvssScore)
	}

	if len(scores) == 0 {
		log.Info().Msgf("The project %s does not have vulnerabilities", projectName)
		return
	}

	// Sorts the vulnerabilities by score, takes the highest and reports if it is above threshold
	sort.Float64s(scores)

	if scores[len(scores)-1] >= targetScore {
		log.Fatal().Msgf("The project %s has vulnerabilities above the current threshold of %.1f", projectName, targetScore)
	} else {
		log.Info().Msgf("The project %s does not have vulnerabilities above the current threshold of %.1f", projectName, targetScore)
	}
}
