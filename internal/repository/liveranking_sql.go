package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	repo "github.com/NiskuT/cross-api/internal/domain/repository"
)

var (
	// ErrLiverankingNotFound is returned when a liveranking cannot be found
	ErrLiverankingNotFound = errors.New("liveranking not found")
)

// SQLLiverankingRepository is an implementation of the LiverankingRepository interface that uses SQL
type SQLLiverankingRepository struct {
	db *sql.DB
}

// NewSQLLiverankingRepository creates a new SQLLiverankingRepository
func NewSQLLiverankingRepository(db *sql.DB) repo.LiverankingRepository {
	return &SQLLiverankingRepository{
		db: db,
	}
}

// UpsertLiveranking creates a new liveranking if it doesn't exist, or adds the points and penality to the existing liveranking
func (r *SQLLiverankingRepository) UpsertLiveranking(ctx context.Context, liveranking *aggregate.Liveranking) error {
	// First check if liveranking exists
	query := `
		SELECT EXISTS(
			SELECT 1 FROM liverankings 
			WHERE competition_id = ? AND dossard_number = ?
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, liveranking.GetCompetitionID(), liveranking.GetDossard()).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Update existing liveranking
		updateQuery := `
			UPDATE liverankings
			SET number_of_runs = number_of_runs + 1,
				total_points = total_points + ?,
				penality = penality + ?,
				chrono_sec = chrono_sec + ?
			WHERE competition_id = ? AND dossard_number = ?
		`
		_, err = r.db.ExecContext(
			ctx,
			updateQuery,
			liveranking.GetTotalPoints(),
			liveranking.GetPenality(),
			liveranking.GetChronoSec(),
			liveranking.GetCompetitionID(),
			liveranking.GetDossard(),
		)
		return err
	}

	// If the liveranking doesn't exist, we need to create it

	// Check if participant exists
	participantQuery := `
		SELECT EXISTS(
			SELECT 1 FROM participants 
			WHERE competition_id = ? AND dossard_number = ?
		)
	`
	var participantExists bool
	err = r.db.QueryRowContext(ctx, participantQuery, liveranking.GetCompetitionID(), liveranking.GetDossard()).Scan(&participantExists)
	if err != nil {
		return err
	}

	if !participantExists {
		return ErrParticipantNotFound
	}

	// Insert new liveranking
	insertQuery := `
		INSERT INTO liverankings (competition_id, dossard_number, number_of_runs, total_points, penality, chrono_sec)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(
		ctx,
		insertQuery,
		liveranking.GetCompetitionID(),
		liveranking.GetDossard(),
		1, // Starting with 1 run
		liveranking.GetTotalPoints(),
		liveranking.GetPenality(),
		liveranking.GetChronoSec(),
	)
	return err
}

// ListLiveranking lists liveranking entries sorted by desc total points, asc penality, and desc chrono sec
func (r *SQLLiverankingRepository) ListLiveranking(ctx context.Context, competitionID, pageNumber, pageSize int32) ([]*aggregate.Liveranking, int32, error) {
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	if pageNumber <= 0 {
		pageNumber = 1 // Default page number
	}

	// Get total count first
	countQuery := `
		SELECT COUNT(*)
		FROM liverankings l
		WHERE l.competition_id = ?
	`
	var totalCount int32
	err := r.db.QueryRowContext(ctx, countQuery, competitionID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	offset := (pageNumber - 1) * pageSize

	query := `
		SELECT l.competition_id, l.dossard_number, p.first_name, p.last_name, p.category, 
		       l.number_of_runs, l.total_points, l.penality, l.chrono_sec
		FROM liverankings l
		JOIN participants p ON l.competition_id = p.competition_id AND l.dossard_number = p.dossard_number
		WHERE l.competition_id = ?
		ORDER BY l.total_points DESC, l.penality ASC, l.chrono_sec DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, competitionID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var liverankings []*aggregate.Liveranking
	for rows.Next() {
		liveranking := aggregate.NewLiveranking()

		var competitionID, dossardNumber, numberOfRuns, totalPoints, penality, chronoSec int32
		var firstName, lastName, category string

		err := rows.Scan(
			&competitionID,
			&dossardNumber,
			&firstName,
			&lastName,
			&category,
			&numberOfRuns,
			&totalPoints,
			&penality,
			&chronoSec,
		)
		if err != nil {
			return nil, 0, err
		}

		liveranking.SetCompetitionID(competitionID)
		liveranking.SetDossard(dossardNumber)
		liveranking.SetFirstName(firstName)
		liveranking.SetLastName(lastName)
		liveranking.SetCategory(category)
		liveranking.SetNumberOfRuns(numberOfRuns)
		liveranking.SetTotalPoints(totalPoints)
		liveranking.SetPenality(penality)
		liveranking.SetChronoSec(chronoSec)

		liverankings = append(liverankings, liveranking)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return liverankings, totalCount, nil
}
