package repository

import (
	"context"
	"database/sql"

	"github.com/Identityplane/GoAM/internal/db/sqlite_adapter"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository implements UserRepository for testing (pure mock)
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) NewUserModel(state *model.AuthenticationSession) (*model.User, error) {
	args := m.Called(state)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByAttributeIndex(ctx context.Context, attributeType, index string) (*model.User, error) {
	args := m.Called(ctx, attributeType, index)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) CreateOrUpdate(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) CreateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error {
	args := m.Called(ctx, attribute)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error {
	args := m.Called(ctx, attribute)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUserAttribute(ctx context.Context, attributeID string) error {
	args := m.Called(ctx, attributeID)
	return args.Error(0)
}

// TestUserRepository is a real SQLite-based repository for integration testing
type TestUserRepository struct {
	*UserRepositoryImpl
	db *sql.DB
}

// NewTestUserRepository creates a new test repository using SQLite in-memory database
func NewTestUserRepository(tenant, realm string) (*TestUserRepository, error) {
	// Create SQLite in-memory database
	sqliteDB, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	// Run migrations to create tables
	err = sqlite_adapter.RunMigrations(sqliteDB)
	if err != nil {
		sqliteDB.Close()
		return nil, err
	}

	// Create database adapters
	userDB, err := sqlite_adapter.NewUserDB(sqliteDB)
	if err != nil {
		sqliteDB.Close()
		return nil, err
	}

	userAttributeDB, err := sqlite_adapter.NewUserAttributeDB(sqliteDB)
	if err != nil {
		sqliteDB.Close()
		return nil, err
	}

	// Create the repository
	repo := &TestUserRepository{
		UserRepositoryImpl: &UserRepositoryImpl{
			tenant:       tenant,
			realm:        realm,
			db:           userDB,
			attributesDB: userAttributeDB,
		},
		db: sqliteDB,
	}

	return repo, nil
}

// Close closes the underlying database connection
func (t *TestUserRepository) Close() error {
	return t.db.Close()
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {

	mockUserRepo := new(MockUserRepository)
	mockUserRepo.On("NewUserModel", mock.Anything).Maybe().Return(&model.User{
		ID:     uuid.NewString(),
		Tenant: "acme",
		Realm:  "customers",
		Status: "active",
	}, nil)
	return mockUserRepo
}
