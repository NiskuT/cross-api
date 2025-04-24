package service

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type CompetitionService interface {
	CreateCompetition(ctx context.Context, competition *aggregate.Competition) (int32, error)
	AddZone(ctx context.Context, competitionID int32, zone *aggregate.Scale) error
}
