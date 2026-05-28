package model

import (
	"context"
)

type Repositories struct {
	UserRepo    UserRepository
	EmailSender EmailSender
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByAttributeIndex(ctx context.Context, attributeType, index string) (*User, error)

	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	CreateOrUpdate(ctx context.Context, user *User) error

	CreateUserAttribute(ctx context.Context, attribute *UserAttribute) error
	UpdateUserAttribute(ctx context.Context, attribute *UserAttribute) error
	DeleteUserAttribute(ctx context.Context, attributeID string) error

	// Creates a new user model based on the context value
	// This initializes the user according to the realm requirements with id, and state
	NewUserModel(state *AuthenticationSession) (*User, error)
}
