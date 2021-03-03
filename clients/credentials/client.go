package credentials

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/rs/zerolog/log"
)

var (
	ErrEmptyToken = errors.New("Token is empty")
)

type Client interface {
	GetToken(ctx context.Context, snykAPITokenJSON string) (token string, err error)
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

func (c *client) GetToken(ctx context.Context, snykAPITokenJSON string) (token string, err error) {

	if snykAPITokenJSON == "" {
		return token, ErrEmptyToken
	}

	log.Info().Msg("Unmarshalling injected snyk api token credentials")
	var credentials []APITokenCredentials
	err = json.Unmarshal([]byte(snykAPITokenJSON), &credentials)
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
