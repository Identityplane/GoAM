package graph

import (
	"context"
	"fmt"
	"goiam/internal/auth/repository"
	"goiam/internal/model"
)

var LoadUserByUsernameNode = &NodeDefinition{
	Name:                 "loadUserByUsername",
	PrettyName:           "Load User from Database",
	Description:          "Loads a user from the database based on the username. The username must be provided in the context.",
	Category:             "User",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username"},
	OutputContext:        []string{"user_id", "username"},
	PossibleResultStates: []string{"loaded", "not_found"},
	Run:                  RunLoadUserNode,
}

func RunLoadUserNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	ctx := context.Background()
	username := state.Context["username"]

	userRepo := services.UserRepo
	if userRepo == nil {
		return model.NewNodeResultWithTextError("UserRepo not initialized")
	}

	// Check for existing user
	existing, err := userRepo.GetByUsername(ctx, username)

	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to get user: %w", err))
	}

	if existing == nil {
		return model.NewNodeResultWithCondition("not_found")
	}

	state.User = existing
	state.Context["user_id"] = existing.ID
	state.Context["username"] = existing.Username

	return model.NewNodeResultWithCondition("loaded")
}
