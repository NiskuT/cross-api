package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	repo "github.com/NiskuT/cross-api/internal/domain/repository"
)

var (
	// ErrScaleNotFound is returned when a scale cannot be found
	ErrScaleNotFound = errors.New("scale not found")
	// ErrDuplicateScale is returned when a scale with the same competition ID, category, and zone already exists
	ErrDuplicateScale = errors.New("scale with this competition ID, category, and zone already exists")
)

// SQLScaleRepository is an implementation of the ScaleRepository interface that uses SQL
type SQLScaleRepository struct {
	db *sql.DB
}

// NewSQLScaleRepository creates a new SQLScaleRepository
func NewSQLScaleRepository(db *sql.DB) repo.ScaleRepository {
	return &SQLScaleRepository{
		db: db,
	}
}

// Scale is an internal representation of a scale for DB operations
type Scale struct {
	CompetitionID int32
	Category      string
	Zone          string
	PointsDoor1   int32
	PointsDoor2   int32
	PointsDoor3   int32
	PointsDoor4   int32
	PointsDoor5   int32
	PointsDoor6   int32
}

// GetScale retrieves a scale by its primary key (competition ID, category, zone)
func (r *SQLScaleRepository) GetScale(ctx context.Context, competitionID int32, category string, zone string) (*aggregate.Scale, error) {
	query := `
		SELECT competition_id, category, zone, points_door1, points_door2, points_door3, points_door4, points_door5, points_door6
		FROM scales
		WHERE competition_id = ? AND category = ? AND zone = ?
	`

	var scale Scale
	row := r.db.QueryRowContext(ctx, query, competitionID, category, zone)
	err := row.Scan(
		&scale.CompetitionID,
		&scale.Category,
		&scale.Zone,
		&scale.PointsDoor1,
		&scale.PointsDoor2,
		&scale.PointsDoor3,
		&scale.PointsDoor4,
		&scale.PointsDoor5,
		&scale.PointsDoor6,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrScaleNotFound
		}
		return nil, err
	}

	scaleAggregate := aggregate.NewScale()
	scaleAggregate.SetCompetitionID(scale.CompetitionID)
	scaleAggregate.SetCategory(scale.Category)
	scaleAggregate.SetZone(scale.Zone)
	scaleAggregate.SetPointsDoor1(scale.PointsDoor1)
	scaleAggregate.SetPointsDoor2(scale.PointsDoor2)
	scaleAggregate.SetPointsDoor3(scale.PointsDoor3)
	scaleAggregate.SetPointsDoor4(scale.PointsDoor4)
	scaleAggregate.SetPointsDoor5(scale.PointsDoor5)
	scaleAggregate.SetPointsDoor6(scale.PointsDoor6)

	return scaleAggregate, nil
}

// CreateScale creates a new scale
func (r *SQLScaleRepository) CreateScale(ctx context.Context, scale *aggregate.Scale) error {
	query := `
		INSERT INTO scales (competition_id, category, zone, points_door1, points_door2, points_door3, points_door4, points_door5, points_door6)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		scale.GetCompetitionID(),
		scale.GetCategory(),
		scale.GetZone(),
		scale.GetPointsDoor1(),
		scale.GetPointsDoor2(),
		scale.GetPointsDoor3(),
		scale.GetPointsDoor4(),
		scale.GetPointsDoor5(),
		scale.GetPointsDoor6(),
	)

	if err != nil {
		// Check for duplicate key error
		if isDuplicateKeyError(err) {
			return ErrDuplicateScale
		}
		return err
	}

	return nil
}

// UpdateScale updates an existing scale
func (r *SQLScaleRepository) UpdateScale(ctx context.Context, scale *aggregate.Scale) error {
	query := `
		UPDATE scales
		SET points_door1 = ?, points_door2 = ?, points_door3 = ?, points_door4 = ?, points_door5 = ?, points_door6 = ?
		WHERE competition_id = ? AND category = ? AND zone = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		scale.GetPointsDoor1(),
		scale.GetPointsDoor2(),
		scale.GetPointsDoor3(),
		scale.GetPointsDoor4(),
		scale.GetPointsDoor5(),
		scale.GetPointsDoor6(),
		scale.GetCompetitionID(),
		scale.GetCategory(),
		scale.GetZone(),
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrScaleNotFound
	}

	return nil
}

// DeleteScale deletes a scale by its primary key
func (r *SQLScaleRepository) DeleteScale(ctx context.Context, competitionID int32, category string, zone string) error {
	query := `
		DELETE FROM scales
		WHERE competition_id = ? AND category = ? AND zone = ?
	`

	result, err := r.db.ExecContext(ctx, query, competitionID, category, zone)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrScaleNotFound
	}

	return nil
}

// ListZones lists all zones for a competition
func (r *SQLScaleRepository) ListZones(ctx context.Context, competitionID int32) ([]aggregate.ZoneInfo, error) {
	query := `
		SELECT DISTINCT zone, category
		FROM scales
		WHERE competition_id = ?
		ORDER BY category, zone
	`

	rows, err := r.db.QueryContext(ctx, query, competitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []aggregate.ZoneInfo
	for rows.Next() {
		var zone, category string
		if err := rows.Scan(&zone, &category); err != nil {
			return nil, err
		}

		zoneInfo := aggregate.NewZoneInfo()
		zoneInfo.SetZone(zone)
		zoneInfo.SetCategory(category)
		zones = append(zones, *zoneInfo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return zones, nil
}
