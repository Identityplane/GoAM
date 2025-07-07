package service

import (
	"github.com/gianlucafrei/GoAM/internal/db"
)

// Services holds all service instances
type Services struct {
	UserService                UserAdminService
	RealmService               RealmService
	FlowService                FlowService
	ApplicationService         ApplicationService
	SessionsService            SessionsService
	StaticConfigurationService StaticConfigurationService
	OAuth2Service              *OAuth2Service
	JWTService                 JWTService
	CacheService               CacheService
	AdminAuthzService          AdminAuthzService
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
	Databases *DatabaseConnections
)

// InitServices initializes all services with their dependencies
func InitServices(connections DatabaseConnections) *Services {
	Databases = &connections

	// Initialize cache service first
	cacheService, err := NewCacheService()
	if err != nil {
		panic("failed to initialize cache service: " + err.Error())
	}

	services = &Services{
		UserService:                NewUserService(Databases.UserDB),
		RealmService:               NewCachedRealmService(NewRealmService(Databases.RealmDB, Databases.UserDB), cacheService),
		FlowService:                NewCachedFlowService(NewFlowService(Databases.FlowDB), cacheService),
		ApplicationService:         NewApplicationService(Databases.ApplicationsDB),
		SessionsService:            NewCachedSessionsService(NewSessionsService(Databases.ClientSessionDB, Databases.AuthSessionDB), cacheService),
		StaticConfigurationService: NewStaticConfigurationService(),
		OAuth2Service:              NewOAuth2Service(),
		JWTService:                 NewCachedJWTService(NewJWTService(Databases.SigningKeyDB), cacheService),
		CacheService:               cacheService,
		AdminAuthzService:          NewAdminAuthzService(),
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
