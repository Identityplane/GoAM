package node_user

import (
	"context"

	"github.com/Identityplane/GoAM/pkg/model"
)

var AskUserIDNode = &model.NodeDefinition{
	Name:            "askUserID",
	PrettyName:      "Ask for User ID",
	Description:     "Prompts the user to enter their user ID for authentication",
	Category:        "User Input",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{},
	OutputContext:   []string{"user_id"},
	PossiblePrompts: map[string]string{
		"user_id": "text",
	},
	PossibleResultStates: []string{"submitted", "not_found"},
	Run:                  RunAskUserIDNode,
}

func RunAskUserIDNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
	userID := input["user_id"]
	if userID == "" {
		return model.NewNodeResultWithPrompts(map[string]string{"user_id": "text"})
	}

	// Check if the user ID is valid
	user, err := services.UserRepo.GetByID(context.Background(), userID)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}
	if user == nil {
		state.Context["user_id"] = userID
		return model.NewNodeResultWithCondition("not_found")
	}

	// Set the user in the context
	state.User = user
	state.Context["user_id"] = user.ID

	return model.NewNodeResultWithCondition("submitted")
}
