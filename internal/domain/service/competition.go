package service

import (
	"context"
	"io"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type CompetitionService interface {
	CreateCompetition(ctx context.Context, competition *aggregate.Competition) (int32, error)
	AddScale(ctx context.Context, competitionID int32, scale *aggregate.Scale) error
	AddParticipants(ctx context.Context, competitionID int32, category string, excelFile io.Reader) error
	ListCompetitions(ctx context.Context) ([]*aggregate.Competition, error)
	GetParticipant(ctx context.Context, competitionID int32, dossardNumber int32) (*aggregate.Participant, error)
	ListZones(ctx context.Context, competitionID int32) ([]aggregate.ZoneInfo, error)
	GetScale(ctx context.Context, competitionID int32, category string, zone string) (*aggregate.Scale, error)
	UpdateScale(ctx context.Context, competitionID int32, scale *aggregate.Scale) error
	DeleteScale(ctx context.Context, competitionID int32, category string, zone string) error
}
