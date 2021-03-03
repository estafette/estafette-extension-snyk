package extension

import (
	"context"
	"errors"

	"github.com/estafette/estafette-extension-snyk/clients/snykapi"
	"github.com/rs/zerolog/log"
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

	log.Info().Msg("Fetching organizations for user")

	organizations, err := s.snykapiClient.GetOrganizations(ctx)
	if err != nil {
		return
	}

	log.Info().Interface("organizations", organizations).Msgf("Retrieved %v organizations for user", len(organizations))

	for _, org := range organizations {
		log.Info().Msgf("Fetching projects for org %v", org.Name)

		projects, innerErr := s.snykapiClient.GetProjects(ctx, org)
		if innerErr != nil {
			return innerErr
		}

		if len(projects) > 100 {
			log.Info().Msgf("Retrieved %v projects for org %v", len(projects), org.Name)
		} else {
			log.Info().Interface("projects", projects).Msgf("Retrieved %v projects for org %v", len(projects), org.Name)
		}

		for _, p := range projects {
			log.Debug().Msgf("%v", p.Name)
		}
	}

	// // get status from snyk
	// status, err := s.snykapiClient.GetStatus(ctx, repoSource, repoOwner, repoName)
	// if err != nil {
	// 	return
	// }

	// // check if status is higher than minimumValueToSucceed
	// statusInt, err := strconv.Atoi(status)
	// if err != nil {
	// 	return
	// }

	// if statusInt < minimumValueToSucceed {
	// 	return ErrStatusTooLow
	// }

	return nil
}
