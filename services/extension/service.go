package extension

import (
	"context"
	"errors"
	"strconv"

	"github.com/estafette/estafette-extension-snyk/clients/snykapi"
)

var (
	ErrStatusTooLow = errors.New("Returned status is too low")
)

type Service interface {
	Run(ctx context.Context, repoSource, repoOwner, repoName, repoBranch string, minimumValueToSucceed int) (err error)
}

func NewService(snykapiClient snykapi.Client) Service {
	return &service{
		snykapiClient: snykapiClient,
	}
}

type service struct {
	snykapiClient snykapi.Client
}

func (s *service) Run(ctx context.Context, repoSource, repoOwner, repoName, repoBranch string, minimumValueToSucceed int) (err error) {

	// get status from snyk
	status, err := s.snykapiClient.GetStatus(ctx, repoSource, repoOwner, repoName)
	if err != nil {
		return
	}

	// check if status is higher than minimumValueToSucceed
	statusInt, err := strconv.Atoi(status)
	if err != nil {
		return
	}

	if statusInt < minimumValueToSucceed {
		return ErrStatusTooLow
	}

	return nil
}
