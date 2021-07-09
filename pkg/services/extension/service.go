package extension

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/estafette/estafette-extension-snyk/pkg/api"
	"github.com/estafette/estafette-extension-snyk/pkg/clients/credentials"
	"github.com/estafette/estafette-extension-snyk/pkg/clients/snykcli"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/rs/zerolog/log"
)

var (
	ErrStatusTooLow = errors.New("Returned status is too low")
)

type Service interface {
	AugmentFlags(ctx context.Context, flags api.SnykFlags, repoOwner, repoName string) (api.SnykFlags, error)
	Run(ctx context.Context, flags api.SnykFlags) (err error)
}

func NewService(credentialsClient credentials.Client, snykcliClient snykcli.Client) Service {
	return &service{
		credentialsClient: credentialsClient,
		snykcliClient:     snykcliClient,
	}
}

type service struct {
	credentialsClient credentials.Client
	snykcliClient     snykcli.Client
}

func (s *service) AugmentFlags(ctx context.Context, flags api.SnykFlags, repoOwner, repoName string) (api.SnykFlags, error) {

	log.Info().Msg("Detecting sub-projects...")

	var err error
	flags.SubProjects, err = s.detectSubProjects(ctx)
	if err != nil {
		return flags, err
	}

	log.Info().Msgf("Detected %v sub-projects", len(flags.SubProjects))

	if flags.ProjectName == "" && repoOwner != "" && repoName != "" {
		flags.ProjectName = fmt.Sprintf("%v/%v", repoOwner, repoName)
		log.Info().Msgf("Automatically set projectName to %v", flags.ProjectName)
	}

	return flags, nil
}

func (s *service) detectSubProjects(ctx context.Context) (map[api.PackageManager][]string, error) {

	// https://github.com/snyk/snyk/blob/97808254747dd7db8c7033e76dafcd46a7976d54/src/lib/package-managers.ts#L19-L49
	// https://github.com/snyk/snyk/blob/97808254747dd7db8c7033e76dafcd46a7976d54/src/lib/detect.ts#L68-L100
	supportedPackageManagerFiles := map[api.PackageManager][]string{
		api.PackageManagerNpm:       []string{"package.json"},
		api.PackageManagerMaven:     []string{"pom.xml"},
		api.PackageManagerPip:       []string{"requirements.txt", "Pipfile", "setup.py"},
		api.PackageManagerGoModules: []string{"go.mod"},
		api.PackageManagerNuget:     []string{"project.assets.json", "packages.config", "project.json", "*.sln"},
		api.PackageManagerDocker:    []string{"Dockerfile*"},
	}

	detectedSubProjects := map[api.PackageManager][]string{}
	for pm, sf := range supportedPackageManagerFiles {
		for _, sfi := range sf {
			matches, err := s.findFileMatches(".", sfi)
			if err != nil {
				continue
			}
			if len(matches) > 0 {
				// ensure the map entry is initialized
				if _, ok := detectedSubProjects[pm]; !ok {
					detectedSubProjects[pm] = []string{}
				}

				detectedSubProjects[pm] = append(detectedSubProjects[pm], matches...)
			}
		}
	}

	return detectedSubProjects, nil
}

func (s *service) findFileMatches(root, pattern string) ([]string, error) {
	var matches []string

	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := file.Name()
		if matched, err := filepath.Match(pattern, path); err != nil {
			return nil, err
		} else if matched {
			matches = append(matches, path)
		}
	}

	return matches, nil
}

func (s *service) Run(ctx context.Context, flags api.SnykFlags) (err error) {

	// run package manager specific actions
	if len(flags.SubProjects) == 0 {
		log.Info().Msg("Could not find supported package manager files, exiting...")
		return
	}

	for pm, paths := range flags.SubProjects {
		switch pm {
		case api.PackageManagerMaven:
			if !foundation.FileExists("/root/.m2/settings.xml") {
				var credential api.APITokenCredentials
				credential, err = s.credentialsClient.GetCredential(ctx)
				if err != nil {
					return
				}

				if credential.AdditionalProperties.MavenMirrorUrl != "" && credential.AdditionalProperties.MavenUsername != "" && credential.AdditionalProperties.MavenPassword != "" {
					log.Info().Msg("Initializing maven settings...")
					foundation.RunCommand(ctx, "mkdir -p /root/.m2")

					log.Info().Msgf("Generating settings.xml with url %v, username %v, password %v", credential.AdditionalProperties.MavenMirrorUrl, credential.AdditionalProperties.MavenUsername, credential.AdditionalProperties.MavenPassword)
					settingsTemplate, err := template.New("settings.xml").ParseFiles("/settings.xml")
					if err != nil {
						log.Fatal().Err(err).Msg("Failed parsing settings.xml")
					}

					data := struct {
						MirrorUrl string
						Username  string
						Password  string
					}{credential.AdditionalProperties.MavenMirrorUrl, credential.AdditionalProperties.MavenUsername, credential.AdditionalProperties.MavenPassword}

					var renderedSettings bytes.Buffer
					err = settingsTemplate.Execute(&renderedSettings, data)
					if err != nil {
						log.Fatal().Err(err).Msg("Failed rendering settings.xml")
					}

					err = ioutil.WriteFile("/root/.m2/settings.xml", renderedSettings.Bytes(), 0644)
					if err != nil {
						log.Fatal().Err(err).Msg("Failed writing settings.xml")
					}
				}
			}

		case api.PackageManagerNuget:
			for _, path := range paths {
				innerErr := foundation.RunCommandExtended(ctx, "dotnet restore --packages .nuget/packages %v", path)
				if innerErr != nil {
					if pm.IgnoreErrors() {
						log.Warn().Err(innerErr).Msgf("Failed preparing %v application, ignoring until package manager is fully supported...", pm)
					} else {
						return innerErr
					}
				}
			}

		case api.PackageManagerPip:
			for _, path := range paths {
				innerErr := foundation.RunCommandExtended(ctx, "pip install -r %v", path)
				if innerErr != nil {
					if pm.IgnoreErrors() {
						log.Warn().Err(innerErr).Msgf("Failed preparing %v application, ignoring until package manager is fully supported...", pm)
					} else {
						return innerErr
					}
				}
			}
		}
	}

	err = s.snykcliClient.Auth(ctx)
	if err != nil {
		return
	}

	err = s.snykcliClient.Test(ctx, flags)
	if err != nil {
		return
	}

	err = s.snykcliClient.Monitor(ctx, flags)
	if err != nil {
		return
	}

	return nil
}
