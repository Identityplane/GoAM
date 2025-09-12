package service

import (
	"github.com/Identityplane/GoAM/internal/db"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

// DatabaseConnections holds all database connections
type DatabaseConnections struct {
	UserDB          db.UserDB
	UserAttributeDB db.UserAttributeDB
	RealmDB         db.RealmDB
	FlowDB          db.FlowDB
	ApplicationsDB  db.ApplicationDB
	ClientSessionDB db.ClientSessionDB
	SigningKeyDB    db.SigningKeyDB
	AuthSessionDB   db.AuthSessionDB
}

var (
	// Global service registry
	services  *services_interface.Services
	Databases *DatabaseConnections
)

// InitServices initializes all services with their dependencies
func InitServices(connections DatabaseConnections) *services_interface.Services {
	Databases = &connections

	// Initialize cache service first
	cacheService, err := NewCacheService()
	if err != nil {
		panic("failed to initialize cache service: " + err.Error())
	}

	services = &services_interface.Services{
		UserService:                NewUserService(Databases.UserDB, Databases.UserAttributeDB),
		UserAttributeService:       NewUserAttributeService(Databases.UserAttributeDB, Databases.UserDB),
		RealmService:               NewCachedRealmService(NewRealmService(Databases.RealmDB, Databases.UserDB, Databases.UserAttributeDB), cacheService),
		FlowService:                NewCachedFlowService(NewFlowService(Databases.FlowDB), cacheService),
		ApplicationService:         NewApplicationService(Databases.ApplicationsDB),
		SessionsService:            NewCachedSessionsService(NewSessionsService(Databases.ClientSessionDB, Databases.AuthSessionDB), cacheService),
		StaticConfigurationService: NewStaticConfigurationService(),
		OAuth2Service:              NewOAuth2Service(),
		JWTService:                 NewCachedJWTService(NewJWTService(Databases.SigningKeyDB), cacheService),
		CacheService:               cacheService,
		TemplatesService:           NewTemplatesService(),
		AdminAuthzService:          NewAdminAuthzService(),
	}

	return services
}

// GetServices returns the global service registry
func GetServices() *services_interface.Services {
	if services == nil {
		panic("services not initialized - call InitServices first")
	}
	return services
}
