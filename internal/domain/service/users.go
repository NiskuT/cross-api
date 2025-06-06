package service

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type UserService interface {
	Login(ctx context.Context, email, password string) (*aggregate.JwtToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*aggregate.JwtToken, error)
	AddUserToCompetition(ctx context.Context, email string, competitionID int32) error
	InviteUser(ctx context.Context, firstName, lastName, email string, competitionID int32) error
	SetUserAsAdmin(ctx context.Context, email string, competitionID int32) (*aggregate.JwtToken, error)
}
