package graph

import (
	"testing"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPasswordUpdateAndValidation(t *testing.T) {
	// Create test user
	testUser := &model.User{
		ID:                          uuid.NewString(),
		Username:                    "testuser",
		FailedLoginAttemptsPassword: 0,
		PasswordLocked:              false,
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
	}

	// Create mock repository
	mockUserRepo := new(MockUserRepository)
	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test session
	session := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "testuser",
		},
		User: testUser,
	}

	// Test 1: Update password
	updateNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	// First call should return password prompt
	result, err := RunUpdatePasswordNode(session, updateNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "password", result.Prompts["password"])

	// Update password
	newPassword := "newSecurePassword123!"
	mockUserRepo.On("Update", mock.Anything, testUser).Return(nil)
	result, err = RunUpdatePasswordNode(session, updateNode, map[string]string{"password": newPassword}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Condition)

	// Test 2: Validate password
	validateNode := &model.GraphNode{
		CustomConfig: map[string]string{
			"max_failed_password_attempts": "2",
		},
	}

	// Test correct password
	mockUserRepo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil)
	mockUserRepo.On("Update", mock.Anything, testUser).Return(nil)
	session.Context["password"] = newPassword
	result, err = RunValidateUsernamePasswordNode(session, validateNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Condition)
	assert.Equal(t, 0, testUser.FailedLoginAttemptsPassword) // Counter should be reset

	// Test first wrong password attempt
	session.Context["password"] = "wrongPassword"
	result, err = RunValidateUsernamePasswordNode(session, validateNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "fail", result.Condition)
	assert.Equal(t, 1, testUser.FailedLoginAttemptsPassword) // Counter should increase
	assert.False(t, testUser.PasswordLocked)                 // Account should not be locked yet

	// Test second wrong password attempt (should lock)
	result, err = RunValidateUsernamePasswordNode(session, validateNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "locked", result.Condition)
	assert.Equal(t, 2, testUser.FailedLoginAttemptsPassword) // Counter should be at max
	assert.True(t, testUser.PasswordLocked)                  // Account should be locked

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
}
