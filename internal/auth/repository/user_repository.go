package repository

import (
	"context"
	"goiam/internal/model"
)

type ServiceRegistry struct {
	UserRepo UserRepository
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
}
