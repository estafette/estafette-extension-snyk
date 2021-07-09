package extension

import (
	"bytes"
	"context"
	"errors"
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

	// do some prep work for certain package managers
	err = s.prepare(ctx, flags)
	if err != nil {
		return
	}

	err = s.snykcliClient.Auth(ctx)
	if err != nil {
		return
	}

	err = s.snykcliClient.Monitor(ctx, flags)
	if err != nil {
		return
	}

	err = s.snykcliClient.Test(ctx, flags)
	if err != nil {
		return
	}

	return nil
}

func (s *service) prepare(ctx context.Context, flags api.SnykFlags) (err error) {
	// https://github.com/snyk/snyk/blob/97808254747dd7db8c7033e76dafcd46a7976d54/src/lib/package-managers.ts#L19-L49
	// https://github.com/snyk/snyk/blob/97808254747dd7db8c7033e76dafcd46a7976d54/src/lib/detect.ts#L68-L100

	err = s.prepareMaven(ctx, flags)
	if err != nil {
		return
	}

	err = s.prepareNuget(ctx, flags)
	if err != nil {
		return
	}

	err = s.preparePip(ctx, flags)
	if err != nil {
		return
	}

	return nil
}

func (s *service) prepareMaven(ctx context.Context, flags api.SnykFlags) (err error) {
	matches, err := s.findFileMatches(".", "pom.xml")
	if err != nil {
		return
	}

	if len(matches) == 0 {
		return
	}

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

	return nil
}

func (s *service) prepareNuget(ctx context.Context, flags api.SnykFlags) (err error) {
	matches, err := s.findFileMatches(".", "*.sln|project.assets.json|packages.config|project.json")
	if err != nil {
		return
	}

	if len(matches) == 0 {
		return
	}

	for _, path := range matches {
		innerErr := foundation.RunCommandExtended(ctx, "dotnet restore --packages .nuget/packages %v", path)
		if innerErr != nil {
			return innerErr
		}
	}

	return nil
}

func (s *service) preparePip(ctx context.Context, flags api.SnykFlags) (err error) {
	matches, err := s.findFileMatches(".", "requirements.txt|Pipfile|setup.py")
	if err != nil {
		return
	}

	if len(matches) == 0 {
		return
	}

	for _, path := range matches {
		innerErr := foundation.RunCommandExtended(ctx, "pip install -r %v", path)
		if innerErr != nil {
			log.Warn().Err(innerErr).Msgf("Failed preparing python application, ignoring until pip package manager is fully supported...")
		}
	}

	return nil
}
