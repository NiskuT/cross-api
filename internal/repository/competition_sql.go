package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	repo "github.com/NiskuT/cross-api/internal/domain/repository"
)

var (
	// ErrCompetitionNotFound is returned when a competition cannot be found
	ErrCompetitionNotFound = errors.New("competition not found")
	// ErrDuplicateCompetition is returned when a competition with the same ID already exists
	ErrDuplicateCompetition = errors.New("competition with this ID already exists")
)

// SQLCompetitionRepository is an implementation of the CompetitionRepository interface that uses SQL
type SQLCompetitionRepository struct {
	db *sql.DB
}

// NewSQLCompetitionRepository creates a new SQLCompetitionRepository
func NewSQLCompetitionRepository(db *sql.DB) repo.CompetitionRepository {
	return &SQLCompetitionRepository{
		db: db,
	}
}

// Competition is an internal representation of a competition for DB operations
type Competition struct {
	ID          int32
	Name        string
	Description string
	Date        string
	Location    string
	Organizer   string
	Contact     string
}

// GetCompetition retrieves a competition by ID
func (r *SQLCompetitionRepository) GetCompetition(ctx context.Context, id int32) (*aggregate.Competition, error) {
	query := `
		SELECT id, name, description, date, location, organizer, contact
		FROM competitions
		WHERE id = ?
	`

	var competition Competition
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&competition.ID,
		&competition.Name,
		&competition.Description,
		&competition.Date,
		&competition.Location,
		&competition.Organizer,
		&competition.Contact,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCompetitionNotFound
		}
		return nil, err
	}

	competitionAggregate := aggregate.NewCompetition()
	competitionAggregate.SetID(competition.ID)
	competitionAggregate.SetName(competition.Name)
	competitionAggregate.SetDescription(competition.Description)
	competitionAggregate.SetDate(competition.Date)
	competitionAggregate.SetLocation(competition.Location)
	competitionAggregate.SetOrganizer(competition.Organizer)
	competitionAggregate.SetContact(competition.Contact)

	return competitionAggregate, nil
}

// CreateCompetition creates a new competition
func (r *SQLCompetitionRepository) CreateCompetition(ctx context.Context, competition *aggregate.Competition) (int32, error) {
	query := `
		INSERT INTO competitions (name, description, date, location, organizer, contact)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		competition.GetName(),
		competition.GetDescription(),
		competition.GetDate(),
		competition.GetLocation(),
		competition.GetOrganizer(),
		competition.GetContact(),
	)

	if err != nil {
		// Check for duplicate key error
		if isDuplicateKeyError(err) {
			return 0, ErrDuplicateCompetition
		}
		return 0, err
	}

	// Get the auto-incremented ID
	if id, err := result.LastInsertId(); err == nil {
		competition.SetID(int32(id))
	}

	return competition.GetID(), nil
}

// UpdateCompetition updates an existing competition
func (r *SQLCompetitionRepository) UpdateCompetition(ctx context.Context, competition *aggregate.Competition) error {
	query := `
		UPDATE competitions
		SET name = ?, description = ?, date = ?, location = ?, organizer = ?, contact = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		competition.GetName(),
		competition.GetDescription(),
		competition.GetDate(),
		competition.GetLocation(),
		competition.GetOrganizer(),
		competition.GetContact(),
		competition.GetID(),
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCompetitionNotFound
	}

	return nil
}

// DeleteCompetition deletes a competition by ID
func (r *SQLCompetitionRepository) DeleteCompetition(ctx context.Context, id int32) error {
	query := `
		DELETE FROM competitions
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCompetitionNotFound
	}

	return nil
}

// ListCompetitions lists all competitions
func (r *SQLCompetitionRepository) ListCompetitions(ctx context.Context) ([]*aggregate.Competition, error) {

	query := `
		SELECT id, name, description, date, location, organizer, contact
		FROM competitions
		ORDER BY date DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	competitions := []*aggregate.Competition{}

	for rows.Next() {
		var competition Competition
		if err := rows.Scan(&competition.ID, &competition.Name, &competition.Description, &competition.Date, &competition.Location, &competition.Organizer, &competition.Contact); err != nil {
			return nil, err
		}

		competitionAggregate := aggregate.NewCompetition()
		competitionAggregate.SetID(competition.ID)
		competitionAggregate.SetName(competition.Name)
		competitionAggregate.SetDescription(competition.Description)
		competitionAggregate.SetDate(competition.Date)
		competitionAggregate.SetLocation(competition.Location)
		competitionAggregate.SetOrganizer(competition.Organizer)
		competitionAggregate.SetContact(competition.Contact)
		competitions = append(competitions, competitionAggregate)
	}

	return competitions, nil
}
