package service

import (
	"goiam/internal/db"
)

// Services holds all service instances
type Services struct {
	UserService                UserAdminService
	RealmService               RealmService
	FlowService                FlowService
	ApplicationService         ApplicationService
	SessionsService            *SessionsService
	StaticConfigurationService StaticConfigurationService
	OAuth2Service              *OAuth2Service
}

// DatabaseConnections holds all database connections
type DatabaseConnections struct {
	UserDB          db.UserDB
	RealmDB         db.RealmDB
	FlowDB          db.FlowDB
	ApplicationsDB  db.ApplicationDB
	ClientSessionDB db.ClientSessionDB
}

var (
	// Global service registry
	services  *Services
	databases *DatabaseConnections
)

// InitServices initializes all services with their dependencies
func InitServices(connections DatabaseConnections) *Services {

	databases = &connections

	services = &Services{
		UserService:                NewUserService(databases.UserDB),
		RealmService:               NewRealmService(databases.RealmDB, databases.UserDB),
		FlowService:                NewFlowService(databases.FlowDB),
		ApplicationService:         NewApplicationService(databases.ApplicationsDB),
		SessionsService:            NewSessionsService(databases.ClientSessionDB),
		StaticConfigurationService: NewStaticConfigurationService(),
		OAuth2Service:              NewOAuth2Service(),
	}

	return services
}

// GetServices returns the global service registry
func GetServices() *Services {
	if services == nil {
		panic("services not initialized - call InitServices first")
	}
	return services
}
