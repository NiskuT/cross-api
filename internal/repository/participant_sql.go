package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	repo "github.com/NiskuT/cross-api/internal/domain/repository"
	"github.com/go-sql-driver/mysql"
)

var (
	// ErrParticipantNotFound is returned when a participant cannot be found
	ErrParticipantNotFound = errors.New("participant not found")
	// ErrDuplicateParticipant is returned when a participant with the same competition ID and dossard number already exists
	ErrDuplicateParticipant = errors.New("participant with this competition ID and dossard number already exists")
	// ErrInvalidCompetitionID is returned when the competition ID format is invalid
	ErrInvalidCompetitionID = errors.New("invalid competition ID format")
)

// SQLParticipantRepository is an implementation of the ParticipantRepository interface that uses SQL
type SQLParticipantRepository struct {
	db *sql.DB
}

// NewSQLParticipantRepository creates a new SQLParticipantRepository
func NewSQLParticipantRepository(db *sql.DB) repo.ParticipantRepository {
	return &SQLParticipantRepository{
		db: db,
	}
}

type Participant struct {
	CompetitionID int32
	DossardNumber int32
	FirstName     string
	LastName      string
	Category      string
	Gender        string
}

// GetParticipant retrieves a participant by competition ID and dossard number
func (r *SQLParticipantRepository) GetParticipant(ctx context.Context, competitionID int32, dossardNumber int32) (*aggregate.Participant, error) {
	query := `
		SELECT competition_id, dossard_number, first_name, last_name, category, gender
		FROM participants
		WHERE competition_id = ? AND dossard_number = ?
	`

	var participant Participant
	row := r.db.QueryRowContext(ctx, query, competitionID, dossardNumber)
	err := row.Scan(
		&participant.CompetitionID,
		&participant.DossardNumber,
		&participant.FirstName,
		&participant.LastName,
		&participant.Category,
		&participant.Gender,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrParticipantNotFound
		}
		return nil, err
	}

	participantAggregate := aggregate.NewParticipant()
	participantAggregate.SetCompetitionID(participant.CompetitionID)
	participantAggregate.SetDossardNumber(participant.DossardNumber)
	participantAggregate.SetFirstName(participant.FirstName)
	participantAggregate.SetLastName(participant.LastName)
	participantAggregate.SetCategory(participant.Category)
	participantAggregate.SetGender(participant.Gender)

	return participantAggregate, nil
}

// CreateParticipant creates a new participant
func (r *SQLParticipantRepository) CreateParticipant(ctx context.Context, participant *aggregate.Participant) error {
	query := `
		INSERT INTO participants (competition_id, dossard_number, first_name, last_name, category, gender)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		participant.GetCompetitionID(),
		participant.GetDossardNumber(),
		participant.GetFirstName(),
		participant.GetLastName(),
		participant.GetCategory(),
		participant.GetGender(),
	)

	if err != nil {
		// Check for duplicate key error
		if isDuplicateKeyError(err) {
			return ErrDuplicateParticipant
		}
		return err
	}

	return nil
}

// UpdateParticipant updates an existing participant
func (r *SQLParticipantRepository) UpdateParticipant(ctx context.Context, participant *aggregate.Participant) error {
	query := `
		UPDATE participants
		SET first_name = ?, last_name = ?, category = ?, gender = ?
		WHERE competition_id = ? AND dossard_number = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		participant.GetFirstName(),
		participant.GetLastName(),
		participant.GetCategory(),
		participant.GetGender(),
		participant.GetCompetitionID(),
		participant.GetDossardNumber(),
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrParticipantNotFound
	}

	return nil
}

// DeleteParticipant deletes a participant by competition ID and dossard number
func (r *SQLParticipantRepository) DeleteParticipant(ctx context.Context, competitionID int32, dossardNumber int32) error {
	query := `
		DELETE FROM participants
		WHERE competition_id = ? AND dossard_number = ?
	`

	result, err := r.db.ExecContext(ctx, query, competitionID, dossardNumber)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrParticipantNotFound
	}

	return nil
}

// ListParticipantsByCategory retrieves all participants for a competition by category
func (r *SQLParticipantRepository) ListParticipantsByCategory(ctx context.Context, competitionID int32, category string) ([]*aggregate.Participant, error) {
	query := `
		SELECT competition_id, dossard_number, first_name, last_name, category, gender
		FROM participants
		WHERE competition_id = ? AND category = ?
		ORDER BY dossard_number
	`

	rows, err := r.db.QueryContext(ctx, query, competitionID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*aggregate.Participant
	for rows.Next() {
		var participant Participant
		err := rows.Scan(
			&participant.CompetitionID,
			&participant.DossardNumber,
			&participant.FirstName,
			&participant.LastName,
			&participant.Category,
			&participant.Gender,
		)

		if err != nil {
			return nil, err
		}

		participantAggregate := aggregate.NewParticipant()
		participantAggregate.SetCompetitionID(participant.CompetitionID)
		participantAggregate.SetDossardNumber(participant.DossardNumber)
		participantAggregate.SetFirstName(participant.FirstName)
		participantAggregate.SetLastName(participant.LastName)
		participantAggregate.SetCategory(participant.Category)
		participantAggregate.SetGender(participant.Gender)

		participants = append(participants, participantAggregate)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return participants, nil
}

// Helper function to check if an error is a duplicate key error
func isDuplicateKeyError(err error) bool {
	// For MySQL
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	return false
}
