package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	repo "github.com/NiskuT/cross-api/internal/domain/repository"
)

var (
	// ErrRunNotFound is returned when a run cannot be found
	ErrRunNotFound = errors.New("run not found")
	// ErrDuplicateRun is returned when a run with the same keys already exists
	ErrDuplicateRun = errors.New("run with this combination of competition ID, run number, and dossard already exists")
	// ErrParticipantNotFoundForRun is returned when trying to create a run for a non-existent participant
	ErrParticipantNotFoundForRun = errors.New("participant not found for this run")
)

// SQLRunRepository is an implementation of the RunRepository interface that uses SQL
type SQLRunRepository struct {
	db *sql.DB
}

// NewSQLRunRepository creates a new SQLRunRepository
func NewSQLRunRepository(db *sql.DB) repo.RunRepository {
	return &SQLRunRepository{
		db: db,
	}
}

// Run is an internal representation of a run for DB operations
type Run struct {
	CompetitionID int32
	Dossard       int32
	RunNumber     int32
	Zone          string
	Door1         bool
	Door2         bool
	Door3         bool
	Door4         bool
	Door5         bool
	Door6         bool
	Penality      int32
	ChronoSec     int32
	RefereeId     int32
}

// GetRun retrieves a run by its primary key (competition ID, run number, dossard)
func (r *SQLRunRepository) GetRun(ctx context.Context, competitionID, runNumber, dossard int32) (*aggregate.Run, error) {
	query := `
		SELECT competition_id, dossard, run_number, zone, door1, door2, door3, door4, door5, door6, penality, chrono_sec, referee_id
		FROM runs
		WHERE competition_id = ? AND run_number = ? AND dossard = ?
	`

	var run Run
	row := r.db.QueryRowContext(ctx, query, competitionID, runNumber, dossard)
	err := row.Scan(
		&run.CompetitionID,
		&run.Dossard,
		&run.RunNumber,
		&run.Zone,
		&run.Door1,
		&run.Door2,
		&run.Door3,
		&run.Door4,
		&run.Door5,
		&run.Door6,
		&run.Penality,
		&run.ChronoSec,
		&run.RefereeId,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRunNotFound
		}
		return nil, err
	}

	return mapToRunAggregate(&run), nil
}

// ListRuns lists all runs for a competition
func (r *SQLRunRepository) ListRuns(ctx context.Context, competitionID int32) ([]*aggregate.Run, error) {
	query := `
		SELECT competition_id, dossard, run_number, zone, door1, door2, door3, door4, door5, door6, penality, chrono_sec, referee_id
		FROM runs
		WHERE competition_id = ?
		ORDER BY dossard, run_number
	`

	rows, err := r.db.QueryContext(ctx, query, competitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*aggregate.Run
	for rows.Next() {
		var run Run
		err := rows.Scan(
			&run.CompetitionID,
			&run.Dossard,
			&run.RunNumber,
			&run.Zone,
			&run.Door1,
			&run.Door2,
			&run.Door3,
			&run.Door4,
			&run.Door5,
			&run.Door6,
			&run.Penality,
			&run.ChronoSec,
			&run.RefereeId,
		)

		if err != nil {
			return nil, err
		}

		runs = append(runs, mapToRunAggregate(&run))
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}

// ListRunsByDossard lists all runs for a specific participant in a competition
func (r *SQLRunRepository) ListRunsByDossard(ctx context.Context, competitionID int32, dossard int32) ([]*aggregate.Run, error) {
	query := `
		SELECT competition_id, dossard, run_number, zone, door1, door2, door3, door4, door5, door6, penality, chrono_sec, referee_id
		FROM runs
		WHERE competition_id = ? AND dossard = ?
		ORDER BY run_number
	`

	rows, err := r.db.QueryContext(ctx, query, competitionID, dossard)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*aggregate.Run
	for rows.Next() {
		var run Run
		err := rows.Scan(
			&run.CompetitionID,
			&run.Dossard,
			&run.RunNumber,
			&run.Zone,
			&run.Door1,
			&run.Door2,
			&run.Door3,
			&run.Door4,
			&run.Door5,
			&run.Door6,
			&run.Penality,
			&run.ChronoSec,
			&run.RefereeId,
		)

		if err != nil {
			return nil, err
		}

		runs = append(runs, mapToRunAggregate(&run))
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}

// ListRunsByDossardWithDetails retrieves all runs for a participant with referee names
func (r *SQLRunRepository) ListRunsByDossardWithDetails(ctx context.Context, competitionID, dossard int32) ([]*aggregate.Run, error) {
	query := `
		SELECT 
			r.competition_id, r.dossard, r.run_number, r.zone,
			r.door1, r.door2, r.door3, r.door4, r.door5, r.door6,
			r.penality, r.chrono_sec, r.referee_id,
			COALESCE(u.name, '') as referee_name
		FROM runs r
		LEFT JOIN users u ON r.referee_id = u.id
		WHERE r.competition_id = ? AND r.dossard = ?
		ORDER BY r.run_number
	`

	rows, err := r.db.QueryContext(ctx, query, competitionID, dossard)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*aggregate.Run
	for rows.Next() {
		var run Run
		var refereeName string

		err := rows.Scan(
			&run.CompetitionID,
			&run.Dossard,
			&run.RunNumber,
			&run.Zone,
			&run.Door1,
			&run.Door2,
			&run.Door3,
			&run.Door4,
			&run.Door5,
			&run.Door6,
			&run.Penality,
			&run.ChronoSec,
			&run.RefereeId,
			&refereeName,
		)
		if err != nil {
			return nil, err
		}

		// Map to aggregate and set referee name
		runAggregate := mapToRunAggregate(&run)
		runAggregate.SetRefereeName(refereeName)

		runs = append(runs, runAggregate)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}

// CreateRun creates a new run with auto-incrementing run number per participant
func (r *SQLRunRepository) CreateRun(ctx context.Context, run *aggregate.Run) error {
	// First, verify that the participant exists
	checkQuery := `
		SELECT 1 FROM participants
		WHERE competition_id = ? AND dossard_number = ?
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, checkQuery, run.GetCompetitionID(), run.GetDossard()).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrParticipantNotFoundForRun
		}
		return err
	}

	// Auto-increment the run number for this participant if not explicitly set
	if run.GetRunNumber() <= 0 {
		// Find the max run number for this participant
		maxQuery := `
			SELECT COALESCE(MAX(run_number), 0)
			FROM runs
			WHERE competition_id = ? AND dossard = ?
		`
		var maxRunNumber int32
		err := r.db.QueryRowContext(ctx, maxQuery, run.GetCompetitionID(), run.GetDossard()).Scan(&maxRunNumber)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// Set the next run number
		run.SetRunNumber(maxRunNumber + 1)
	}

	// Now insert the run with the calculated run number
	query := `
		INSERT INTO runs (competition_id, dossard, run_number, zone, door1, door2, door3, door4, door5, door6, penality, chrono_sec, referee_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		run.GetCompetitionID(),
		run.GetDossard(),
		run.GetRunNumber(),
		run.GetZone(),
		run.GetDoor1(),
		run.GetDoor2(),
		run.GetDoor3(),
		run.GetDoor4(),
		run.GetDoor5(),
		run.GetDoor6(),
		run.GetPenality(),
		run.GetChronoSec(),
		run.GetRefereeId(),
	)

	if err != nil {
		// Check for duplicate key error
		if isDuplicateKeyError(err) {
			return ErrDuplicateRun
		}
		return err
	}

	return nil
}

// UpdateRun updates an existing run
func (r *SQLRunRepository) UpdateRun(ctx context.Context, run *aggregate.Run) error {
	query := `
		UPDATE runs
		SET zone = ?, door1 = ?, door2 = ?, door3 = ?, door4 = ?, door5 = ?, door6 = ?, penality = ?, chrono_sec = ?, referee_id = ?
		WHERE competition_id = ? AND run_number = ? AND dossard = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		run.GetZone(),
		run.GetDoor1(),
		run.GetDoor2(),
		run.GetDoor3(),
		run.GetDoor4(),
		run.GetDoor5(),
		run.GetDoor6(),
		run.GetPenality(),
		run.GetChronoSec(),
		run.GetRefereeId(),
		run.GetCompetitionID(),
		run.GetRunNumber(),
		run.GetDossard(),
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRunNotFound
	}

	return nil
}

// DeleteRun deletes a run by its primary key
func (r *SQLRunRepository) DeleteRun(ctx context.Context, competitionID, runNumber, dossard int32) error {
	query := `
		DELETE FROM runs
		WHERE competition_id = ? AND run_number = ? AND dossard = ?
	`

	result, err := r.db.ExecContext(ctx, query, competitionID, runNumber, dossard)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRunNotFound
	}

	return nil
}

// Helper function to map a Run struct to a Run aggregate
func mapToRunAggregate(run *Run) *aggregate.Run {
	runAggregate := aggregate.NewRun()
	runAggregate.SetCompetitionID(run.CompetitionID)
	runAggregate.SetDossard(run.Dossard)
	runAggregate.SetRunNumber(run.RunNumber)
	runAggregate.SetZone(run.Zone)
	runAggregate.SetDoor1(run.Door1)
	runAggregate.SetDoor2(run.Door2)
	runAggregate.SetDoor3(run.Door3)
	runAggregate.SetDoor4(run.Door4)
	runAggregate.SetDoor5(run.Door5)
	runAggregate.SetDoor6(run.Door6)
	runAggregate.SetPenality(run.Penality)
	runAggregate.SetChronoSec(run.ChronoSec)
	runAggregate.SetRefereeId(run.RefereeId)
	return runAggregate
}
