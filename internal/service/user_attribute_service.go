package service

import (
	"context"

	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
)

// UserAttributeService defines the business logic for user attribute operations
type UserAttributeService interface {
	// List all attributes for a user
	ListUserAttributes(ctx context.Context, tenant, realm, userID string) ([]model.UserAttribute, error)
	// Get a specific attribute by ID
	GetUserAttributeByID(ctx context.Context, tenant, realm, attributeID string) (*model.UserAttribute, error)
	// Create a new attribute for a user
	CreateUserAttribute(ctx context.Context, attribute model.UserAttribute) (*model.UserAttribute, error)
	// Update an existing attribute
	UpdateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error
	// Delete a specific attribute
	DeleteUserAttribute(ctx context.Context, tenant, realm, attributeID string) error
}

// userAttributeServiceImpl implements UserAttributeService
type userAttributeServiceImpl struct {
	userAttributeDB db.UserAttributeDB
	userDB          db.UserDB
}

// NewUserAttributeService creates a new UserAttributeService instance
func NewUserAttributeService(userAttributeDB db.UserAttributeDB, userDB db.UserDB) UserAttributeService {
	return &userAttributeServiceImpl{
		userAttributeDB: userAttributeDB,
		userDB:          userDB,
	}
}

func (s *userAttributeServiceImpl) ListUserAttributes(ctx context.Context, tenant, realm, userID string) ([]model.UserAttribute, error) {
	// Verify user exists
	user, err := s.userDB.GetUserByID(ctx, tenant, realm, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // User not found
	}

	return s.userAttributeDB.ListUserAttributes(ctx, tenant, realm, userID)
}

func (s *userAttributeServiceImpl) GetUserAttributeByID(ctx context.Context, tenant, realm, attributeID string) (*model.UserAttribute, error) {
	return s.userAttributeDB.GetUserAttributeByID(ctx, tenant, realm, attributeID)
}

func (s *userAttributeServiceImpl) CreateUserAttribute(ctx context.Context, attribute model.UserAttribute) (*model.UserAttribute, error) {
	// Verify user exists
	user, err := s.userDB.GetUserByID(ctx, attribute.Tenant, attribute.Realm, attribute.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // User not found
	}

	// Generate UUID for the new attribute
	if attribute.ID == "" {
		attribute.ID = uuid.NewString()
	}

	// Create the attribute
	err = s.userAttributeDB.CreateUserAttribute(ctx, attribute)
	if err != nil {
		return nil, err
	}

	// Return the created attribute (with ID populated)
	return s.userAttributeDB.GetUserAttributeByID(ctx, attribute.Tenant, attribute.Realm, attribute.ID)
}

func (s *userAttributeServiceImpl) UpdateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error {
	// Verify attribute exists
	existing, err := s.userAttributeDB.GetUserAttributeByID(ctx, attribute.Tenant, attribute.Realm, attribute.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil // Attribute not found
	}

	return s.userAttributeDB.UpdateUserAttribute(ctx, attribute)
}

func (s *userAttributeServiceImpl) DeleteUserAttribute(ctx context.Context, tenant, realm, attributeID string) error {
	return s.userAttributeDB.DeleteUserAttribute(ctx, tenant, realm, attributeID)
}
