package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/internal/model"

	"github.com/google/uuid"
)

var CreateUserNode = &NodeDefinition{
	Name:            "createUser",
	PrettyName:      "Create User",
	Description:     "Creates a new user account in the database with the provided username and password",
	Category:        "User Management",
	Type:            model.NodeTypeLogic,
	RequiredContext: []string{"username", "password", "email"},
	CustomConfigOptions: map[string]string{
		"checkUsernameUnique": "If set to 'true' the username will be checked for uniqueness. In that case username must be present in the context or user object.",
		"checkEmailUnique":    "If set to 'true' the email will be checked for uniqueness. In that case email must be present in the context or user object.",
	},
	OutputContext:        []string{"user_id"},
	PossibleResultStates: []string{"success", "existing"},
	Run:                  RunCreateUserNode,
}

func RunCreateUserNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {
	ctx := context.Background()

	// Check if we have a user in the context
	if state.User == nil {
		state.User = &model.User{
			ID:        uuid.NewString(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// If we have a username in the context we set it
	if state.Context["username"] != "" {
		state.User.Username = state.Context["username"]
	}

	// If we have an email in the context we set it
	if state.Context["email"] != "" {
		state.User.Email = state.Context["email"]
	}

	// If we have a password in the context we set it
	if state.Context["password"] != "" {
		password := state.Context["password"]
		hashed, err := lib.HashPassword(password)
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to hash password: %w", err))
		}
		state.User.PasswordCredential = string(hashed)
	}

	// TODO Currently we have uniqueness for usernames. This needs to go away but for now we just hack a random id into the username if it is not set by tacking the first 8 characters of the id onto the username
	if state.User.Username == "" {
		state.User.Username = state.User.ID[:8]
	}

	userRepo := services.UserRepo
	if userRepo == nil {
		return model.NewNodeResultWithTextError("UserRepo not initialized")
	}

	// Check for existing user
	existing, _ := userRepo.GetByID(ctx, state.User.ID)
	if existing != nil {
		return model.NewNodeResultWithCondition("existing")
	}

	if node.CustomConfig["checkUsernameUnique"] == "true" {
		existing, _ = userRepo.GetByUsername(ctx, state.User.Username)
		if existing != nil {
			return model.NewNodeResultWithCondition("existing")
		}
	}

	if node.CustomConfig["checkEmailUnique"] == "true" {
		existing, _ = userRepo.GetByEmail(ctx, state.User.Email)
		if existing != nil {
			return model.NewNodeResultWithCondition("existing")
		}
	}

	if err := userRepo.Create(ctx, state.User); err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to create user: %w", err))
	}

	return model.NewNodeResultWithCondition("success")
}
