package graph

import (
	"context"

	"github.com/Identityplane/GoAM/pkg/model"
)

var CheckUsernameAvailableNode = &model.NodeDefinition{
	Name:                 "checkUsernameAvailable",
	PrettyName:           "Check Username Availability",
	Description:          "Checks if a username is available for registration by querying the user database",
	Category:             "User Management",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username"},
	OutputContext:        []string{},
	PossibleResultStates: []string{"available", "taken"},
	Run:                  RunCheckUsernameAvailableNode,
}

func RunCheckUsernameAvailableNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
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
