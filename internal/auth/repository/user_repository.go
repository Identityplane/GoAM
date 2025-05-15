package repository

import (
	"context"
	"goiam/internal/model"
)

type Repositories struct {
	UserRepo UserRepository
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByLoginIdentifier(ctx context.Context, loginIdentifier string) (*model.User, error)

	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
}
