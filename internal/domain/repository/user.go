package repository

import (
	"context"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
)

type UserRepository interface {
	GetUser(ctx context.Context, id int32) (*aggregate.User, error)
	CreateUser(ctx context.Context, user *aggregate.User) error
	UpdateUser(ctx context.Context, user *aggregate.User) error
	DeleteUser(ctx context.Context, id int32) error
	GetUserByEmail(ctx context.Context, email string) (*aggregate.User, error)
}
