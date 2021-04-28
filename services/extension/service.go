package extension

import (
	"context"
	"errors"

	"github.com/estafette/estafette-extension-snyk/api"
	"github.com/estafette/estafette-extension-snyk/clients/snykcli"
)

var (
	ErrStatusTooLow = errors.New("Returned status is too low")
)

type Service interface {
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

func (s *service) Run(ctx context.Context, flags api.SnykFlags) (err error) {

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
