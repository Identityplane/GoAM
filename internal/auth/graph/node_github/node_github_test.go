package node_github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGithubLoginAndCreateUserNodes(t *testing.T) {
	// Create mock GitHub API server
	githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/login/oauth/access_token":
			// Verify request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Failed to read request body: %v", err)
			}

			// Parse the request body
			var requestBody map[string]string
			if err := json.Unmarshal(body, &requestBody); err != nil {
				t.Fatalf("Failed to parse request body: %v", err)
			}

			// Verify required fields
			if requestBody["client_id"] != "test_client_id" ||
				requestBody["client_secret"] != "test_client_secret" ||
				requestBody["code"] != "test_code" {
				t.Fatalf("Invalid request body: %s", string(body))
			}

			// Mock token response
			response := githubAccessTokenResponse{
				AccessToken:  "mock_access_token",
				RefreshToken: "mock_refresh_token",
				TokenType:    "bearer",
				Scope:        "user",
			}
			json.NewEncoder(w).Encode(response)
		case "/user":
			// Verify Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "token mock_access_token" {
				t.Fatalf("Invalid Authorization header: %s", authHeader)
			}

			// Mock user data response
			response := GitHubUser{
				Login:     "testuser",
				ID:        12345,
				AvatarURL: "https://github.com/avatar.png",
				Email:     "test@example.com",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer githubServer.Close()

	// Override GitHub API URLs for testing
	originalTokenURL := githubTokenURL
	originalUserURL := githubUserURL
	githubTokenURL = githubServer.URL + "/login/oauth/access_token"
	githubUserURL = githubServer.URL + "/user"
	defer func() {
		githubTokenURL = originalTokenURL
		githubUserURL = originalUserURL
	}()

	// Create test user
	testUser := &model.User{
		ID:     uuid.NewString(),
		Status: "active",
	}

	// Create mock repository
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			"github-client-id":     "test_client_id",
			"github-client-secret": "test_client_secret",
			"github-scope":         "user",
		},
	}

	// Test 1: Initial state - should return redirect URL
	session := &model.AuthenticationSession{
		LoginUri: "http://localhost:8080/callback",
		Context:  make(map[string]string),
	}
	result, err := RunGithubLoginNode(session, node, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Contains(t, result.Prompts["__redirect"], "github.com/login/oauth/authorize")

	// Test 2: With code - existing user
	mockUserRepo.ExpectedCalls = nil // Reset mock expectations
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeGitHub, "12345").Return(testUser, nil)
	session = &model.AuthenticationSession{
		LoginUri: "http://localhost:8080/callback",
		Context:  make(map[string]string),
	}
	result, err = RunGithubLoginNode(session, node, map[string]string{"code": "test_code"}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "existing-user", result.Condition)
	assert.Equal(t, testUser, session.User)

	// Test 3: With code - new user
	mockUserRepo.ExpectedCalls = nil // Reset mock expectations
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeGitHub, "12345").Return(nil, nil)
	session = &model.AuthenticationSession{
		LoginUri: "http://localhost:8080/callback",
		Context:  make(map[string]string),
	}
	result, err = RunGithubLoginNode(session, node, map[string]string{"code": "test_code"}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-user", result.Condition)

	// Test 4: Create new user
	mockUserRepo.ExpectedCalls = nil // Reset mock expectations
	mockUserRepo.On("CreateOrUpdate", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
	result, err = RunGithubCreateUserNode(session, node, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "created", result.Condition)
	assert.NotNil(t, session.User)

	// Verify that the GitHub attribute was created correctly
	assert.Len(t, session.User.UserAttributes, 1)
	githubAttr := session.User.UserAttributes[0]
	assert.Equal(t, model.AttributeTypeGitHub, githubAttr.Type)
	assert.Equal(t, "12345", githubAttr.Index) // Should use GitHub User ID as index

	// Verify the GitHub attribute value
	githubValue, ok := githubAttr.Value.(model.GitHubAttributeValue)
	assert.True(t, ok, "GitHub attribute value should be of type GitHubAttributeValue")
	assert.Equal(t, "12345", githubValue.GitHubUserID)
	assert.Equal(t, "testuser", githubValue.GitHubUsername)
	assert.Equal(t, "test@example.com", githubValue.GitHubEmail)
	assert.Equal(t, "https://github.com/avatar.png", githubValue.GitHubAvatarURL)
	assert.Equal(t, "mock_access_token", githubValue.GitHubAccessToken)
	assert.Equal(t, "mock_refresh_token", githubValue.GitHubRefreshToken)
	assert.Equal(t, "bearer", githubValue.GitHubTokenType)
	assert.Equal(t, "user", githubValue.GitHubScope)

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
}

func TestGithubCreateUserNode_WithoutGitHubContext(t *testing.T) {
	// Test that the node fails when no GitHub context is provided
	session := &model.AuthenticationSession{
		Context: make(map[string]string), // No github context
	}

	node := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	services := &model.Repositories{}

	result, err := RunGithubCreateUserNode(session, node, map[string]string{}, services)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "github is required in state context")
}

func TestGithubCreateUserNode_WithInvalidGitHubContext(t *testing.T) {
	// Test that the node fails when GitHub context is invalid JSON
	session := &model.AuthenticationSession{
		Context: map[string]string{
			"github": "invalid-json",
		},
	}

	node := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	services := &model.Repositories{}

	result, err := RunGithubCreateUserNode(session, node, map[string]string{}, services)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse github attribute")
}

func TestGithubCreateUserNode_WithValidGitHubContext(t *testing.T) {
	// Test that the node creates a user with GitHub attributes correctly
	githubAttributeValue := model.GitHubAttributeValue{
		GitHubUserID:       "12345",
		GitHubRefreshToken: "refresh_token",
		GitHubEmail:        "test@example.com",
		GitHubAvatarURL:    "https://github.com/avatar.png",
		GitHubUsername:     "testuser",
		GitHubAccessToken:  "access_token",
		GitHubTokenType:    "bearer",
		GitHubScope:        "user",
	}

	// Marshal the GitHub attribute to JSON
	githubAttributeJSON, err := json.Marshal(githubAttributeValue)
	assert.NoError(t, err)

	session := &model.AuthenticationSession{
		Context: map[string]string{
			"github": string(githubAttributeJSON),
		},
		User: nil, // No existing user
	}

	node := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	// Create mock repository
	mockUserRepo := repository.NewMockUserRepository()
	mockUserRepo.On("CreateOrUpdate", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	result, err := RunGithubCreateUserNode(session, node, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "created", result.Condition)

	// Verify that a new user was created
	assert.NotNil(t, session.User)
	assert.Equal(t, "active", session.User.Status)

	// Verify that the GitHub attribute was added correctly
	assert.Len(t, session.User.UserAttributes, 1)
	githubAttr := session.User.UserAttributes[0]
	assert.Equal(t, model.AttributeTypeGitHub, githubAttr.Type)
	assert.Equal(t, "12345", githubAttr.Index)

	// Verify the attribute value
	githubValue, ok := githubAttr.Value.(model.GitHubAttributeValue)
	assert.True(t, ok)
	assert.Equal(t, "12345", githubValue.GitHubUserID)
	assert.Equal(t, "testuser", githubValue.GitHubUsername)
	assert.Equal(t, "test@example.com", githubValue.GitHubEmail)

	// Verify mock expectations
	mockUserRepo.AssertExpectations(t)
}

func TestGithubCreateUserNode_WithExistingUser(t *testing.T) {
	// Test that the node adds GitHub attributes to an existing user
	existingUser := &model.User{
		ID:     uuid.NewString(),
		Status: "active",
		UserAttributes: []model.UserAttribute{
			{
				ID:    uuid.NewString(),
				Type:  "email",
				Index: "existing@example.com",
				Value: "existing@example.com",
			},
		},
	}

	githubAttributeValue := model.GitHubAttributeValue{
		GitHubUserID:       "67890",
		GitHubRefreshToken: "refresh_token_2",
		GitHubEmail:        "github@example.com",
		GitHubAvatarURL:    "https://github.com/avatar2.png",
		GitHubUsername:     "githubuser",
		GitHubAccessToken:  "access_token_2",
		GitHubTokenType:    "bearer",
		GitHubScope:        "user:email",
	}

	// Marshal the GitHub attribute to JSON
	githubAttributeJSON, err := json.Marshal(githubAttributeValue)
	assert.NoError(t, err)

	session := &model.AuthenticationSession{
		Context: map[string]string{
			"github": string(githubAttributeJSON),
		},
		User: existingUser, // Existing user
	}

	node := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	// Create test repository with real SQLite database for integration testing
	testUserRepo, err := repository.NewTestUserRepository("acme", "customers")
	assert.NoError(t, err)
	defer testUserRepo.Close()

	services := &model.Repositories{
		UserRepo: testUserRepo,
	}

	result, err := RunGithubCreateUserNode(session, node, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "created", result.Condition)

	// Verify that the existing user was preserved
	assert.Equal(t, existingUser.ID, session.User.ID)
	assert.Equal(t, "active", session.User.Status)

	// Verify that the GitHub attribute was added to existing attributes
	assert.Len(t, session.User.UserAttributes, 2)

	// Find the GitHub attribute
	var githubAttr *model.UserAttribute
	for _, attr := range session.User.UserAttributes {
		if attr.Type == model.AttributeTypeGitHub {
			githubAttr = &attr
			break
		}
	}

	assert.NotNil(t, githubAttr, "GitHub attribute should be present")
	assert.Equal(t, model.AttributeTypeGitHub, githubAttr.Type)
	assert.Equal(t, "67890", githubAttr.Index)

	// Verify the attribute value
	githubValue, ok := githubAttr.Value.(model.GitHubAttributeValue)
	assert.True(t, ok)
	assert.Equal(t, "67890", githubValue.GitHubUserID)
	assert.Equal(t, "githubuser", githubValue.GitHubUsername)
	assert.Equal(t, "github@example.com", githubValue.GitHubEmail)

	// Verify that the user was actually persisted to the database by loading it
	// This tests that CreateOrUpdate actually worked
	persistedUser, err := testUserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeGitHub, "67890")
	assert.NoError(t, err)
	assert.NotNil(t, persistedUser, "User should be retrievable from database after creation")

	// Verify the persisted user has the correct attributes
	assert.Len(t, persistedUser.UserAttributes, 2)

	githubAttrVal, attr, err := model.GetAttribute[model.GitHubAttributeValue](persistedUser, model.AttributeTypeGitHub)
	assert.NoError(t, err)
	assert.NotNil(t, githubAttrVal, "GitHub attribute value should be present in persisted user")
	assert.NotNil(t, attr, "GitHub attribute should be present in persisted user")

	assert.Equal(t, "67890", githubAttrVal.GitHubUserID)
	assert.Equal(t, "githubuser", githubAttrVal.GitHubUsername)
	assert.Equal(t, "github@example.com", githubAttrVal.GitHubEmail)
	assert.Equal(t, "https://github.com/avatar2.png", githubAttrVal.GitHubAvatarURL)
	assert.Equal(t, "access_token_2", githubAttrVal.GitHubAccessToken)
	assert.Equal(t, "refresh_token_2", githubAttrVal.GitHubRefreshToken)
	assert.Equal(t, "bearer", githubAttrVal.GitHubTokenType)
	assert.Equal(t, "user:email", githubAttrVal.GitHubScope)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
