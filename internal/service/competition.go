package service

import (
	"context"
	"encoding/csv"
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
	ErrInvalidFileFormat = errors.New("invalid file format: expected CSV or Excel file with columns for dossard number, category, last name, first name, and gender (H/F)")
	ErrParticipantExists = errors.New("participant with this dossard number already exists in the competition")
	ErrCategoryAndGender = errors.New("category and gender cannot be empty")
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

// Helper function to check if error is because participant already exists
func isParticipantAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

// AddParticipants creates multiple participants from a CSV or Excel file for a competition
func (s *CompetitionService) AddParticipants(ctx context.Context, competitionID int32, file io.Reader, filename string) error {
	// Check if competition exists
	_, err := s.competitionRepo.GetCompetition(ctx, competitionID)
	if err != nil {
		return err
	}

	// Determine file type based on extension
	isCSV := strings.HasSuffix(strings.ToLower(filename), ".csv")
	isExcel := strings.HasSuffix(strings.ToLower(filename), ".xlsx") || strings.HasSuffix(strings.ToLower(filename), ".xls")

	if !isCSV && !isExcel {
		return fmt.Errorf("unsupported file format: %s. Only CSV and Excel files are supported", filename)
	}

	var rows [][]string

	if isCSV {
		// Handle CSV file
		rows, err = s.readCSVFile(file)
		if err != nil {
			return fmt.Errorf("failed to read CSV file: %w", err)
		}
	} else {
		// Handle Excel file
		rows, err = s.readExcelFile(file)
		if err != nil {
			return fmt.Errorf("failed to read Excel file: %w", err)
		}
	}

	if len(rows) < 2 { // At least header row and one data row required
		return ErrInvalidFileFormat
	}

	// Process participants
	for i, row := range rows {
		// Skip header row
		if i == 0 {
			continue
		}

		// File should have at least 5 columns: dossard number, category, last name, first name, gender
		if len(row) < 5 {
			return fmt.Errorf("invalid format on row %d: expected 5 columns (dossard number, category, last name, first name, gender)", i+1)
		}

		// Parse dossard number (first column)
		dossardStr := strings.TrimSpace(row[0])
		dossard, err := strconv.ParseInt(dossardStr, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid dossard number on row %d: %w", i+1, err)
		}

		// Get category from file (second column)
		categoryFromFile := strings.TrimSpace(row[1])
		// Get last name (third column)
		lastName := strings.TrimSpace(row[2])
		// Get first name (fourth column)
		firstName := strings.TrimSpace(row[3])
		// Get gender (fifth column)
		gender := strings.TrimSpace(strings.ToUpper(row[4]))

		// Validate gender
		if gender != "H" && gender != "F" {
			return fmt.Errorf("invalid gender on row %d: expected 'H' or 'F', got '%s'", i+1, gender)
		}

		// Create participant
		participant := aggregate.NewParticipant()
		participant.SetCompetitionID(competitionID)
		participant.SetDossardNumber(int32(dossard))
		participant.SetFirstName(firstName)
		participant.SetLastName(lastName)
		participant.SetCategory(categoryFromFile)
		participant.SetGender(gender)

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

// readCSVFile reads data from a CSV file
func (s *CompetitionService) readCSVFile(file io.Reader) ([][]string, error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// readExcelFile reads data from an Excel file
func (s *CompetitionService) readExcelFile(file io.Reader) ([][]string, error) {
	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer xlsx.Close()

	// Get active sheet
	sheetName := xlsx.GetSheetName(0)

	// Read rows from Excel
	rows, err := xlsx.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *CompetitionService) ListCompetitions(ctx context.Context) ([]*aggregate.Competition, error) {
	competitions, err := s.competitionRepo.ListCompetitions(ctx)
	if err != nil {
		return nil, err
	}

	return competitions, nil
}

// CreateParticipant creates a single participant for a competition
func (s *CompetitionService) CreateParticipant(ctx context.Context, participant *aggregate.Participant) error {
	// Check if competition exists
	_, err := s.competitionRepo.GetCompetition(ctx, participant.GetCompetitionID())
	if err != nil {
		return err
	}

	// Create participant
	return s.participantRepo.CreateParticipant(ctx, participant)
}

// GetParticipant retrieves a participant by competition ID and dossard number
func (s *CompetitionService) GetParticipant(ctx context.Context, competitionID int32, dossardNumber int32) (*aggregate.Participant, error) {
	// Get participant from repository
	participant, err := s.participantRepo.GetParticipant(ctx, competitionID, dossardNumber)
	if err != nil {
		return nil, err
	}

	return participant, nil
}

// ListParticipantsByCategory retrieves all participants for a competition by category
func (s *CompetitionService) ListParticipantsByCategory(ctx context.Context, competitionID int32, category string) ([]*aggregate.Participant, error) {
	// Verify the competition exists
	_, err := s.competitionRepo.GetCompetition(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	// Get participants from repository
	return s.participantRepo.ListParticipantsByCategory(ctx, competitionID, category)
}

// ListZones lists all zones for a competition
func (s *CompetitionService) ListZones(ctx context.Context, competitionID int32) ([]aggregate.ZoneInfo, error) {
	// Verify the competition exists
	_, err := s.competitionRepo.GetCompetition(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	// Get zones from repository
	return s.scaleRepo.ListZones(ctx, competitionID)
}

func (s *CompetitionService) GetScale(ctx context.Context, competitionID int32, category string, zone string) (*aggregate.Scale, error) {
	// Verify the competition exists
	_, err := s.competitionRepo.GetCompetition(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	return s.scaleRepo.GetScale(ctx, competitionID, category, zone)
}

func (s *CompetitionService) AddScale(ctx context.Context, competitionID int32, scale *aggregate.Scale) error {
	// check if competition exists
	_, err := s.competitionRepo.GetCompetition(ctx, competitionID)
	if err != nil {
		return err
	}

	return s.scaleRepo.CreateScale(ctx, scale)
}

func (s *CompetitionService) UpdateScale(ctx context.Context, competitionID int32, scale *aggregate.Scale) error {
	// check if scale exists
	_, err := s.scaleRepo.GetScale(ctx, competitionID, scale.GetCategory(), scale.GetZone())
	if err != nil {
		return err
	}

	return s.scaleRepo.UpdateScale(ctx, scale)
}

func (s *CompetitionService) DeleteScale(ctx context.Context, competitionID int32, category string, zone string) error {
	// check if the scale exists
	_, err := s.scaleRepo.GetScale(ctx, competitionID, category, zone)
	if err != nil {
		return err
	}

	return s.scaleRepo.DeleteScale(ctx, competitionID, category, zone)
}

func (s *CompetitionService) GetLiveranking(ctx context.Context, competitionID int32, category, gender string, pageNumber, pageSize int32) ([]*aggregate.Liveranking, int32, error) {
	// check if competition exists
	_, err := s.competitionRepo.GetCompetition(ctx, competitionID)
	if err != nil {
		return nil, 0, err
	}

	if category == "" && gender == "" {
		return nil, 0, ErrCategoryAndGender
	}

	return s.liverankingRepo.ListLiverankingByCategoryAndGender(ctx, competitionID, category, gender, pageNumber, pageSize)
}
