package node_system

import "github.com/Identityplane/GoAM/pkg/model"

var ErrorNode = &model.NodeDefinition{
	Name:                 model.NODE_ERROR,
	PrettyName:           "Error",
	Description:          "Terminal node that indicates an error during flow execution",
	Category:             "System",
	Type:                 model.NodeTypeResult,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{}, // terminal node
	CustomConfigOptions: map[string]string{
		"title":   "Title of the error",
		"message": "Message of the error",
	},
	Run: RunErrorNode,
}

func RunErrorNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	return &model.NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil
}
