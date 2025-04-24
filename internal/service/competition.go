package service

import (
	"context"

	"github.com/NiskuT/cross-api/internal/config"
	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	"github.com/NiskuT/cross-api/internal/domain/repository"
)

type CompetitionService struct {
	competitionRepo repository.CompetitionRepository
	scaleRepo       repository.ScaleRepository
	liverankingRepo repository.LiverankingRepository
	participantRepo repository.ParticipantRepository
	cfg             *config.Config
}

type CompetitionServiceConfiguration func(c *CompetitionService) error

func NewCompetitionService(cfgs ...CompetitionServiceConfiguration) *CompetitionService {
	impl := new(CompetitionService)

	for _, cfg := range cfgs {
		if err := cfg(impl); err != nil {
			panic(err)
		}
	}

	return impl
}

func CompetitionConfWithCompetitionRepo(repo repository.CompetitionRepository) CompetitionServiceConfiguration {
	return func(c *CompetitionService) error {
		c.competitionRepo = repo
		return nil
	}
}

func CompetitionConfWithConfig(cfg *config.Config) CompetitionServiceConfiguration {
	return func(c *CompetitionService) error {
		c.cfg = cfg
		return nil
	}
}

func CompetitionConfWithScaleRepo(repo repository.ScaleRepository) CompetitionServiceConfiguration {
	return func(c *CompetitionService) error {
		c.scaleRepo = repo
		return nil
	}
}

func CompetitionConfWithLiverankingRepo(repo repository.LiverankingRepository) CompetitionServiceConfiguration {
	return func(c *CompetitionService) error {
		c.liverankingRepo = repo
		return nil
	}
}

func CompetitionConfWithParticipantRepo(repo repository.ParticipantRepository) CompetitionServiceConfiguration {
	return func(c *CompetitionService) error {
		c.participantRepo = repo
		return nil
	}
}

func (s *CompetitionService) CreateCompetition(ctx context.Context, competition *aggregate.Competition) (int32, error) {
	id, err := s.competitionRepo.CreateCompetition(ctx, competition)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *CompetitionService) AddZone(ctx context.Context, competitionID int32, zone *aggregate.Scale) error {
	// check if competition exists
	_, err := s.competitionRepo.GetCompetition(ctx, competitionID)
	if err != nil {
		return err
	}

	return s.scaleRepo.CreateScale(ctx, zone)
}
