package sqlite_adapter

import (
	"context"
	"fmt"
	"goiam/internal/auth/repository"
	"goiam/internal/db"
	"goiam/internal/model"
)

type SQLiteUserRepository struct {
	tenant string
	realm  string
	db     db.UserDB
}

func NewUserRepository(tenant, realm string, db db.UserDB) repository.UserRepository {
	return &SQLiteUserRepository{tenant: tenant, realm: realm, db: db}
}

func (r *SQLiteUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {

	return r.db.GetUserByUsername(ctx, r.tenant, r.realm, username)
}

func (r *SQLiteUserRepository) Create(ctx context.Context, user *model.User) error {

	// panic if the tenant or realm is set to a different value except ""
	if user.Tenant != "" && user.Tenant != r.tenant {
		return fmt.Errorf("tenant is set to a different value")
	}
	if user.Realm != "" && user.Realm != r.realm {
		return fmt.Errorf("realm is set to a different value")
	}

	// Ensure the tenant and realm are set to the repository values
	user.Tenant = r.tenant
	user.Realm = r.realm

	return r.db.CreateUser(ctx, *user)
}

func (r *SQLiteUserRepository) Update(ctx context.Context, user *model.User) error {

	// panic if the tenant or realm is set to a different value except ""
	if user.Tenant != "" && user.Tenant != r.tenant {
		return fmt.Errorf("tenant is set to a different value")
	}
	if user.Realm != "" && user.Realm != r.realm {
		return fmt.Errorf("realm is set to a different value")
	}

	// Ensure the tenant and realm are set to the repository values
	user.Tenant = r.tenant
	user.Realm = r.realm

	return r.db.UpdateUser(ctx, user)
}
