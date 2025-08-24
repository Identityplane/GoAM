package node_user

import (
	"github.com/Identityplane/GoAM/pkg/model"
)

var InitUserNode = &model.NodeDefinition{
	Name:            "initUser",
	PrettyName:      "Init User",
	Description:     "Initializes an empty user object in the context, or overrides the existing user object with an empty one",
	Category:        "User Management",
	Type:            model.NodeTypeLogic,
	RequiredContext: []string{},
	PossibleResultStates: []string{
		"success",
	},
	Run: RunInitUserNode,
}

func RunInitUserNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	state.User = &model.User{}

	return model.NewNodeResultWithCondition("success")
}
