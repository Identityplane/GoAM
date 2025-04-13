package graph

import (
	"context"
)

var CheckUsernameAvailableNode = &NodeDefinition{
	Name:            "checkUsernameAvailable",
	Type:            NodeTypeLogic,
	RequiredContext: []string{"username"},
	OutputContext:   []string{},
	Conditions:      []string{"available", "taken"},
	Run:             RunCheckUsernameAvailableNode,
}

func RunCheckUsernameAvailableNode(state *FlowState, node *GraphNode, input map[string]string) (*NodeResult, error) {
	username := state.Context["username"]
	ctx := context.Background()

	// Load use Repository
	userRepo := Services.UserRepo
	if userRepo == nil {
		return NewNodeResultWithTextError("UserRepo not initialized")
	}

	existing, err := userRepo.GetByUsername(ctx, username)
	if err != nil {
		return NewNodeResultWithCondition("taken")
	}
	if existing != nil {
		return NewNodeResultWithCondition("taken")
	}
	return NewNodeResultWithCondition("available")
}
