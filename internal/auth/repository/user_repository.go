package repository

import (
	"context"
	"goiam/internal/db/model"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
}
