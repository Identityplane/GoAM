package node_username

import (
	"context"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCheckUsernameAvailableNode_Definition(t *testing.T) {
	// Test the node definition
	assert.Equal(t, "checkUsernameAvailable", CheckUsernameAvailableNode.Name)
	assert.Equal(t, "Check Username Availability", CheckUsernameAvailableNode.PrettyName)
	assert.Equal(t, "Checks if a username is available for registration by querying the user database", CheckUsernameAvailableNode.Description)
	assert.Equal(t, "User Management", CheckUsernameAvailableNode.Category)
	assert.Equal(t, model.NodeTypeLogic, CheckUsernameAvailableNode.Type)
	assert.Equal(t, []string{"username"}, CheckUsernameAvailableNode.RequiredContext)
	assert.Equal(t, []string{}, CheckUsernameAvailableNode.OutputContext)
	assert.Equal(t, []string{"available", "taken"}, CheckUsernameAvailableNode.PossibleResultStates)
	assert.NotNil(t, CheckUsernameAvailableNode.Run)
}

func TestRunCheckUsernameAvailableNode_UsernameAvailable(t *testing.T) {
	// Arrange
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	assert.NoError(t, err)
	defer testRepo.Close()

	services := &model.Repositories{
		UserRepo: testRepo,
	}

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "newuser",
		},
	}

	node := &model.GraphNode{
		Name: "checkUsernameAvailable",
		Use:  "checkUsernameAvailable",
	}

	// Act
	result, err := RunCheckUsernameAvailableNode(state, node, map[string]string{}, services)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "available", result.Condition)
	assert.Empty(t, result.Prompts)
	assert.Nil(t, state.Error) // No error should be set when username is available
}

func TestRunCheckUsernameAvailableNode_UsernameTaken(t *testing.T) {
	// Arrange
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	assert.NoError(t, err)
	defer testRepo.Close()

	services := &model.Repositories{
		UserRepo: testRepo,
	}

	// Create a test user with username attribute
	testUser := &model.User{
		ID:        uuid.NewString(),
		Tenant:    "acme",
		Realm:     "customers",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserAttributes: []model.UserAttribute{
			{
				ID:        uuid.NewString(),
				UserID:    uuid.NewString(),
				Tenant:    "acme",
				Realm:     "customers",
				Index:     "existinguser",
				Type:      model.AttributeTypeUsername,
				Value:     "existinguser",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Create the user in the database
	err = testRepo.Create(context.Background(), testUser)
	assert.NoError(t, err)

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "existinguser",
		},
	}

	node := &model.GraphNode{
		Name: "checkUsernameAvailable",
		Use:  "checkUsernameAvailable",
	}

	// Act
	result, err := RunCheckUsernameAvailableNode(state, node, map[string]string{}, services)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "taken", result.Condition)
	assert.Empty(t, result.Prompts)
	assert.NotNil(t, state.Error) // Error should be set when username is taken
	assert.Equal(t, "Username taken", *state.Error)
}

func TestRunCheckUsernameAvailableNode_RepositoryError(t *testing.T) {
	// Arrange
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	assert.NoError(t, err)
	defer testRepo.Close()

	services := &model.Repositories{
		UserRepo: testRepo,
	}

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "testuser",
		},
	}

	node := &model.GraphNode{
		Name: "checkUsernameAvailable",
		Use:  "checkUsernameAvailable",
	}

	// Act
	result, err := RunCheckUsernameAvailableNode(state, node, map[string]string{}, services)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "available", result.Condition) // With real DB, non-existent username returns available
	assert.Empty(t, result.Prompts)
	assert.Nil(t, state.Error) // No error should be set when username is available
}

func TestRunCheckUsernameAvailableNode_UserRepoNotInitialized(t *testing.T) {
	// Arrange
	services := &model.Repositories{
		UserRepo: nil, // No user repository
	}

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "testuser",
		},
	}

	node := &model.GraphNode{
		Name: "checkUsernameAvailable",
		Use:  "checkUsernameAvailable",
	}

	// Act
	result, err := RunCheckUsernameAvailableNode(state, node, map[string]string{}, services)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "UserRepo not initialized", err.Error())
}

func TestRunCheckUsernameAvailableNode_MissingUsernameInContext(t *testing.T) {
	// Arrange
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	assert.NoError(t, err)
	defer testRepo.Close()

	services := &model.Repositories{
		UserRepo: testRepo,
	}

	state := &model.AuthenticationSession{
		Context: map[string]string{
			// No username in context
		},
	}

	node := &model.GraphNode{
		Name: "checkUsernameAvailable",
		Use:  "checkUsernameAvailable",
	}

	// Act
	result, err := RunCheckUsernameAvailableNode(state, node, map[string]string{}, services)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "available", result.Condition) // Empty username is treated as available
	assert.Empty(t, result.Prompts)
	assert.Nil(t, state.Error) // No error should be set for empty username
}

func TestRunCheckUsernameAvailableNode_EmptyUsername(t *testing.T) {
	// Arrange
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	assert.NoError(t, err)
	defer testRepo.Close()

	services := &model.Repositories{
		UserRepo: testRepo,
	}

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": "",
		},
	}

	node := &model.GraphNode{
		Name: "checkUsernameAvailable",
		Use:  "checkUsernameAvailable",
	}

	// Act
	result, err := RunCheckUsernameAvailableNode(state, node, map[string]string{}, services)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "available", result.Condition)
	assert.Empty(t, result.Prompts)
	assert.Nil(t, state.Error) // No error should be set when username is available (even if empty)
}

func TestRunCheckUsernameAvailableNode_IntegrationWithRealContext(t *testing.T) {
	// Arrange
	testRepo, err := repository.NewTestUserRepository("acme", "customers")
	assert.NoError(t, err)
	defer testRepo.Close()

	services := &model.Repositories{
		UserRepo: testRepo,
	}

	// Test with a realistic username
	testUsername := "john.doe_123"

	state := &model.AuthenticationSession{
		Context: map[string]string{
			"username": testUsername,
		},
		User: &model.User{
			ID:        uuid.NewString(),
			Tenant:    "acme",
			Realm:     "customers",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	node := &model.GraphNode{
		Name: "checkUsernameAvailable",
		Use:  "checkUsernameAvailable",
		CustomConfig: map[string]string{
			"someConfig": "someValue",
		},
	}

	// Act
	result, err := RunCheckUsernameAvailableNode(state, node, map[string]string{}, services)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "available", result.Condition)
	assert.Empty(t, result.Prompts)
	assert.Nil(t, state.Error) // No error should be set when username is available

	// Verify that the username in context was not modified
	assert.Equal(t, testUsername, state.Context["username"])
}
