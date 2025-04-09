package sqlite

import (
	"context"
	"goiam/internal/auth/repository"
	"goiam/internal/db/model"
)

type SQLiteUserRepository struct{}

func NewUserRepository() repository.UserRepository {
	return &SQLiteUserRepository{}
}

func (r *SQLiteUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return GetUserByUsername(ctx, username)
}

func (r *SQLiteUserRepository) Create(ctx context.Context, user *model.User) error {
	return CreateUser(ctx, *user)
}

func (r *SQLiteUserRepository) Update(ctx context.Context, user *model.User) error {
	return UpdateUser(ctx, user)
}
