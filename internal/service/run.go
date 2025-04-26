package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/NiskuT/cross-api/internal/config"
	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	"github.com/NiskuT/cross-api/internal/domain/repository"
	"github.com/NiskuT/cross-api/internal/domain/service"
)

// Define error constants
var (
	ErrInvalidRunData = errors.New("invalid run data")
	ErrScaleNotFound  = errors.New("scale not found for this zone and category")
)

// RunService implements the RunService interface
type RunService struct {
	runRepo         repository.RunRepository
	participantRepo repository.ParticipantRepository
	liverankingRepo repository.LiverankingRepository
	scaleRepo       repository.ScaleRepository
	cfg             *config.Config
}

// RunServiceConfiguration is a function that configures a RunService
type RunServiceConfiguration func(r *RunService) error

// NewRunService creates a new RunService
func NewRunService(cfgs ...RunServiceConfiguration) service.RunService {
	impl := new(RunService)

	for _, cfg := range cfgs {
		if err := cfg(impl); err != nil {
			panic(err)
		}
	}

	return impl
}

// RunConfWithRunRepo configures the RunService with a RunRepository
func RunConfWithRunRepo(repo repository.RunRepository) RunServiceConfiguration {
	return func(r *RunService) error {
		r.runRepo = repo
		return nil
	}
}

// RunConfWithParticipantRepo configures the RunService with a ParticipantRepository
func RunConfWithParticipantRepo(repo repository.ParticipantRepository) RunServiceConfiguration {
	return func(r *RunService) error {
		r.participantRepo = repo
		return nil
	}
}

// RunConfWithLiverankingRepo configures the RunService with a LiverankingRepository
func RunConfWithLiverankingRepo(repo repository.LiverankingRepository) RunServiceConfiguration {
	return func(r *RunService) error {
		r.liverankingRepo = repo
		return nil
	}
}

// RunConfWithScaleRepo configures the RunService with a ScaleRepository
func RunConfWithScaleRepo(repo repository.ScaleRepository) RunServiceConfiguration {
	return func(r *RunService) error {
		r.scaleRepo = repo
		return nil
	}
}

// RunConfWithConfig configures the RunService with a Config
func RunConfWithConfig(cfg *config.Config) RunServiceConfiguration {
	return func(r *RunService) error {
		r.cfg = cfg
		return nil
	}
}

// CreateRun creates a new run and updates the liveranking
func (s *RunService) CreateRun(ctx context.Context, run *aggregate.Run) error {
	if run.GetCompetitionID() <= 0 || run.GetDossard() <= 0 || run.GetZone() == "" {
		return ErrInvalidRunData
	}

	// Get the participant to retrieve the category
	participant, err := s.participantRepo.GetParticipant(ctx, run.GetCompetitionID(), run.GetDossard())
	if err != nil {
		return fmt.Errorf("participant not found: %w", err)
	}

	// Get the scale for the category and zone
	scale, err := s.scaleRepo.GetScale(ctx, run.GetCompetitionID(), participant.GetCategory(), run.GetZone())
	if err != nil {
		return fmt.Errorf("failed to get scale: %w", err)
	}

	// Create the run
	err = s.runRepo.CreateRun(ctx, run)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	// Calculate points based on doors passed and scale
	totalPoints := int32(0)
	if run.GetDoor1() {
		totalPoints += scale.GetPointsDoor1()
	}
	if run.GetDoor2() {
		totalPoints += scale.GetPointsDoor2()
	}
	if run.GetDoor3() {
		totalPoints += scale.GetPointsDoor3()
	}
	if run.GetDoor4() {
		totalPoints += scale.GetPointsDoor4()
	}
	if run.GetDoor5() {
		totalPoints += scale.GetPointsDoor5()
	}
	if run.GetDoor6() {
		totalPoints += scale.GetPointsDoor6()
	}

	// Create or update liveranking entry
	liveranking := aggregate.NewLiveranking()
	liveranking.SetCompetitionID(run.GetCompetitionID())
	liveranking.SetDossard(run.GetDossard())
	liveranking.SetFirstName(participant.GetFirstName())
	liveranking.SetLastName(participant.GetLastName())
	liveranking.SetCategory(participant.GetCategory())
	liveranking.SetTotalPoints(totalPoints)
	liveranking.SetPenality(run.GetPenality())
	liveranking.SetChronoSec(run.GetChronoSec())

	// Update the liveranking
	err = s.liverankingRepo.UpsertLiveranking(ctx, liveranking)
	if err != nil {
		return fmt.Errorf("failed to update liveranking: %w", err)
	}

	return nil
}

// GetRun retrieves a run by its identifiers
func (s *RunService) GetRun(ctx context.Context, competitionID, runNumber, dossard int32) (*aggregate.Run, error) {
	return s.runRepo.GetRun(ctx, competitionID, runNumber, dossard)
}

// ListRuns lists all runs for a competition
func (s *RunService) ListRuns(ctx context.Context, competitionID int32) ([]*aggregate.Run, error) {
	return s.runRepo.ListRuns(ctx, competitionID)
}

// ListRunsByDossard lists all runs for a participant in a competition
func (s *RunService) ListRunsByDossard(ctx context.Context, competitionID int32, dossard int32) ([]*aggregate.Run, error) {
	return s.runRepo.ListRunsByDossard(ctx, competitionID, dossard)
}

// UpdateRun updates an existing run
func (s *RunService) UpdateRun(ctx context.Context, run *aggregate.Run) error {
	return s.runRepo.UpdateRun(ctx, run)
}

// DeleteRun deletes a run
func (s *RunService) DeleteRun(ctx context.Context, competitionID, runNumber, dossard int32) error {
	return s.runRepo.DeleteRun(ctx, competitionID, runNumber, dossard)
}
