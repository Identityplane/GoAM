package pkg

import (
	"github.com/Identityplane/GoAM/internal/auth/graph"
	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/pkg/server_settings"
)

// RegisterNode registers a node definition. Use this to add a new custom node
func RegisterNode(name string, def *model.NodeDefinition) {
	graph.NodeDefinitions[name] = def
}

// UnregisterNode unregisters a node definition. Use this to remove a node
func UnregisterNode(name string) {
	delete(graph.NodeDefinitions, name)
}

// GetServerConfig returns a pointer to the current server runtime configuration
func GetServerConfig() *server_settings.GoamServerSettings {
	return config.ServerSettings
}
