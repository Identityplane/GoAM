package node_passkeys

import (
	"fmt"
	"testing"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunCheckUserHasPasskeyNode_UserHasPasskey(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create a test user with a passkey
	testUser := &model.User{
		ID:             "test-user-123",
		Tenant:         "acme",
		Realm:          "customers",
		Status:         "active",
		UserAttributes: []model.UserAttribute{},
	}

	// Add a passkey attribute
	testUser.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypePasskey,
		Index: "passkey-1",
		Value: model.PasskeyAttributeValue{
			RPID:         "localhost",
			DisplayName:  "alice",
			CredentialID: "passkey-1",
		},
	})

	// Create authentication session with user context
	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "alice",
		},
		User: testUser, // Set user directly to avoid database lookup
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "checkPasskeyRegistered",
		Use:  "checkPasskeyRegistered",
	}

	// Run the node
	result, err := RunCheckUserHasPasskeyNode(state, node, map[string]string{}, mockRepos)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "registered", result.Condition, "User with passkey should return 'registered'")
}

func TestRunCheckUserHasPasskeyNode_UserNoPasskey(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create a test user without passkeys
	testUser := &model.User{
		ID:             "test-user-123",
		Tenant:         "acme",
		Realm:          "customers",
		Status:         "active",
		UserAttributes: []model.UserAttribute{},
	}

	// Add a username attribute but no passkey
	testUser.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypeUsername,
		Index: "alice",
		Value: model.UsernameAttributeValue{
			Username: "alice",
		},
	})

	// Create authentication session with user context
	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "alice",
		},
		User: testUser, // Set user directly to avoid database lookup
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "checkPasskeyRegistered",
		Use:  "checkPasskeyRegistered",
	}

	// Run the node
	result, err := RunCheckUserHasPasskeyNode(state, node, map[string]string{}, mockRepos)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "not_registered", result.Condition, "User without passkey should return 'not_registered'")
}

func TestRunCheckUserHasPasskeyNode_UserNotFound(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Setup mock to return error when trying to find user
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeUsername, "nonexistent-user").Return(nil, fmt.Errorf("user not found"))

	// Create authentication session with username but no user loaded
	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "nonexistent-user",
		},
		User: nil, // No user loaded
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "checkPasskeyRegistered",
		Use:  "checkPasskeyRegistered",
	}

	// Run the node
	result, err := RunCheckUserHasPasskeyNode(state, node, map[string]string{}, mockRepos)

	// Verify result
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot load user")
	assert.Nil(t, result)
}

func TestRunCheckUserHasPasskeyNode_UserWithMultiplePasskeys(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create a test user with multiple passkeys
	testUser := &model.User{
		ID:             "test-user-123",
		Tenant:         "acme",
		Realm:          "customers",
		Status:         "active",
		UserAttributes: []model.UserAttribute{},
	}

	// Add multiple passkey attributes
	testUser.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypePasskey,
		Index: "passkey-1",
		Value: model.PasskeyAttributeValue{
			RPID:         "localhost",
			DisplayName:  "alice",
			CredentialID: "passkey-1",
		},
	})

	testUser.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypePasskey,
		Index: "passkey-2",
		Value: model.PasskeyAttributeValue{
			RPID:         "localhost",
			DisplayName:  "alice",
			CredentialID: "passkey-2",
		},
	})

	// Create authentication session with user context
	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "alice",
		},
		User: testUser, // Set user directly to avoid database lookup
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "checkPasskeyRegistered",
		Use:  "checkPasskeyRegistered",
	}

	// Run the node
	result, err := RunCheckUserHasPasskeyNode(state, node, map[string]string{}, mockRepos)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "registered", result.Condition, "User with multiple passkeys should return 'registered'")
}
