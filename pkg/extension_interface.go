package pkg

import (
	"github.com/Identityplane/GoAM/internal/auth/graph"
	"github.com/Identityplane/GoAM/pkg/model"
)

// RegisterNode registers a node definition. Use this to add a new custom node
func RegisterNode(name string, def *model.NodeDefinition) {
	graph.NodeDefinitions[name] = def
}

// UnregisterNode unregisters a node definition. Use this to remove a node
func UnregisterNode(name string) {
	delete(graph.NodeDefinitions, name)
}
