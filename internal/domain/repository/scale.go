package repository

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type ScaleRepository interface {
	GetScale(ctx context.Context, competitionID int32, category string) (*aggregate.Scale, error)
	CreateScale(ctx context.Context, scale *aggregate.Scale) error
	UpdateScale(ctx context.Context, scale *aggregate.Scale) error
	DeleteScale(ctx context.Context, competitionID int32, category string) error
}
