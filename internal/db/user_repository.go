package db

import (
	"context"
	"fmt"
	"goiam/internal/auth/repository"
	"goiam/internal/model"
)

// The user repository is a simplified interface for the user database, to be used by the auth service
// It porivides additional abstractions over the database, such as tenant and realm aware operations
type UserRepository struct {
	tenant string
	realm  string
	db     UserDB
}

func NewUserRepository(tenant, realm string, db UserDB) repository.UserRepository {

	repo := &UserRepository{tenant: tenant, realm: realm, db: db}

	return repo
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {

	return r.db.GetUserByUsername(ctx, r.tenant, r.realm, username)
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {

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

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {

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
