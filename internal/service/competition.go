package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/NiskuT/cross-api/internal/config"
	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	"github.com/NiskuT/cross-api/internal/domain/repository"
	"github.com/xuri/excelize/v2"
)

// Define error constants
var (
	ErrInvalidExcelFormat = errors.New("invalid excel format: expected columns for last name, first name, and dossard number")
	ErrParticipantExists  = errors.New("participant with this dossard number already exists in the competition")
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

// Helper function to check if error is because participant already exists
func isParticipantAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

// AddParticipants creates multiple participants from an Excel file for a competition
func (s *CompetitionService) AddParticipants(ctx context.Context, competitionID int32, category string, excelFile io.Reader) error {
	// Check if competition exists
	_, err := s.competitionRepo.GetCompetition(ctx, competitionID)
	if err != nil {
		return err
	}

	// Open Excel file
	xlsx, err := excelize.OpenReader(excelFile)
	if err != nil {
		return fmt.Errorf("failed to open excel file: %w", err)
	}
	defer xlsx.Close()

	// Get active sheet
	sheetName := xlsx.GetSheetName(0)

	// Read rows from Excel
	rows, err := xlsx.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to read rows from excel: %w", err)
	}

	if len(rows) < 2 { // At least header row and one data row required
		return ErrInvalidExcelFormat
	}

	// Process participants
	for i, row := range rows {
		// Skip header row
		if i == 0 {
			continue
		}

		// Excel should have at least 3 columns: first name, last name, dossard number
		if len(row) < 3 {
			return ErrInvalidExcelFormat
		}

		lastName := row[0]
		firstName := row[1]

		// Parse dossard number
		dossardStr := row[2]
		dossard, err := strconv.ParseInt(dossardStr, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid dossard number on row %d: %w", i+1, err)
		}

		// Create participant
		participant := aggregate.NewParticipant()
		participant.SetCompetitionID(competitionID)
		participant.SetDossardNumber(int32(dossard))
		participant.SetFirstName(firstName)
		participant.SetLastName(lastName)
		participant.SetCategory(category)

		// Add participant to database
		err = s.participantRepo.CreateParticipant(ctx, participant)
		if err != nil {
			// Check for duplicate participant error, continue with other participants if possible
			if isParticipantAlreadyExistsError(err) {
				// Log the error or handle it as needed
				continue
			}
			return fmt.Errorf("failed to create participant (row %d): %w", i+1, err)
		}
	}

	return nil
}
