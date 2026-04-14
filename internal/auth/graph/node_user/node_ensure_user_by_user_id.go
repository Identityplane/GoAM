package node_user

import (
	"context"
	"errors"

	"github.com/Identityplane/GoAM/pkg/model"
)

var EnsureUserByUserIdNode = &model.NodeDefinition{
	Name:                 "ensureUserByUserId",
	PrettyName:           "Ensure user by user_id",
	Description:          "Loads a user by context user_id, or creates a new user with that id if missing.",
	Category:             "User Management",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"user_id"},
	PossibleResultStates: []string{"success"},
	Run:                  RunEnsureUserByUserIdNode,
}

func RunEnsureUserByUserIdNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
	id := state.Context["user_id"]
	if id == "" {
		return model.NewNodeResultWithError(errors.New("user_id missing from context"))
	}

	ctx := context.Background()
	user, err := services.UserRepo.GetByID(ctx, id)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}
	if user != nil {
		state.User = user
		return model.NewNodeResultWithCondition("success")
	}

	u, err := services.UserRepo.NewUserModel(state)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}
	state.User = u
	if err := services.UserRepo.Create(ctx, u); err != nil {
		return model.NewNodeResultWithError(err)
	}

	return model.NewNodeResultWithCondition("success")
}
