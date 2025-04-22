package service

import (
	"context"

	"goiam/internal/db"
	"goiam/internal/model"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`      // 1-based page number
	PageSize int `json:"page_size"` // Number of items per page
}

// UserAdminService defines the business logic for user operations
type UserAdminService interface {
	// List users with pagination, returns usersn, total count and users
	ListUsers(ctx context.Context, tenant, realm string, pagination PaginationParams) ([]model.User, int64, error)
	GetUser(ctx context.Context, tenant, realm, username string) (*model.User, error)
	UpdateUser(ctx context.Context, tenant, realm, username string, updateUser model.User) (*model.User, error)
	DeleteUser(ctx context.Context, tenant, realm, username string) error
	// Get user statistics
	GetUserStats(ctx context.Context, tenant, realm string) (*model.UserStats, error)
	// Create a new user
	CreateUser(ctx context.Context, tenant, realm string, createUser model.User) (*model.User, error)
}

// userServiceImpl implements UserService
type userServiceImpl struct {
	userDB db.UserDB
}

// NewUserService creates a new UserService instance
func NewUserService(userDB db.UserDB) UserAdminService {
	return &userServiceImpl{
		userDB: userDB,
	}
}

func (s *userServiceImpl) ListUsers(ctx context.Context, tenant, realm string, pagination PaginationParams) ([]model.User, int64, error) {
	// Get total count first
	total, err := s.userDB.CountUsers(ctx, tenant, realm)
	if err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (pagination.Page - 1) * pagination.PageSize

	// Get paginated users
	users, err := s.userDB.ListUsersWithPagination(ctx, tenant, realm, offset, pagination.PageSize)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *userServiceImpl) GetUser(ctx context.Context, tenant, realm, username string) (*model.User, error) {
	return s.userDB.GetUserByUsername(ctx, tenant, realm, username)
}

func (s *userServiceImpl) UpdateUser(ctx context.Context, tenant, realm, username string, updateUser model.User) (*model.User, error) {
	// Get existing user
	user, err := s.userDB.GetUserByUsername(ctx, tenant, realm, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil // User not found
	}

	// Update user fields
	user.DisplayName = updateUser.DisplayName
	user.GivenName = updateUser.GivenName
	user.FamilyName = updateUser.FamilyName
	user.Email = updateUser.Email
	user.Phone = updateUser.Phone
	user.EmailVerified = updateUser.EmailVerified
	user.PhoneVerified = updateUser.PhoneVerified
	user.Locale = updateUser.Locale
	user.Status = updateUser.Status
	user.Roles = updateUser.Roles
	user.Groups = updateUser.Groups
	user.Attributes = updateUser.Attributes

	// Update user in database
	if err := s.userDB.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userServiceImpl) DeleteUser(ctx context.Context, tenant, realm, username string) error {
	// Check if user exists
	user, err := s.userDB.GetUserByUsername(ctx, tenant, realm, username)
	if err != nil {
		return err
	}

	if user == nil {
		return nil // User not found, but that's fine for idempotency
	}

	// Delete the user
	return s.userDB.DeleteUser(ctx, tenant, realm, username)
}

func (s *userServiceImpl) GetUserStats(ctx context.Context, tenant, realm string) (*model.UserStats, error) {

	stats, err := s.userDB.GetUserStats(ctx, tenant, realm)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *userServiceImpl) CreateUser(ctx context.Context, tenant, realm string, createUser model.User) (*model.User, error) {
	// Check if user already exists
	existing, err := s.userDB.GetUserByUsername(ctx, tenant, realm, createUser.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, nil // User already exists
	}

	// Set required fields
	createUser.Tenant = tenant
	createUser.Realm = realm
	createUser.Status = "active" // Default status

	// Create user in database
	err = s.userDB.CreateUser(ctx, createUser)
	if err != nil {
		return nil, err
	}

	// Return the created user
	return &createUser, nil
}
