package node_user

import (
	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

var LoadUserNode = &model.NodeDefinition{
	Name:                 "loadUser",
	PrettyName:           "Load User from Database",
	Description:          "Loads a user from the database.",
	Category:             "User",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username", "email", "loginIdentifier"},
	OutputContext:        []string{"user_id", "username"},
	PossibleResultStates: []string{"loaded", "not_found"},
	Run:                  RunLoadUserNode,
}

func RunLoadUserNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	if user == nil {
		return model.NewNodeResultWithCondition("not_found")
	}

	state.User = user
	return model.NewNodeResultWithCondition("loaded")
}
