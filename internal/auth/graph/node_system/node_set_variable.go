package node_system

import (
	"fmt"

	"github.com/Identityplane/GoAM/pkg/model"
)

var SetVariableNode = &model.NodeDefinition{
	Name:                 "setVariable",
	PrettyName:           "Set Variable",
	Description:          "Sets a variable in the flow context with a specified key and value",
	Category:             "Flow Control",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{"done"},
	CustomConfigOptions: map[string]string{
		"key":   "The key to set in the context (required)",
		"value": "The value to set for the key in the context (required)",
	},
	Run: RunSetVariableNode,
}

func RunSetVariableNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// check if key is set
	if node.CustomConfig["key"] == "" {
		return nil, fmt.Errorf("key is not set")
	}

	// check if value is set
	if node.CustomConfig["value"] == "" {
		return nil, fmt.Errorf("value is not set")
	}

	key := node.CustomConfig["key"]
	value := node.CustomConfig["value"]

	state.Context[key] = value

	return model.NewNodeResultWithCondition("done")
}
