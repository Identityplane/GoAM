package node_system

import "github.com/Identityplane/GoAM/pkg/model"

var InitSecondaryNode = &model.NodeDefinition{
	Name:                 "initSecondary",
	PrettyName:           "Init Secondary",
	Description:          "Initializes a flow from a secondary device",
	Category:             "Secondary",
	Type:                 model.NodeTypeInit,
	RequiredContext:      []string{},
	OutputContext:        []string{"secondary"},
	PossibleResultStates: []string{"start"},
	Run:                  RunInitSecondaryNode,
}

func RunInitSecondaryNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	return model.NewNodeResultWithCondition("start")
}
