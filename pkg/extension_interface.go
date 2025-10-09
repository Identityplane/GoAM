package pkg

import (
	"github.com/Identityplane/GoAM/internal/auth/graph"
	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/pkg/server_settings"
	service_interface "github.com/Identityplane/GoAM/pkg/services"
	"github.com/fasthttp/router"
)

var serverStartCallbacks []func(settings *server_settings.GoamServerSettings) error

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

// OnServerStart registers a callback to be called when the server starts (before the web adapters are initialized)
// If the callback returns an error, the server will not start
// The callbacks are executed in order of registration
func OnServerStart(f func(settings *server_settings.GoamServerSettings) error) {
	serverStartCallbacks = append(serverStartCallbacks, f)
}

// GetServices returns the collection of services to access GoAM internal apis directly
func GetServices() (*service_interface.Services, error) {
	return service.GetServices(), nil
}

// Returns the fasthttp request router to extend or wrap with additional middleware
// This will only be available after the web adapter has been started
// OnServerStart can be used to inject additional routes and middleware this way
func GetRouter() *router.Router {

	return fasthttpRouter
}
