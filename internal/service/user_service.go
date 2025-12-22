package service

import (
	"context"

	"github.com/Identityplane/GoAM/pkg/db"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
	"github.com/google/uuid"
)

// userServiceImpl implements UserService
type userServiceImpl struct {
	userDB       db.UserDB
	attributesDB db.UserAttributeDB
}

// NewUserService creates a new UserService instance
func NewUserService(userDB db.UserDB, attributesDB db.UserAttributeDB) services_interface.UserAdminService {
	return &userServiceImpl{
		userDB:       userDB,
		attributesDB: attributesDB,
	}
}

func (s *userServiceImpl) ListUsers(ctx context.Context, tenant, realm string, pagination services_interface.PaginationParams) ([]model.User, int64, error) {
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

func (s *userServiceImpl) GetUserWithAttributesByID(ctx context.Context, tenant, realm, userID string) (*model.User, error) {

	user, err := s.attributesDB.GetUserWithAttributes(ctx, tenant, realm, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return user, nil
}

func (s *userServiceImpl) GetUserByID(ctx context.Context, tenant, realm, userID string) (*model.User, error) {
	return s.userDB.GetUserByID(ctx, tenant, realm, userID)
}

func (s *userServiceImpl) UpdateUserByID(ctx context.Context, tenant, realm, userID string, updateUser model.User) (*model.User, error) {
	// Get existing user
	user, err := s.userDB.GetUserByID(ctx, tenant, realm, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil // User not found
	}

	// Update user fields
	user.Status = updateUser.Status

	// Update user in database
	if err := s.userDB.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userServiceImpl) DeleteUserByID(ctx context.Context, tenant, realm, userID string) error {
	// Check if user exists
	user, err := s.userDB.GetUserByID(ctx, tenant, realm, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return nil // User not found, but that's fine for idempotency
	}

	// Delete the user
	return s.userDB.DeleteUser(ctx, tenant, realm, userID)
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
	existing, err := s.userDB.GetUserByID(ctx, tenant, realm, createUser.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, nil // User already exists
	}

	// Set required fields
	if createUser.ID == "" {
		createUser.ID = uuid.NewString()
	}
	createUser.Tenant = tenant
	createUser.Realm = realm

	// Create user in database
	err = s.userDB.CreateUser(ctx, createUser)
	if err != nil {
		return nil, err
	}

	// Return the created user (now with ID)
	return &createUser, nil
}

func (s *userServiceImpl) CreateUserWithAttributes(ctx context.Context, tenant, realm string, user model.User) (*model.User, error) {

	// No check if a user with this id already exists

	// If the user has no id we set a new uuid
	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	// If the user has no status we set it to active
	if user.Status == "" {
		user.Status = "active"
	}

	// Ensure realm and tenant are matching
	user.Tenant = tenant
	user.Realm = realm

	// Create user in database with all attributes
	err := s.attributesDB.CreateUserWithAttributes(ctx, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userServiceImpl) UpdateUserWithAttributes(ctx context.Context, tenant, realm string, user model.User) (*model.User, error) {

	// No check if a user with this id already exists

	// Ensure realm and tenant are matching
	user.Tenant = tenant
	user.Realm = realm

	// Update user in database with all attributes
	err := s.attributesDB.UpdateUserWithAttributes(ctx, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userServiceImpl) CreateOrUpdateUserWithAttributes(ctx context.Context, tenant, realm string, user model.User) (*model.User, error) {

	// check if the user already exists
	existing := false

	// Ensure realm and tenant are matching
	user.Tenant = tenant
	user.Realm = realm

	// If we have an id we check if the user is already existing
	if user.ID != "" {
		existing_user, err := s.userDB.GetUserByID(ctx, tenant, realm, user.ID)
		if err != nil {
			return nil, err
		}

		if existing_user != nil {
			existing = true
		}
	}

	if existing {
		return s.UpdateUserWithAttributes(ctx, tenant, realm, user)
	} else {
		return s.CreateUserWithAttributes(ctx, tenant, realm, user)
	}
}
