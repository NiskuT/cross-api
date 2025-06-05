package repository

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type ParticipantRepository interface {
	GetParticipant(ctx context.Context, competitionID int32, dossardNumber int32) (*aggregate.Participant, error)
	CreateParticipant(ctx context.Context, participant *aggregate.Participant) error
	UpdateParticipant(ctx context.Context, participant *aggregate.Participant) error
	DeleteParticipant(ctx context.Context, competitionID int32, dossardNumber int32) error
	ListParticipantsByCategory(ctx context.Context, competitionID int32, category string) ([]*aggregate.Participant, error)
}
