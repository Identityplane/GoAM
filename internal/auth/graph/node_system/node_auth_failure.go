package node_system

import "github.com/Identityplane/GoAM/pkg/model"

var FailureResultNode = &model.NodeDefinition{
	Name:                 model.NODE_FAILURE_RESULT,
	PrettyName:           "Authentication Failure",
	Description:          "Terminal node that indicates failed authentication",
	Category:             "Results",
	Type:                 model.NodeTypeResult,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{}, // terminal node
	Run:                  RunAuthFailureNode,
}

func RunAuthFailureNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
	state.Result = &model.FlowResult{
		UserID:        "",
		Authenticated: false}

	return &model.NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil
}
