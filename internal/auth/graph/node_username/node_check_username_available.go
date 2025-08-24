package node_username

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

	existing, err := userRepo.GetByAttributeIndex(ctx, model.AttributeTypeUsername, username)
	if err != nil {
		state.Error = ptr("Username taken")
		return model.NewNodeResultWithCondition("taken")
	}
	if existing != nil {
		state.Error = ptr("Username taken")
		return model.NewNodeResultWithCondition("taken")
	}
	return model.NewNodeResultWithCondition("available")
}

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}
