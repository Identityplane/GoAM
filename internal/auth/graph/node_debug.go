package graph

import (
	"github.com/Identityplane/GoAM/pkg/model"
)

var DebugNode = &model.NodeDefinition{
	Name:            "debug",
	PrettyName:      "Debug Breakpoint",
	Description:     "This node is used to break the flow and debug the flow. Shows the current flow information if debug mode is enabled.",
	Category:        "Information",
	Type:            model.NodeTypeQueryWithLogic,
	Run:             RunDebugNode,
	RequiredContext: []string{},
	OutputContext:   []string{},
	PossiblePrompts: map[string]string{
		"continue": "boolean",
	},
	PossibleResultStates: []string{"continue"},
	CustomConfigOptions:  map[string]string{},
}

func RunDebugNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// If continue is true, we continue the flow
	if input["continue"] == "true" {
		return model.NewNodeResultWithCondition("continue")
	}

	// If debug mode is enabled, we show the debug information throught the template renderer
	if state.Debug {
		return model.NewNodeResultWithPrompts(map[string]string{"continue": "boolean"})
	}

	// If debug mode is disabled, we continue the flow
	return model.NewNodeResultWithCondition("continue")
}
