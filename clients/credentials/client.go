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
	GetCredential(ctx context.Context) (credential api.APITokenCredentials, err error)
	GetToken(ctx context.Context) (token string, err error)
}

// NewClient returns a new credentials.Client
func NewClient(snykAPITokenPath string) Client {
	if runtime.GOOS == "windows" {
		snykAPITokenPath = "C:" + snykAPITokenPath
	}
	return &client{
		snykAPITokenPath: snykAPITokenPath,
	}
}

type client struct {
	snykAPITokenPath string
}

func (c *client) GetCredential(ctx context.Context) (credential api.APITokenCredentials, err error) {
	// get api token from injected credentials
	if !foundation.FileExists(c.snykAPITokenPath) {
		return credential, ErrInjectedCredentialsFileMissing
	}

	log.Info().Msgf("Reading credentials from file at path %v...", c.snykAPITokenPath)
	credentialsFileContent, err := ioutil.ReadFile(c.snykAPITokenPath)
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
		return credential, ErrEmptyToken
	}

	return credentials[0], nil
}

func (c *client) GetToken(ctx context.Context) (token string, err error) {

	credential, err := c.GetCredential(ctx)
	if err != nil {
		return
	}

	token = credential.AdditionalProperties.Token

	if token == "" {
		return token, ErrEmptyToken
	}

	return
}
