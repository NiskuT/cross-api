package service

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

// RunService defines the operations for managing runs
type RunService interface {
	// CreateRun creates a new run and updates the liveranking
	CreateRun(ctx context.Context, run *aggregate.Run) error

	// GetRun retrieves a run by its identifiers
	GetRun(ctx context.Context, competitionID, runNumber, dossard int32) (*aggregate.Run, error)

	// ListRuns lists all runs for a competition
	ListRuns(ctx context.Context, competitionID int32) ([]*aggregate.Run, error)

	// ListRunsByDossard lists all runs for a participant in a competition
	ListRunsByDossard(ctx context.Context, competitionID int32, dossard int32) ([]*aggregate.Run, error)

	// UpdateRun updates an existing run
	UpdateRun(ctx context.Context, run *aggregate.Run) error

	// DeleteRun deletes a run
	DeleteRun(ctx context.Context, competitionID, runNumber, dossard int32) error
}
