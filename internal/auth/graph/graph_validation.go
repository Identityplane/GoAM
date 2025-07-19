package graph

import (
	"errors"
	"fmt"

	"github.com/Identityplane/GoAM/internal/model"
)

// ValidateFlowDefinition checks for basic structural integrity of the flow
func ValidateFlowDefinition(def *model.FlowDefinition) error {

	// Check if start node is in map
	_, ok := def.Nodes[def.Start]
	if !ok {
		return fmt.Errorf("start node '%s' not found in nodes", def.Start)
	}

	// Check start node is of type 'init'
	if def.Start == "" {
		return errors.New("flow start node is not defined")
	}
	if def.Nodes[def.Start] == nil {
		return fmt.Errorf("start node '%s' is missing from the graph", def.Start)
	}
	if nodeDef := def.Nodes[def.Start]; nodeDef.Use != "init" {
		return fmt.Errorf("start node '%s' must be of type 'init'", def.Start)
	}

	// Check non-terminal nodes have a Next map
	for name, node := range def.Nodes {
		nodeType := model.NodeTypeInit
		if def := getNodeDefinitionByName(node.Use); def != nil {
			nodeType = def.Type
		} else if node.Use == "successResult" || node.Use == "failureResult" {
			nodeType = model.NodeTypeResult
		}

		if nodeType != model.NodeTypeResult && node.Next == nil {
			return fmt.Errorf("node '%s' must define a 'Next' map", name)
		}
	}
	return nil
}
