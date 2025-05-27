package repository

import (
	"context"
	"fmt"
	"goiam/internal/db"
	"goiam/internal/model"
)

// The user repository is a simplified interface for the user database, to be used by the auth service
// It porivides additional abstractions over the database, such as tenant and realm aware operations
type UserRepositoryImpl struct {
	tenant string
	realm  string
	db     db.UserDB
}

func NewUserRepository(tenant, realm string, db db.UserDB) UserRepositoryImpl {

	repo := &UserRepositoryImpl{tenant: tenant, realm: realm, db: db}

	return *repo
}

func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*model.User, error) {

	return r.db.GetUserByUsername(ctx, r.tenant, r.realm, username)
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User) error {

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

func (r *UserRepositoryImpl) Update(ctx context.Context, user *model.User) error {

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

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*model.User, error) {
	return r.db.GetUserByID(ctx, r.tenant, r.realm, id)
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.db.GetUserByEmail(ctx, r.tenant, r.realm, email)
}

func (r *UserRepositoryImpl) GetByLoginIdentifier(ctx context.Context, loginIdentifier string) (*model.User, error) {
	return r.db.GetUserByLoginIdentifier(ctx, r.tenant, r.realm, loginIdentifier)
}

func (r *UserRepositoryImpl) GetByFederatedIdentifier(ctx context.Context, provider, identifier string) (*model.User, error) {
	return r.db.GetUserByFederatedIdentifier(ctx, r.tenant, r.realm, provider, identifier)
}
