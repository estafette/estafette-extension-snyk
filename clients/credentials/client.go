package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"runtime"

	"github.com/estafette/estafette-extension-snyk/api"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/rs/zerolog/log"
)

var (
	ErrInjectedCredentialsFileMissing = errors.New("The injected credentials file is missing")
	ErrEmptyToken                     = errors.New("Token is empty")
)

type Client interface {
	GetToken(ctx context.Context, snykAPITokenPath string) (token string, err error)
}

// NewClient returns a new credentials.Client
func NewClient() Client {
	return &client{}
}

type client struct {
}

func (c *client) GetToken(ctx context.Context, snykAPITokenPath string) (token string, err error) {

	// get api token from injected credentials
	if runtime.GOOS == "windows" {
		snykAPITokenPath = "C:" + snykAPITokenPath
	}
	if !foundation.FileExists(snykAPITokenPath) {
		return token, ErrInjectedCredentialsFileMissing
	}

	log.Info().Msgf("Reading credentials from file at path %v...", snykAPITokenPath)
	credentialsFileContent, err := ioutil.ReadFile(snykAPITokenPath)
	if err != nil {
		return
	}

	log.Info().Msg("Unmarshalling injected snyk api token credentials")
	var credentials []api.APITokenCredentials
	err = json.Unmarshal([]byte(credentialsFileContent), &credentials)
	if err != nil {
		return
	}
	if len(credentials) == 0 {
		return token, ErrEmptyToken
	}

	token = credentials[0].AdditionalProperties.Token

	if token == "" {
		return token, ErrEmptyToken
	}

	return
}
