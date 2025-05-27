package graph

import (
	"context"
	"errors"
	"goiam/internal/auth/repository"
	"goiam/internal/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of the user repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByLoginIdentifier(ctx context.Context, identifier string) (*model.User, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByFederatedIdentifier(ctx context.Context, provider, identifier string) (*model.User, error) {
	args := m.Called(ctx, provider, identifier)
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

func TestRunLoadUserNode(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		mockUser       *model.User
		mockError      error
		expectedResult string
		expectError    bool
	}{
		{
			name:     "User found successfully",
			username: "testuser",
			mockUser: &model.User{
				ID:        uuid.NewString(),
				Username:  "testuser",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockError:      nil,
			expectedResult: "loaded",
			expectError:    false,
		},
		{
			name:           "User not found",
			username:       "nonexistent",
			mockUser:       nil,
			mockError:      nil,
			expectedResult: "not_found",
			expectError:    false,
		},
		{
			name:           "Repository error",
			username:       "testuser",
			mockUser:       nil,
			mockError:      errors.New("database error"),
			expectedResult: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockUserRepository)
			mockRepo.On("GetByUsername", mock.Anything, tt.username).Return(tt.mockUser, tt.mockError)

			// Create services with mock repository
			services := &repository.Repositories{
				UserRepo: mockRepo,
			}

			// Create test state
			state := &model.AuthenticationSession{
				Context: map[string]string{
					"username": tt.username,
				},
			}

			// Create minimal node (not used in logic but required)
			node := &model.GraphNode{
				Name: "loadUserByUsername",
				Use:  "loadUserByUsername",
			}

			// Run the node
			result, err := RunLoadUserNode(state, node, nil, services)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult, result.Condition)

				if tt.mockUser != nil {
					// Verify state was updated correctly
					assert.Equal(t, tt.mockUser, state.User)
					assert.Equal(t, tt.mockUser.ID, state.Context["user_id"])
					assert.Equal(t, tt.mockUser.Username, state.Context["username"])
				}
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}
