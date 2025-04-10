package graph

import (
	"context"
	"errors"
)

var CheckUsernameAvailableNode = &NodeDefinition{
	Name:       "checkUsernameAvailable",
	Type:       NodeTypeLogic,
	Inputs:     []string{"username"},
	Outputs:    []string{},
	Conditions: []string{"available", "taken"},
}

func RunCheckUsernameAvailableNode(state *FlowState, node *GraphNode) (string, error) {
	username := state.Context["username"]
	ctx := context.Background()

	userRepo := Services.UserRepo
	if userRepo == nil {
		return "taken", errors.New("UserRepo not initialized")
	}

	existing, err := userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "taken", err
	}
	if existing != nil {
		return "taken", nil
	}
	return "available", nil
}
