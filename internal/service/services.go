package service

import (
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

var (
	// Global service registry
	services *services_interface.Services
)

// InitServices initializes all services with their dependencies
func SetServices(theServices *services_interface.Services) {

	services = theServices

}

// GetServices returns the global service registry
func GetServices() *services_interface.Services {
	if services == nil {
		panic("services not initialized - call InitServices first")
	}
	return services
}
