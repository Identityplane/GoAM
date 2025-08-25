package node_system

import "github.com/Identityplane/GoAM/pkg/model"

var InitNode = &model.NodeDefinition{
	Name:                 "init",
	PrettyName:           "Initialize Flow",
	Description:          "Initializes the authentication flow and sets up the starting state",
	Category:             "Flow Control",
	Type:                 model.NodeTypeInit,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{"start"},
	Run:                  RunInitNode,
}

func RunInitNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	return model.NewNodeResultWithCondition("start")
}
