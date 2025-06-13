package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/NiskuT/cross-api/internal/config"
	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	"github.com/NiskuT/cross-api/internal/domain/repository"
	"github.com/NiskuT/cross-api/internal/utils"
	"github.com/xuri/excelize/v2"
)

// Define error constants
var (
	ErrInvalidFileFormat = errors.New("invalid file format: expected CSV or Excel file with columns for dossard number, category, last name, first name, gender (H/F), and club")
	ErrParticipantExists = errors.New("participant with this dossard number already exists in the competition")
	ErrCategoryAndGender = errors.New("category and gender cannot be empty")
)

type CompetitionService struct {
	competitionRepo repository.CompetitionRepository
	scaleRepo       repository.ScaleRepository
	liverankingRepo repository.LiverankingRepository
	participantRepo repository.ParticipantRepository
	runRepo         repository.RunRepository
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

func CompetitionConfWithRunRepo(repo repository.RunRepository) CompetitionServiceConfiguration {
	return func(c *CompetitionService) error {
		c.runRepo = repo
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
			return fmt.Errorf("invalid format on row %d: expected at least 5 columns (dossard number, category, last name, first name, gender, club)", i+1)
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
		// Get club (sixth column, optional)
		var club string
		if len(row) > 5 {
			club = strings.TrimSpace(row[5])
		}

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
		participant.SetClub(club)

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

func (s *CompetitionService) ExportCompetitionResults(ctx context.Context, competitionID int32) ([]byte, string, error) {
	// Get competition details for filename
	competition, err := s.competitionRepo.GetCompetition(ctx, competitionID)
	if err != nil {
		return nil, "", err
	}

	// Create filename from competition name
	filename := strings.ReplaceAll(competition.GetName(), " ", "_") + "_results.xlsx"

	// Get all participants for this competition
	participants, err := s.getAllParticipants(ctx, competitionID)
	if err != nil {
		return nil, "", err
	}

	// Get all runs for this competition
	runs, err := s.getAllRuns(ctx, competitionID)
	if err != nil {
		return nil, "", err
	}

	// Get all scales for this competition
	scales, err := s.getAllScales(ctx, competitionID)
	if err != nil {
		return nil, "", err
	}

	// Group participants by category and gender
	participantGroups := s.groupParticipantsByCategoryGender(participants)

	// Create Excel file
	excelData, err := s.generateExcelFile(ctx, competitionID, participantGroups, runs, scales)
	if err != nil {
		return nil, "", err
	}

	return excelData, filename, nil
}

// Helper method to get all participants for a competition
func (s *CompetitionService) getAllParticipants(ctx context.Context, competitionID int32) ([]*aggregate.Participant, error) {
	// Get all categories first
	zones, err := s.scaleRepo.ListZones(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	// Extract unique categories
	categorySet := make(map[string]bool)
	for _, zone := range zones {
		categorySet[zone.GetCategory()] = true
	}

	// Get participants for each category
	var allParticipants []*aggregate.Participant
	for category := range categorySet {
		participants, err := s.participantRepo.ListParticipantsByCategory(ctx, competitionID, category)
		if err != nil {
			return nil, err
		}
		allParticipants = append(allParticipants, participants...)
	}

	return allParticipants, nil
}

// Helper method to get all runs for a competition
func (s *CompetitionService) getAllRuns(ctx context.Context, competitionID int32) (map[string][]*aggregate.Run, error) {
	allRuns, err := s.runRepo.ListRuns(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	// Group runs by participant (competitionID_dossard)
	runsByParticipant := make(map[string][]*aggregate.Run)
	for _, run := range allRuns {
		key := fmt.Sprintf("%d_%d", run.GetCompetitionID(), run.GetDossard())
		runsByParticipant[key] = append(runsByParticipant[key], run)
	}

	return runsByParticipant, nil
}

// Helper method to get all scales for a competition
func (s *CompetitionService) getAllScales(ctx context.Context, competitionID int32) (map[string]*aggregate.Scale, error) {
	zones, err := s.scaleRepo.ListZones(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	scales := make(map[string]*aggregate.Scale)
	for _, zone := range zones {
		scale, err := s.scaleRepo.GetScale(ctx, competitionID, zone.GetCategory(), zone.GetZone())
		if err != nil {
			continue // Skip if scale not found
		}
		key := fmt.Sprintf("%s_%s", zone.GetCategory(), zone.GetZone())
		scales[key] = scale
	}

	return scales, nil
}

// Helper method to group participants by category and gender
func (s *CompetitionService) groupParticipantsByCategoryGender(participants []*aggregate.Participant) map[string][]*aggregate.Participant {
	groups := make(map[string][]*aggregate.Participant)

	for _, participant := range participants {
		key := fmt.Sprintf("%s_%s", participant.GetCategory(), participant.GetGender())
		groups[key] = append(groups[key], participant)
	}

	return groups
}

// Helper method to generate Excel file
func (s *CompetitionService) generateExcelFile(ctx context.Context,
	competitionID int32,
	participantGroups map[string][]*aggregate.Participant,
	runs map[string][]*aggregate.Run,
	scales map[string]*aggregate.Scale,
) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Remove default sheet
	f.DeleteSheet("Sheet1")

	sheetIndex := 0
	for groupKey, participants := range participantGroups {
		parts := strings.Split(groupKey, "_")
		if len(parts) != 2 {
			continue
		}
		category, gender := parts[0], parts[1]

		// Get zones for this category in lexical order
		zones, err := s.getZonesForCategory(ctx, competitionID, category)
		if err != nil {
			continue
		}

		// Create sheet for this category-gender combination
		sheetName := fmt.Sprintf("%s-%s", category, gender)
		if sheetIndex == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}

		// Generate sheet content
		err = s.generateSheetContent(f, sheetName, participants, zones, runs, scales, competitionID)
		if err != nil {
			continue
		}

		sheetIndex++
	}

	// Save to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// Helper method to get zones for a category in lexical order
func (s *CompetitionService) getZonesForCategory(ctx context.Context, competitionID int32, category string) ([]string, error) {
	allZones, err := s.scaleRepo.ListZones(ctx, competitionID)
	if err != nil {
		return nil, err
	}

	var zones []string
	for _, zone := range allZones {
		if zone.GetCategory() == category {
			zones = append(zones, zone.GetZone())
		}
	}

	// Sort zones lexically
	sort.Strings(zones)
	return zones, nil
}

// ParticipantResult represents the calculated results for a participant
type ParticipantResult struct {
	Participant  *aggregate.Participant
	ZoneResults  []ZoneResult
	TotalPoints  int32
	TotalPenalty int32
	TotalTime    int32
	HasError     bool
}

// ZoneResult represents the result for a specific zone
type ZoneResult struct {
	Points  int32
	Penalty int32
	Time    int32
	IsError bool
}

// Helper method to generate content for a sheet
func (s *CompetitionService) generateSheetContent(f *excelize.File, sheetName string, participants []*aggregate.Participant, zones []string, runs map[string][]*aggregate.Run, scales map[string]*aggregate.Scale, competitionID int32) error {
	// Create headers based on zone count
	headers := []string{"Position", "Dossard", "Nom", "Prénom", "Club"}

	// Determine expected runs per zone
	expectedRunsPerZone := 1
	if len(zones) == 2 {
		expectedRunsPerZone = 2
	}

	// Add zone headers
	if len(zones) == 2 {
		// 2 zones, 2 runs each: Zone1, Zone2, Zone1, Zone2
		for i := 0; i < 2; i++ {
			for _, zone := range zones {
				headers = append(headers, fmt.Sprintf("%s Points", zone))
				headers = append(headers, fmt.Sprintf("%s Penalités", zone))
				headers = append(headers, fmt.Sprintf("%s Temps", zone))
			}
		}
	} else {
		// 4 zones, 1 run each
		for _, zone := range zones {
			headers = append(headers, fmt.Sprintf("%s Points", zone))
			headers = append(headers, fmt.Sprintf("%s Penalités", zone))
			headers = append(headers, fmt.Sprintf("%s Temps", zone))
		}
	}

	headers = append(headers, "Total Points", "Total Penalités", "Total Temps", "Points Gagnés")

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
	}

	// Calculate results for each participant
	var results []ParticipantResult

	for _, participant := range participants {
		participantKey := fmt.Sprintf("%d_%d", competitionID, participant.GetDossardNumber())
		participantRuns := runs[participantKey]

		result := ParticipantResult{
			Participant: participant,
			ZoneResults: make([]ZoneResult, len(zones)*expectedRunsPerZone),
		}

		// Group runs by zone
		runsByZone := make(map[string][]*aggregate.Run)
		for _, run := range participantRuns {
			runsByZone[run.GetZone()] = append(runsByZone[run.GetZone()], run)
		}

		// Calculate results for each zone
		zoneIndex := 0
		for _, zone := range zones {
			zoneRuns := runsByZone[zone]

			// Check if we have the correct number of runs for this zone
			if len(zoneRuns) != expectedRunsPerZone {
				// Mark all runs for this zone as error
				for i := 0; i < expectedRunsPerZone; i++ {
					result.ZoneResults[zoneIndex] = ZoneResult{IsError: true}
					zoneIndex++
				}
				result.HasError = true
				continue
			}

			// Process each run for this zone
			for _, run := range zoneRuns {
				points := s.calculateRunPoints(run, scales, participant.GetCategory(), zone)
				result.ZoneResults[zoneIndex] = ZoneResult{
					Points:  points,
					Penalty: run.GetPenality(),
					Time:    run.GetChronoSec(),
				}

				if !result.HasError {
					result.TotalPoints += points
					result.TotalPenalty += run.GetPenality()
					result.TotalTime += run.GetChronoSec()
				}
				zoneIndex++
			}
		}

		results = append(results, result)
	}

	// Sort results by ranking (Total Points DESC, Total Penalty ASC, Total Time ASC)
	sort.Slice(results, func(i, j int) bool {
		if results[i].HasError && !results[j].HasError {
			return false
		}
		if !results[i].HasError && results[j].HasError {
			return true
		}
		if results[i].TotalPoints != results[j].TotalPoints {
			return results[i].TotalPoints > results[j].TotalPoints
		}
		if results[i].TotalPenalty != results[j].TotalPenalty {
			return results[i].TotalPenalty < results[j].TotalPenalty
		}
		return results[i].TotalTime < results[j].TotalTime
	})

	// Write data rows
	for i, result := range results {
		row := i + 2 // Start from row 2 (after headers)
		col := 0

		// Position
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), i+1)
		col++

		// Participant info
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), result.Participant.GetDossardNumber())
		col++
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), result.Participant.GetLastName())
		col++
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), result.Participant.GetFirstName())
		col++
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), result.Participant.GetClub())
		col++

		// Zone results
		for _, zoneResult := range result.ZoneResults {
			if zoneResult.IsError {
				f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), "ERROR")
				col++
				f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), "ERROR")
				col++
				f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), "ERROR")
				col++
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), zoneResult.Points)
				col++
				f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), zoneResult.Penalty)
				col++
				f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), zoneResult.Time)
				col++
			}
		}

		// Totals
		if result.HasError {
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), "ERROR")
			col++
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), "ERROR")
			col++
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), "ERROR")
			col++
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), "ERROR")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), result.TotalPoints)
			col++
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), result.TotalPenalty)
			col++
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), result.TotalTime)
			col++
			// Points earned based on ranking
			pointsEarned := utils.GetPointsEarned(int32(i + 1))
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune('A'+col)), row), pointsEarned)
		}
	}

	return nil
}

// Helper method to calculate points for a run
func (s *CompetitionService) calculateRunPoints(run *aggregate.Run, scales map[string]*aggregate.Scale, category, zone string) int32 {
	scaleKey := fmt.Sprintf("%s_%s", category, zone)
	scale, exists := scales[scaleKey]
	if !exists {
		return 0
	}

	points := int32(0)
	if run.GetDoor1() {
		points += scale.GetPointsDoor1()
	}
	if run.GetDoor2() {
		points += scale.GetPointsDoor2()
	}
	if run.GetDoor3() {
		points += scale.GetPointsDoor3()
	}
	if run.GetDoor4() {
		points += scale.GetPointsDoor4()
	}
	if run.GetDoor5() {
		points += scale.GetPointsDoor5()
	}
	if run.GetDoor6() {
		points += scale.GetPointsDoor6()
	}

	return points
}

func (s *CompetitionService) GetCompetition(ctx context.Context, competitionID int32) (*aggregate.Competition, error) {
	return s.competitionRepo.GetCompetition(ctx, competitionID)
}
