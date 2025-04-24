package service

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type UserService interface {
	Login(ctx context.Context, email, password string) (*aggregate.JwtToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*aggregate.JwtToken, error)
}
