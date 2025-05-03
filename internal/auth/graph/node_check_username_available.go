package graph

import (
	"context"
	"goiam/internal/auth/repository"
	"goiam/internal/model"
)

var CheckUsernameAvailableNode = &NodeDefinition{
	Name:                 "checkUsernameAvailable",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username"},
	OutputContext:        []string{},
	PossibleResultStates: []string{"available", "taken"},
	Run:                  RunCheckUsernameAvailableNode,
}

func RunCheckUsernameAvailableNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {
	username := state.Context["username"]
	ctx := context.Background()

	// Load use Repository
	userRepo := services.UserRepo
	if userRepo == nil {
		return model.NewNodeResultWithTextError("UserRepo not initialized")
	}

	existing, err := userRepo.GetByUsername(ctx, username)
	if err != nil {
		return model.NewNodeResultWithCondition("taken")
	}
	if existing != nil {
		return model.NewNodeResultWithCondition("taken")
	}
	return model.NewNodeResultWithCondition("available")
}
