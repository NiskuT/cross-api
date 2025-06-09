package repository

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type RunRepository interface {
	CreateRun(ctx context.Context, run *aggregate.Run) error
	GetRun(ctx context.Context, competitionID, runNumber, dossard int32) (*aggregate.Run, error)
	ListRuns(ctx context.Context, competitionID int32) ([]*aggregate.Run, error)
	ListRunsByDossard(ctx context.Context, competitionID int32, dossard int32) ([]*aggregate.Run, error)
	ListRunsByDossardWithDetails(ctx context.Context, competitionID int32, dossard int32) ([]*aggregate.Run, error)
	UpdateRun(ctx context.Context, run *aggregate.Run) error
	DeleteRun(ctx context.Context, competitionID, runNumber, dossard int32) error
}
