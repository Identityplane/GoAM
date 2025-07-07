package db

import (
	"context"

	"github.com/gianlucafrei/GoAM/internal/model"
)

// Interface for the user db
type UserDB interface {
	CreateUser(ctx context.Context, user model.User) error
	GetUserByUsername(ctx context.Context, tenant, realm, username string) (*model.User, error)
	GetUserByID(ctx context.Context, tenant, realm, userID string) (*model.User, error)
	GetUserByLoginIdentifier(ctx context.Context, tenant, realm, loginIdentifier string) (*model.User, error)
	GetUserByEmail(ctx context.Context, tenant, realm, email string) (*model.User, error)
	GetUserByFederatedIdentifier(ctx context.Context, tenant, realm, provider, identifier string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	ListUsers(ctx context.Context, tenant, realm string) ([]model.User, error)
	ListUsersWithPagination(ctx context.Context, tenant, realm string, offset, limit int) ([]model.User, error)
	CountUsers(ctx context.Context, tenant, realm string) (int64, error)
	GetUserStats(ctx context.Context, tenant, realm string) (*model.UserStats, error)
	DeleteUser(ctx context.Context, tenant, realm, username string) error
}
