package extension

import (
	"context"
	"errors"

	"github.com/estafette/estafette-extension-snyk/clients/snykcli"
)

var (
	ErrStatusTooLow = errors.New("Returned status is too low")
)

type Service interface {
	Run(ctx context.Context, severityThreshold, failOn string) (err error)
}

func NewService(snykcliClient snykcli.Client) Service {
	return &service{
		snykcliClient: snykcliClient,
	}
}

type service struct {
	snykcliClient snykcli.Client
}

func (s *service) Run(ctx context.Context, severityThreshold, failOn string) (err error) {

	err = s.snykcliClient.Auth(ctx)
	if err != nil {
		return
	}

	err = s.snykcliClient.Test(ctx, severityThreshold, failOn)
	if err != nil {
		return
	}

	return nil
}
