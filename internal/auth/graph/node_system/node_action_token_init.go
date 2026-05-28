package node_system

import "github.com/Identityplane/GoAM/pkg/model"

var ActionTokenInitNode = &model.NodeDefinition{
	Name:                 "actionTokenInit",
	PrettyName:           "Action Token Init",
	Description:          "Initializes a flow from an action token",
	Category:             "Action Token",
	Type:                 model.NodeTypeInit,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{"start"},
	Run:                  RunActionTokenInitNode,
}

func RunActionTokenInitNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	return model.NewNodeResultWithCondition("start")
}
