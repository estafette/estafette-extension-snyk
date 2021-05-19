package extension

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/estafette/estafette-extension-snyk/api"
	"github.com/estafette/estafette-extension-snyk/clients/snykcli"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/rs/zerolog/log"
)

var (
	ErrStatusTooLow = errors.New("Returned status is too low")
)

type Service interface {
	AugmentFlags(ctx context.Context, flags api.SnykFlags, toolVersion api.ToolVersions) (api.SnykFlags, error)
	Run(ctx context.Context, flags api.SnykFlags) (err error)
}

func NewService(snykcliClient snykcli.Client) Service {
	return &service{
		snykcliClient: snykcliClient,
	}
}

type service struct {
	snykcliClient snykcli.Client
}

func (s *service) AugmentFlags(ctx context.Context, flags api.SnykFlags, toolVersion api.ToolVersions) (api.SnykFlags, error) {

	log.Info().Msg("Detecting language...")

	var err error
	flags.Language, err = s.detectLanguage(ctx)
	if err != nil {
		return flags, err
	}

	switch flags.Language {
	case api.LanguageGolang:
		log.Info().Str("go", toolVersion.Go).Msg("Detected golang application")
	case api.LanguageNode:
		log.Info().Str("node", toolVersion.Node).Str("npm", toolVersion.Npm).Msg("Detected node application")
	case api.LanguageMaven:
		log.Info().Str("java", toolVersion.Java).Str("maven", toolVersion.Maven).Msg("Detected maven application")
	case api.LanguageDotnet:
		log.Info().Str("dotnet", toolVersion.Dotnet).Msg("Detected dotnet application")
		if flags.File == "" {
			// set file flag if 1 solution file is found
			matches, innerErr := s.findFileMatches(".", "*.sln")
			if innerErr != nil {
				return flags, innerErr
			}
			if len(matches) == 1 {
				flags.File = matches[0]
				log.Info().Msgf("Autodetected file %v and using it as 'file' parameter", flags.File)
			}
		}
	case api.LanguagePython:
		log.Info().Str("python", toolVersion.Python).Str("pip", toolVersion.Pip).Msg("Detected python application")
		// 	if flags.File == "" {
		// 		flags.File = "requirements.txt"
		// 		log.Info().Msgf("Autodetected file %v and using it as 'file' parameter", flags.File)
		// 	}
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
		if flags.MavenMirrorUrl != "" && flags.MavenUsername != "" && flags.MavenPassword != "" {
			log.Info().Msg("Initializing maven settings...")
			foundation.RunCommand(ctx, "mkdir -p /root/.m2")

			log.Info().Msgf("Generating settings.xml with url %v, username %v, password %v", flags.MavenMirrorUrl, flags.MavenUsername, flags.MavenPassword)
			settingsTemplate, err := template.New("settings.xml").ParseFiles("/settings.xml")
			if err != nil {
				log.Fatal().Err(err).Msg("Failed parsing settings.xml")
			}

			data := struct {
				MirrorUrl string
				Username  string
				Password  string
			}{flags.MavenMirrorUrl, flags.MavenUsername, flags.MavenPassword}

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
		foundation.RunCommand(ctx, "dotnet restore --packages .nuget/packages")

	case api.LanguagePython:
		foundation.RunCommand(ctx, "pip install -r requirements.txt")
	}

	err = s.snykcliClient.Auth(ctx)
	if err != nil {
		return
	}

	err = s.snykcliClient.Test(ctx, flags)
	if err != nil {
		return
	}

	return nil
}
