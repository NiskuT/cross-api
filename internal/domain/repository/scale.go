package repository

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type ScaleRepository interface {
	GetScale(ctx context.Context, competitionID int32, category string, zone string) (*aggregate.Scale, error)
	CreateScale(ctx context.Context, scale *aggregate.Scale) error
	UpdateScale(ctx context.Context, scale *aggregate.Scale) error
	DeleteScale(ctx context.Context, competitionID int32, category string, zone string) error
	ListZones(ctx context.Context, competitionID int32) ([]aggregate.ZoneInfo, error)
}
