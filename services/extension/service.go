package extension

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"io/ioutil"
	"os"
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
	AugmentFlags(ctx context.Context, flags api.SnykFlags) (augmentedFlags api.SnykFlags, err error)
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

func (s *service) AugmentFlags(ctx context.Context, flags api.SnykFlags) (augmentedFlags api.SnykFlags, err error) {

	log.Info().Msg("Detecting language...")

	augmentedFlags = flags
	augmentedFlags.Language, err = s.detectLanguage(ctx)
	if err != nil {
		return
	}

	switch augmentedFlags.Language {
	case api.LanguageGolang:
		log.Info().Msg("Detected golang application")
	case api.LanguageNode:
		log.Info().Msg("Detected node application")
	case api.LanguageMaven:
		log.Info().Msg("Detected maven application")
	case api.LanguageDotnet:
		log.Info().Msg("Detected dotnet application")
		if augmentedFlags.File == "" {
			// set file flag if 1 solution file is found
			matches, innerErr := s.findFileMatches(".", ".+\\.sln")
			if innerErr != nil {
				return augmentedFlags, innerErr
			}
			if len(matches) == 1 {
				augmentedFlags.File = matches[0]
				log.Info().Msgf("Autodetected file %v and using it as 'file' parameter", augmentedFlags.File)
			}
		}
	case api.LanguagePython:
		log.Info().Msg("Detected python application")
	}

	return
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
	matches, err := s.findFileMatches(".", ".+\\.sln")
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
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
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

			if flags.MavenUpdateParent && flags.BuildVersionMajor != "" && flags.BuildVersionMinor != "" {
				log.Info().Msg("Updating parent pom to latest patch version...")
				foundation.RunCommand(ctx, "mvn -DparentVersion=[0.0.0,%v.%v.9999] versions:update-parent", flags.BuildVersionMajor, flags.BuildVersionMinor)
			}
		}

	case api.LanguageDotnet:
		foundation.RunCommand(ctx, "dotnet restore --packages .nuget/packages")
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
