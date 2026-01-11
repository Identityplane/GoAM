package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/internal/web/auth"
	"github.com/Identityplane/GoAM/pkg/db"
	dbinit "github.com/Identityplane/GoAM/pkg/db/init"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/pkg/server_settings"
	services_init "github.com/Identityplane/GoAM/pkg/services/init"

	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

var (
	// All loaded realm configurations, indexed by "tenant/realm"
	LoadedRealms     = map[string]*services_interface.LoadedRealm{}
	UserAdminService services_interface.UserAdminService
	DBConnections    *db.DatabaseConnections
)

// Initialize loads all tenant/realm configurations at startup.
// Each realm must include its own flow configuration.
func Initialize(serverSettings *server_settings.GoamServerSettings) {

	config.InitConfiguration(serverSettings)

	// Print config path
	log := logger.GetGoamLogger()
	log.Debug().Str("config_path", config.ServerSettings.RealmConfigurationFolder).Msg("using config path")

	// Step 1: Initialize database connections
	dbConnections, err := initDatabase()
	if err != nil {
		log.Panic().Err(err).Msg("failed to initialize database connections")
	}
	DBConnections = dbConnections

	// Step 2: Initialize services and realms
	err = initServices(dbConnections)
	if err != nil {
		log.Panic().Err(err).Msg("failed to initialize services")
	}

	// init assets
	err = auth.InitAssets()
	if err != nil {
		log.Panic().Err(err).Msg("failed to initialize assets")
	}

	// init initial admin user
	err = initInitialAdminUser(serverSettings, service.GetServices())
	if err != nil {
		log.Panic().Err(err).Msg("failed to initialize initial admin user")
	}

}

// initDatabase initializes all database connections based on the connection strings
func initDatabase() (*db.DatabaseConnections, error) {

	// Init the db factory if it is not set
	if dbinit.GetDBConnectionsFactory() == nil {

		if strings.HasPrefix(config.ServerSettings.DBConnString, "postgres://") {

			// Init postgres connections factory
			factory, err := dbinit.NewPostgresDBConnectionsFactory()
			if err != nil {
				return nil, fmt.Errorf("failed to initialize postgres connections factory: %w", err)
			}

			dbinit.SetDBConnectionsFactory(factory)

		} else {

			// Init sqlite connections factory
			factory, err := dbinit.NewSQLiteDBConnectionsFactory()
			if err != nil {
				return nil, fmt.Errorf("failed to initialize sqlite connections factory: %w", err)
			}

			dbinit.SetDBConnectionsFactory(factory)
		}
	}

	return dbinit.GetDatabaseConnections()
}

// initServices initializes all services and loads realm configurations
func initServices(dbConnections *db.DatabaseConnections) error {

	// if the services factory is not set we set the default one
	if services_init.GetServicesFactory() == nil {
		services_init.SetServicesFactory(services_init.NewDefaultServicesFactory(dbConnections))
	}

	// Initialize services
	services, err := services_init.GetServicesFactory().CreateServices()
	if err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// Set the services
	service.SetServices(services)

	// Use the static configuration service to load the realm configurations
	err = services.StaticConfigurationService.LoadConfigurationFromFiles(config.ServerSettings.RealmConfigurationFolder)
	if err != nil {
		return fmt.Errorf("failed to load static configuration: %w", err)
	}

	return nil
}

func initInitialAdminUser(serverSettings *server_settings.GoamServerSettings, services *services_interface.Services) error {

	// if the inital admin user already exists, nothing is done
	adminUserId := serverSettings.InitialAdminId
	adminUserPassword := serverSettings.InitialAdminPassword

	if adminUserId == "" || adminUserPassword == "" {
		return nil
	}

	adminUser, err := services.UserService.GetUserByID(context.Background(), "internal", "internal", adminUserId)
	if err != nil {
		return fmt.Errorf("failed to get initial admin user: %w", err)
	}
	if adminUser != nil {
		return nil
	}

	// Create the inital admin user
	adminUser = &model.User{
		ID:        adminUserId,
		Tenant:    "internal",
		Realm:     "internal",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create a password attribute
	hashed, err := lib.HashPassword(adminUserPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	adminUser.AddAttribute(&model.UserAttribute{
		Type: model.AttributeTypePassword,
		Value: model.PasswordAttributeValue{
			PasswordHash:   string(hashed),
			Locked:         false,
			FailedAttempts: 0,
		},
	})

	// Create entitlment attribute with admin access
	adminUser.AddAttribute(&model.UserAttribute{
		Type: model.AttributeTypeEntitlements,
		Value: model.EntitlementSetAttributeValue{
			Entitlements: []model.Entitlement{
				{
					Description: "Initial Admin User",
					Resource:    "**",
					Action:      "*",
					Effect:      model.EffectTypeAllow,
				},
			},
		},
	})

	// Create the user
	_, err = services.UserService.CreateUserWithAttributes(context.Background(), "internal", "internal", *adminUser)
	if err != nil {
		return fmt.Errorf("failed to create initial admin user: %w", err)
	}

	return nil

}
