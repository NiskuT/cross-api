package repository

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type LiverankingRepository interface {
	UpsertLiveranking(ctx context.Context, liveranking *aggregate.Liveranking) error                                                                          // This function will create a new liveranking if it doesn't exist, or ADD the points and penality to the existing liveranking
	ListLiveranking(ctx context.Context, competitionID, pageNumber, pageSize int32) ([]*aggregate.Liveranking, int32, error)                                  // This list function is sorted by desc total points and asc penality and desc chrono sec, also returns total count for pagination
	ListLiverankingByCategory(ctx context.Context, competitionID int32, category string, pageNumber, pageSize int32) ([]*aggregate.Liveranking, int32, error) // This list function filters by category and is sorted by desc total points and asc penality and desc chrono sec, also returns total count for pagination
}
