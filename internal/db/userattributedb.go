package db

import (
	"context"

	"github.com/Identityplane/GoAM/pkg/model"
)

// Interface for the user attribute db
type UserAttributeDB interface {
	// Create a new user attribute
	CreateUserAttribute(ctx context.Context, attribute model.UserAttribute) error

	// Get all attributes for a user (without detailed values)
	ListUserAttributes(ctx context.Context, tenant, realm, userID string) ([]model.UserAttribute, error)

	// Get a specific user attribute by ID with full details
	GetUserAttributeByID(ctx context.Context, tenant, realm, attributeID string) (*model.UserAttribute, error)

	// Update an existing user attribute
	UpdateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error

	// Delete a specific user attribute
	DeleteUserAttribute(ctx context.Context, tenant, realm, attributeID string) error

	// Get user by attribute index (for reverse lookup)
	GetUserByAttributeIndex(ctx context.Context, tenant, realm, attributeType, index string) (*model.User, error)
}
