package repository

import (
	"context"
	"fmt"

	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
)

// The user repository is a simplified interface for the user database, to be used by the auth service
// It porivides additional abstractions over the database, such as tenant and realm aware operations
type UserRepositoryImpl struct {
	tenant       string
	realm        string
	db           db.UserDB
	attributesDB db.UserAttributeDB
}

func NewUserRepository(tenant, realm string, db db.UserDB, attributesDB db.UserAttributeDB) model.UserRepository {

	repo := &UserRepositoryImpl{tenant: tenant, realm: realm, db: db, attributesDB: attributesDB}

	return repo
}

func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*model.User, error) {

	return r.db.GetUserByUsername(ctx, r.tenant, r.realm, username)
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User) error {

	// Ensure the tenant and realm are set to the repository values
	r.ensureTenantAndRealm(user, r.tenant, r.realm)

	// Create the user with all attributes in one transaction
	return r.attributesDB.CreateUserWithAttributes(ctx, user)
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *model.User) error {

	// If the user has no id we return an error
	if user.ID == "" {
		return fmt.Errorf("user has no id - create user first")
	}

	// Ensure the tenant and realm are set to the repository values
	r.ensureTenantAndRealm(user, r.tenant, r.realm)

	// Ensure the tenant and realm are set to the repository values
	user.Tenant = r.tenant
	user.Realm = r.realm

	return r.db.UpdateUser(ctx, user)
}

func (r *UserRepositoryImpl) CreateOrUpdate(ctx context.Context, user *model.User) error {

	// If the user has no id we create a new one
	if user.ID == "" {
		return r.Create(ctx, user)
	}

	return r.Update(ctx, user)
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

func (r *UserRepositoryImpl) CreateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error {
	return r.attributesDB.CreateUserAttribute(ctx, *attribute)
}

func (r *UserRepositoryImpl) UpdateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error {
	return r.attributesDB.UpdateUserAttribute(ctx, attribute)
}

func (r *UserRepositoryImpl) DeleteUserAttribute(ctx context.Context, attributeID string) error {
	return r.attributesDB.DeleteUserAttribute(ctx, r.tenant, r.realm, attributeID)
}

// ensureTenantAndRealm ensures that the tenant and realm are set to the repository values
// If there is no user id it creates a new user id
// Also for each attributes it ensures that the tenant and realm are set and the user id is set as well as a attribute id
func (r *UserRepositoryImpl) ensureTenantAndRealm(user *model.User, tenant, realm string) {

	// If the tenant or realm is set to a different value except "" we return an error
	// If there is no user id we create a new one
	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	// Override the tenant and realm
	user.Tenant = tenant
	user.Realm = realm

	// For each attribute we ensure that the tenant and realm are set and the user id is set as well as a attribute id
	for _, attribute := range user.UserAttributes {
		attribute.Tenant = tenant
		attribute.Realm = realm
		attribute.UserID = user.ID

		if attribute.ID == "" {
			attribute.ID = uuid.NewString()
		}
	}
}
