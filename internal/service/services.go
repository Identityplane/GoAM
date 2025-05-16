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
	JWTService                 JWTService
	CacheService               CacheService
}

// DatabaseConnections holds all database connections
type DatabaseConnections struct {
	UserDB          db.UserDB
	RealmDB         db.RealmDB
	FlowDB          db.FlowDB
	ApplicationsDB  db.ApplicationDB
	ClientSessionDB db.ClientSessionDB
	SigningKeyDB    db.SigningKeyDB
	AuthSessionDB   db.AuthSessionDB
}

var (
	// Global service registry
	services  *Services
	databases *DatabaseConnections
)

// InitServices initializes all services with their dependencies
func InitServices(connections DatabaseConnections) *Services {
	databases = &connections

	// Initialize cache service first
	cacheService, err := NewCacheService()
	if err != nil {
		panic("failed to initialize cache service: " + err.Error())
	}

	services = &Services{
		UserService:                NewUserService(databases.UserDB),
		RealmService:               NewCachedRealmService(NewRealmService(databases.RealmDB, databases.UserDB), cacheService),
		FlowService:                NewCachedFlowService(NewFlowService(databases.FlowDB), cacheService),
		ApplicationService:         NewApplicationService(databases.ApplicationsDB),
		SessionsService:            NewSessionsService(databases.ClientSessionDB, databases.AuthSessionDB),
		StaticConfigurationService: NewStaticConfigurationService(),
		OAuth2Service:              NewOAuth2Service(),
		JWTService:                 NewCachedJWTService(NewJWTService(databases.SigningKeyDB), cacheService),
		CacheService:               cacheService,
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
