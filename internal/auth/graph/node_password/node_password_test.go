package node_password

import (
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdatePasswordNode(t *testing.T) {
	// Create test user with password attribute
	testUser := &model.User{
		ID:        uuid.NewString(),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add password attribute
	passwordAttrValue := model.PasswordAttributeValue{
		PasswordHash:         "old_hash",
		Locked:               false,
		FailedAttempts:       0,
		LastCorrectTimestamp: nil, // No previous correct login
	}
	testUser.AddAttribute(&model.UserAttribute{
		ID:    uuid.NewString(),
		Type:  model.AttributeTypePassword,
		Value: passwordAttrValue,
	})

	// Create mock repository
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test session
	session := &model.AuthenticationSession{
		Context: map[string]string{},
		User:    testUser,
	}

	// Test update password
	updateNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	newPassword := "newSecurePassword123!"
	mockUserRepo.On("Update", mock.Anything, testUser).Return(nil)

	result, err := RunUpdatePasswordNode(session, updateNode, map[string]string{"password": newPassword}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Condition)

	// Verify password attribute was updated in the session user
	passwordAttr, _, err := model.GetAttribute[model.PasswordAttributeValue](session.User, model.AttributeTypePassword)
	assert.NoError(t, err)
	assert.NotNil(t, passwordAttr)
	assert.NotEqual(t, "old_hash", passwordAttr.PasswordHash)
	assert.False(t, passwordAttr.Locked)
	assert.Equal(t, 0, passwordAttr.FailedAttempts)
	assert.NotNil(t, passwordAttr.LastCorrectTimestamp, "LastCorrectTimestamp should be set when password is updated")

	// Verify mock expectations were met
	mockUserRepo.AssertExpectations(t)
}

func TestPasswordUpdateAndValidation(t *testing.T) {
	// Create test user with password attribute
	testUser := &model.User{
		ID:        uuid.NewString(),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add password attribute
	passwordAttrValue := model.PasswordAttributeValue{
		PasswordHash:         "old_hash",
		Locked:               false,
		FailedAttempts:       0,
		LastCorrectTimestamp: nil, // No previous correct login
	}
	testUser.AddAttribute(&model.UserAttribute{
		ID:    uuid.NewString(),
		Type:  model.AttributeTypePassword,
		Value: passwordAttrValue,
	})

	// Add username attribute
	usernameAttrValue := model.UsernameAttributeValue{
		PreferredUsername: "testuser",
	}
	testUser.AddAttribute(&model.UserAttribute{
		ID:    uuid.NewString(),
		Type:  model.AttributeTypeUsername,
		Index: lib.StringPtr("testuser"),
		Value: usernameAttrValue,
	})

	// Create mock repository
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test session
	session := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "testuser",
			"user_id":  testUser.ID, // Add user_id for LoadUserFromContext
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

	// Verify password attribute was updated in the session user
	passwordAttr, _, err := model.GetAttribute[model.PasswordAttributeValue](session.User, model.AttributeTypePassword)
	assert.NoError(t, err)
	assert.NotNil(t, passwordAttr)
	assert.NotEqual(t, "old_hash", passwordAttr.PasswordHash)
	assert.False(t, passwordAttr.Locked)
	assert.Equal(t, 0, passwordAttr.FailedAttempts)
	assert.NotNil(t, passwordAttr.LastCorrectTimestamp, "LastCorrectTimestamp should be set when password is updated")

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
}
