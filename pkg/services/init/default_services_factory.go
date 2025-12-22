package init

import (
	"fmt"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/internal/service/email"
	"github.com/Identityplane/GoAM/pkg/db"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

type DefaultServicesFactory struct {
	dbConnections *db.DatabaseConnections
}

func NewDefaultServicesFactory(dbConnections *db.DatabaseConnections) ServicesFactory {
	return &DefaultServicesFactory{dbConnections: dbConnections}
}

func (f *DefaultServicesFactory) CreateServices() (*services_interface.Services, error) {

	// Initialize cache service first
	cacheService, err := service.NewCacheService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache service: %w", err)
	}

	services := &services_interface.Services{
		UserService:                service.NewUserService(f.dbConnections.UserDB, f.dbConnections.UserAttributeDB),
		UserAttributeService:       service.NewUserAttributeService(f.dbConnections.UserAttributeDB, f.dbConnections.UserDB),
		RealmService:               service.NewCachedRealmService(service.NewRealmService(f.dbConnections.RealmDB, f.dbConnections.UserDB, f.dbConnections.UserAttributeDB), cacheService),
		FlowService:                service.NewCachedFlowService(service.NewFlowService(f.dbConnections.FlowDB), cacheService),
		ApplicationService:         service.NewApplicationService(f.dbConnections.ApplicationsDB),
		SessionsService:            service.NewCachedSessionsService(service.NewSessionsService(f.dbConnections.ClientSessionDB, f.dbConnections.AuthSessionDB), cacheService),
		StaticConfigurationService: service.NewStaticConfigurationService(),
		OAuth2Service:              service.NewOAuth2Service(),
		JWTService:                 service.NewCachedJWTService(service.NewJWTService(f.dbConnections.SigningKeyDB), cacheService),
		CacheService:               cacheService,
		TemplatesService:           service.NewTemplatesService(),
		AdminAuthzService:          service.NewAdminAuthzService(),
		SimpleAuthService:          service.NewSimpleAuthService(),
		EmailService:               email.NewDefaultEmailService(),
		UserClaimsService:          service.NewUserClaimsService(),
	}

	return services, nil

}
