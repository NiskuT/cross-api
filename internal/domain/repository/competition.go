package repository

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type CompetitionRepository interface {
	GetCompetition(ctx context.Context, id int32) (*aggregate.Competition, error)
	CreateCompetition(ctx context.Context, competition *aggregate.Competition) (int32, error)
	UpdateCompetition(ctx context.Context, competition *aggregate.Competition) error
	DeleteCompetition(ctx context.Context, id int32) error
}
