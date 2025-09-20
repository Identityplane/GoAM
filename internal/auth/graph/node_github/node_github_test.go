package node_github

import (
	"testing"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGithubLogin_NewUserCreate(t *testing.T) {
	// Create mock GitHub API server
	githubServer := setupGithubMockServer(t, 12345)
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
			CONFIG_GITHUB_CLIENT_ID:          "test_client_id",
			CONFIG_GITHUB_CLIENT_SECRET:      "test_client_secret",
			CONFIG_GITHUB_SCOPE:              "user",
			CONFIG_CREATE_USER_IF_NOT_EXISTS: "false",
		},
	}

	// Test 1: Initial state - should return redirect URL
	session := &model.AuthenticationSession{
		LoginUri: "http://localhost:8080/callback",
		Context:  make(map[string]string),
	}

	t.Run("Initial state - should return redirect URL", func(t *testing.T) {

		result, err := RunGithubLoginNode(session, node, map[string]string{}, services)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Prompts)
		assert.Contains(t, result.Prompts["__redirect"], "github.com/login/oauth/authorize")
	})

	t.Run("With code - existing user", func(t *testing.T) {

		// Test 2: With code - existing user
		//Arrange: Return nil when querying for the user attribute
		mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeGitHub, "12345").Return(testUser, nil)

		// Act:
		result, err := RunGithubLoginNode(session, node, map[string]string{"code": "test_code"}, services)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Assert:
		assert.Equal(t, "existing-user", result.Condition)
		assert.Equal(t, testUser, session.User)
	})

	t.Run("With code - new user", func(t *testing.T) {

		// Arrange: Return user attribute when querying for the user attribute
		mockUserRepo.ExpectedCalls = nil // Reset mock expectations
		mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeGitHub, "12345").Return(nil, nil)

		// Act
		result, err := RunGithubLoginNode(session, node, map[string]string{"code": "test_code"}, services)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "new-user", result.Condition)

		// Assert:
		// Verify that the GitHub attribute was created correctly
		assert.Len(t, session.User.UserAttributes, 1)
		githubAttr := session.User.UserAttributes[0]
		assert.Equal(t, model.AttributeTypeGitHub, githubAttr.Type)
		assert.Equal(t, stringPtr("12345"), githubAttr.Index) // Should use GitHub User ID as index

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

	})

	t.Run("With code - new user with create user if not exists", func(t *testing.T) {

		// Arrange
		node.CustomConfig[CONFIG_CREATE_USER_IF_NOT_EXISTS] = "true"
		mockUserRepo.ExpectedCalls = nil // Reset mock expectations
		mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeGitHub, "12345").Return(nil, nil)
		mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *model.User) bool {
			return user.ID == session.User.ID
		})).Return(nil)

		// Act
		result, err := RunGithubLoginNode(session, node, map[string]string{"code": "test_code"}, services)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "new-user", result.Condition)

		// Assert:
		mockUserRepo.AssertCalled(t, "Create", mock.Anything, mock.MatchedBy(func(user *model.User) bool {
			return user.ID == session.User.ID
		}))
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
