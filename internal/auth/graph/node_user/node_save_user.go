package node_user

import (
	"context"
	"errors"

	"github.com/Identityplane/GoAM/pkg/model"
)

var SaveUserNode = &model.NodeDefinition{
	Name:            "saveUser",
	PrettyName:      "Save User",
	Description:     "Saves the user object to the database",
	Category:        "User Management",
	Type:            model.NodeTypeLogic,
	RequiredContext: []string{},
	PossibleResultStates: []string{
		"success",
	},
	Run: RunSaveUserNode,
}

func RunSaveUserNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// If there is no user in the context we return an error
	if state.User == nil {
		return model.NewNodeResultWithError(errors.New("user not found in context"))
	}

	// Save the user to the database
	err := services.UserRepo.CreateOrUpdate(context.Background(), state.User)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	return model.NewNodeResultWithCondition("success")
}
