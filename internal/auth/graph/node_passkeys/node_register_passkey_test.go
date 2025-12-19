package node_passkeys

import (
	"encoding/json"
	"testing"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunPasskeyRegisterNode_GenerateOptions(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create a test user
	testUser := &model.User{
		ID:             "test-user-123",
		Tenant:         "acme",
		Realm:          "customers",
		Status:         "active",
		UserAttributes: []*model.UserAttribute{},
	}

	// Add username attribute using the proper method
	testUser.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypeUsername,
		Index: lib.StringPtr("alice"),
		Value: model.UsernameAttributeValue{
			PreferredUsername: "alice",
		},
	})

	// Create authentication session with user context
	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "alice",
		},
		User:         testUser, // Set user directly to avoid database lookup
		LoginUriBase: "https://localhost:8080/acme/customers/auth/login",
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "registerPasskey",
		Use:  "registerPasskey",
	}

	// Step 1: Run the node without input (should generate options)
	result, err := RunPasskeyRegisterNode(state, node, map[string]string{}, mockRepos)

	// Verify first run generates prompts
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)

	// Check that session and options are stored in context
	sessionJSON, sessionExists := state.Context["passkeysSession"]
	optionsJSON, optionsExist := state.Context["passkeysOptions"]

	assert.True(t, sessionExists, "passkeysSession should be set in context")
	assert.True(t, optionsExist, "passkeysOptions should be set in context")
	assert.NotEmpty(t, sessionJSON)
	assert.NotEmpty(t, optionsJSON)

	// Verify the prompts contain the expected keys
	assert.Contains(t, result.Prompts, "passkeysOptions")
	assert.NotEmpty(t, result.Prompts["passkeysOptions"])

	// Parse the session data to verify it's valid JSON
	var session webauthn.SessionData
	err = json.Unmarshal([]byte(sessionJSON), &session)
	assert.NoError(t, err, "passkeysSession should be valid JSON")

	// Parse the options to verify it's valid JSON
	var options protocol.CredentialCreation
	err = json.Unmarshal([]byte(optionsJSON), &options)
	assert.NoError(t, err, "passkeysOptions should be valid JSON")

	// Verify the options contain expected fields
	assert.NotEmpty(t, options.Response.Challenge, "Challenge should be present")
	assert.Equal(t, "localhost", options.Response.RelyingParty.ID, "RP ID should match")

	// Verify the username and display name are correctly set
	assert.Equal(t, "alice", options.Response.User.Name, "Username should match the account name from context")
	assert.Equal(t, "alice", options.Response.User.DisplayName, "Display name should match the account name from context")
}

func TestRunPasskeyRegisterNode_UserNotInContext(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create authentication session without user context
	state := &model.AuthenticationSession{
		Context:      map[string]string{}, // No user identifier in context
		User:         nil,                 // No user loaded
		LoginUriBase: "https://localhost:8080/acme/customers/auth/login",
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "registerPasskey",
		Use:  "registerPasskey",
	}

	// Run the node - should fail because user is not in context
	result, err := RunPasskeyRegisterNode(state, node, map[string]string{}, mockRepos)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user must be loaded before registering a passkey")
	assert.Nil(t, result)
}

func TestRunPasskeyRegisterNode_InvalidCredentialResponse(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create a test user
	testUser := &model.User{
		ID:             "test-user-123",
		Tenant:         "acme",
		Realm:          "customers",
		Status:         "active",
		UserAttributes: []*model.UserAttribute{},
	}

	// Add username attribute using the proper method
	testUser.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypeUsername,
		Index: lib.StringPtr("alice"),
		Value: model.UsernameAttributeValue{
			PreferredUsername: "alice",
		},
	})

	// Setup mock expectations for user loading
	mockUserRepo.On("GetByID", mock.Anything, "test-user-123").Return(testUser, nil)

	// Create authentication session with user context
	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "alice",
		},
		User:         testUser, // Set user directly to avoid database lookup
		LoginUriBase: "https://localhost:8080/acme/customers/auth/login",
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "registerPasskey",
		Use:  "registerPasskey",
	}

	// First run to generate options
	result1, err := RunPasskeyRegisterNode(state, node, map[string]string{}, mockRepos)
	assert.NoError(t, err)
	assert.NotEmpty(t, result1.Prompts)

	// Run with invalid credential response
	input := map[string]string{
		"passkeysFinishRegistrationJson": "invalid-json",
	}

	result2, err := RunPasskeyRegisterNode(state, node, input, mockRepos)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to process passkey registration")
	assert.Nil(t, result2)
}
