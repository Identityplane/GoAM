package graph

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gianlucafrei/GoAM/internal/auth/repository"
	"github.com/gianlucafrei/GoAM/internal/model"

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
		ID:           uuid.NewString(),
		Username:     "testuser",
		Email:        "test@example.com",
		FederatedIDP: stringPtr("github"),
		FederatedID:  stringPtr("12345"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create mock repository
	mockUserRepo := new(MockUserRepository)
	services := &repository.Repositories{
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
	mockUserRepo.On("GetByFederatedIdentifier", mock.Anything, "github", "12345").Return(testUser, nil)
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
	mockUserRepo.On("GetByFederatedIdentifier", mock.Anything, "github", "12345").Return(nil, nil)
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
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
	result, err = RunGithubCreateUserNode(session, node, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "created", result.Condition)
	assert.NotNil(t, session.User)
	assert.Equal(t, "testuser", session.User.Username)
	assert.Equal(t, "test@example.com", session.User.Email)
	assert.Equal(t, "https://github.com/avatar.png", session.User.ProfilePictureURI)

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
