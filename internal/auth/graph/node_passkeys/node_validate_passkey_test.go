package node_passkeys

import (
	"encoding/json"
	"testing"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunPasskeyVerifyNode_GenerateOptions(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create authentication session
	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "alice",
		},
		LoginUriBase: "https://localhost:8080/acme/customers/auth/login",
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "verifyPasskey",
		Use:  "verifyPasskey",
	}

	// Run the node without input (should generate login options)
	result, err := RunPasskeyVerifyNode(state, node, map[string]string{}, mockRepos)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Contains(t, result.Prompts, "passkeysLoginOptions")

	// Check that session and options are stored in context
	sessionJSON, sessionExists := state.Context["passkeysSession"]
	optionsJSON, optionsExist := state.Context["passkeysOptions"]

	assert.True(t, sessionExists, "passkeysSession should be set in context")
	assert.True(t, optionsExist, "passkeysOptions should be set in context")
	assert.NotEmpty(t, sessionJSON)
	assert.NotEmpty(t, optionsJSON)

	// Verify the prompts contain the expected keys
	assert.NotEmpty(t, result.Prompts["passkeysLoginOptions"])
}

func TestRunPasskeyVerifyNode_WrongCredentials(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create authentication session
	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "alice",
		},
		LoginUriBase: "https://localhost:8080/acme/customers/auth/login",
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "verifyPasskey",
		Use:  "verifyPasskey",
	}

	// Step 1: Run the node without input (should generate login options)
	result1, err := RunPasskeyVerifyNode(state, node, map[string]string{}, mockRepos)

	// Verify first run generates prompts
	assert.NoError(t, err)
	assert.NotNil(t, result1)
	assert.NotEmpty(t, result1.Prompts)

	// Step 2: Run with invalid credential response
	input := map[string]string{
		"passkeysFinishLoginJson": "invalid-json",
	}

	result2, err := RunPasskeyVerifyNode(state, node, input, mockRepos)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to process passkey login")
	assert.Nil(t, result2)
}

func TestRunPasskeyVerifyNode_UserNotFound(t *testing.T) {
	// Setup mock repositories
	mockUserRepo := repository.NewMockUserRepository()
	mockRepos := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Setup mock to return nil when trying to find user by credential ID
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypePasskey, "nonexistent-credential").Return(nil, nil)

	// Create authentication session
	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "alice",
		},
		LoginUriBase: "https://localhost:8080/acme/customers/auth/login",
	}

	// Create graph node
	node := &model.GraphNode{
		Name: "verifyPasskey",
		Use:  "verifyPasskey",
	}

	// First run to generate options
	result1, err := RunPasskeyVerifyNode(state, node, map[string]string{}, mockRepos)
	assert.NoError(t, err)
	assert.NotEmpty(t, result1.Prompts)

	// Create a mock credential response with non-existent credential ID
	mockResponse := map[string]interface{}{
		"id":   "nonexistent-credential",
		"type": "public-key",
		"response": map[string]interface{}{
			"authenticatorData": "mock-authenticator-data",
			"clientDataJSON":    "mock-client-data",
			"signature":         "mock-signature",
		},
	}

	responseJSON, err := json.Marshal(mockResponse)
	assert.NoError(t, err)

	// Run with non-existent credential
	input := map[string]string{
		"passkeysFinishLoginJson": string(responseJSON),
	}

	result2, err := RunPasskeyVerifyNode(state, node, input, mockRepos)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to process passkey login")
	assert.Nil(t, result2)
}
