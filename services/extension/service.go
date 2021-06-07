package extension

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/estafette/estafette-extension-snyk/api"
	"github.com/estafette/estafette-extension-snyk/clients/credentials"
	"github.com/estafette/estafette-extension-snyk/clients/snykcli"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/rs/zerolog/log"
)

var (
	ErrStatusTooLow = errors.New("Returned status is too low")
)

type Service interface {
	AugmentFlags(ctx context.Context, flags api.SnykFlags) (api.SnykFlags, error)
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

func (s *service) AugmentFlags(ctx context.Context, flags api.SnykFlags) (api.SnykFlags, error) {

	log.Info().Msg("Detecting language...")

	var err error
	flags.Language, err = s.detectLanguage(ctx)
	if err != nil {
		return flags, err
	}

	log.Info().Msgf("Detected %v application", flags.Language)

	switch flags.Language {
	case api.LanguageDotnet:
		// if flags.File == "" {
		// 	// set file flag if 1 solution file is found
		// 	matches, innerErr := s.findFileMatches(".", "*.sln")
		// 	if innerErr != nil {
		// 		return flags, innerErr
		// 	}
		// 	if len(matches) == 1 {
		// 		flags.File = matches[0]
		// 		log.Info().Msgf("Autodetected file %v and using it as 'file' parameter", flags.File)
		// 	}
		// }
	}

	return flags, nil
}

func (s *service) detectLanguage(ctx context.Context) (api.Language, error) {

	// go.mod => golang
	if foundation.FileExists("go.mod") {
		return api.LanguageGolang, nil
	}

	// package.json => node
	if foundation.FileExists("package.json") {
		return api.LanguageNode, nil
	}

	// pom.xml => maven
	if foundation.FileExists("pom.xml") {
		return api.LanguageMaven, nil
	}

	// *.sln => dotnet
	matches, err := s.findFileMatches(".", "*.sln")
	if err != nil {
		return api.LanguageUnknown, err
	}
	if len(matches) > 0 {
		return api.LanguageDotnet, nil
	}

	// requirements.txt => python
	if foundation.FileExists("requirements.txt") {
		return api.LanguagePython, nil
	}

	// Dockerfile => docker
	if foundation.FileExists("Dockerfile") {
		return api.LanguageDocker, nil
	}

	return api.LanguageUnknown, nil
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

	// run language specific actions
	switch flags.Language {
	case api.LanguageUnknown:
		log.Info().Msg("Could not find supported language, exiting...")
		return

	case api.LanguageMaven:
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

	case api.LanguageDotnet:
		innerErr := foundation.RunCommandExtended(ctx, "dotnet restore --packages .nuget/packages")
		if innerErr != nil {
			if flags.Language.IgnoreErrors() {
				log.Warn().Err(innerErr).Msgf("Failed preparing %v application, ignoring until language is fully supported...", flags.Language)
			} else {
				return innerErr
			}
		}

	case api.LanguagePython:
		innerErr := foundation.RunCommandExtended(ctx, "pip install -r requirements.txt")
		if innerErr != nil {
			if flags.Language.IgnoreErrors() {
				log.Warn().Err(innerErr).Msgf("Failed preparing %v application, ignoring until language is fully supported...", flags.Language)
			} else {
				return innerErr
			}
		}
	}

	err = s.snykcliClient.Auth(ctx)
	if err != nil {
		return
	}

	err = s.snykcliClient.Test(ctx, flags)
	if err != nil {
		if flags.Language.IgnoreErrors() {
			log.Warn().Err(err).Msgf("Failed testing %v application, ignoring until language is fully supported...", flags.Language)
		} else {
			return
		}
	}

	return nil
}
