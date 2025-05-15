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

	username := state.Context["username"]
	email := state.Context["email"]
	loginIdentifier := state.Context["loginIdentifier"]

	ctx := context.Background()

	var user *model.User
	var err error

	userLookupMethod := node.CustomConfig["user_lookup_method"]

	switch userLookupMethod {
	case "email":
		user, err = services.UserRepo.GetByEmail(ctx, email)
	case "loginIdentifier":
		user, err = services.UserRepo.GetByLoginIdentifier(ctx, loginIdentifier)
	default:
		user, err = services.UserRepo.GetByUsername(ctx, username)
	}

	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to get user: %w", err))
	}

	if user == nil {
		return model.NewNodeResultWithCondition("not_found")
	}

	state.User = user
	state.Context["user_id"] = user.ID

	return model.NewNodeResultWithCondition("loaded")
}
