package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"runtime"

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

// APITokenCredentials represents the credentials of type bitbucket-api-tokene as defined in the server config and passed to this trusted image
type APITokenCredentials struct {
	Name                 string                                  `json:"name,omitempty"`
	Type                 string                                  `json:"type,omitempty"`
	AdditionalProperties APITokenCredentialsAdditionalProperties `json:"additionalProperties,omitempty"`
}

// APITokenCredentialsAdditionalProperties contains the non standard fields for this type of credentials
type APITokenCredentialsAdditionalProperties struct {
	Token string `json:"token,omitempty"`
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
	var credentials []APITokenCredentials
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
