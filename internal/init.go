package internal

import (
	"database/sql"
	"strings"

	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/db/postgres_adapter"
	"github.com/Identityplane/GoAM/internal/db/sqlite_adapter"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/internal/web/auth"
	"github.com/Identityplane/GoAM/pkg/server_settings"

	services_interface "github.com/Identityplane/GoAM/pkg/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// All loaded realm configurations, indexed by "tenant/realm"
	LoadedRealms     = map[string]*services_interface.LoadedRealm{}
	UserAdminService services_interface.UserAdminService
	DBConnections    *service.DatabaseConnections
)

// Initialize loads all tenant/realm configurations at startup.
// Each realm must include its own flow configuration.
func Initialize(serverSettings *server_settings.GoamServerSettings) {

	config.InitConfiguration(serverSettings)

	// Print config path
	log := logger.GetLogger()
	log.Debug().Str("config_path", config.ServerSettings.RealmConfigurationFolder).Msg("using config path")

	// Step 1: Initialize database connections
	dbConnections := initDatabase()

	// Step 2: Initialize services and realms
	initServices(dbConnections)

}

// initDatabase initializes all database connections based on the connection strings
func initDatabase() *service.DatabaseConnections {
	connections := &service.DatabaseConnections{}
	var err error
	log := logger.GetLogger()

	if strings.HasPrefix(config.ServerSettings.DBConnString, "postgres://") {

		// Init database connection
		postgresdb, err := initPostgresDB()
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize postgres database")
		}

		// Run migrations
		if config.ServerSettings.RunDBMigrations {

			err = postgres_adapter.RunMigrations(postgresdb)
			if err != nil {
				log.Panic().Err(err).Msg("failed to run postgres migrations")
			}
		}

		// Init user db
		connections.UserDB, err = postgres_adapter.NewPostgresUserDB(postgresdb)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize postgres user db")
		}

		// Init user attribute db
		connections.UserAttributeDB, err = postgres_adapter.NewPostgresUserAttributeDB(postgresdb)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize postgres user attribute db")
		}

		// Init signing key db
		connections.SigningKeyDB, err = postgres_adapter.NewPostgresSigningKeysDB(postgresdb)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize postgres signing key db")
		}

		// Init auth session db
		connections.AuthSessionDB, err = postgres_adapter.NewPostgresAuthSessionDB(postgresdb)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize postgres auth session db")
		}

		// Init client session db
		connections.ClientSessionDB, err = postgres_adapter.NewPostgresClientSessionDB(postgresdb)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize postgres client session db")
		}

		// Init realm db
		connections.RealmDB, err = postgres_adapter.NewPostgresRealmDB(postgresdb)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize postgres realm db")
		}

		// Init flow db
		connections.FlowDB, err = postgres_adapter.NewPostgresFlowDB(postgresdb)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize postgres flow db")
		}

		// Init application db
		connections.ApplicationsDB, err = postgres_adapter.NewPostgresApplicationDB(postgresdb)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize postgres application db")
		}

	} else {

		// init database connection
		sqliteDB, err := initSQLiteDB()
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize sqlite database")
		}

		// Migrate database if enabled
		if config.ServerSettings.RunDBMigrations {
			err = sqlite_adapter.RunMigrations(sqliteDB)
			if err != nil {
				log.Panic().Err(err).Msg("failed to migrate sqlite database")
			}
		}

		// init user db
		connections.UserDB, err = sqlite_adapter.NewUserDB(sqliteDB)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize sqlite user db")
		}

		// init user attribute db
		connections.UserAttributeDB, err = sqlite_adapter.NewUserAttributeDB(sqliteDB)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize sqlite user attribute db")
		}

		// init realms db
		connections.RealmDB, err = sqlite_adapter.NewRealmDB(sqliteDB)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize sqlite realm db")
		}

		// init flows db
		connections.FlowDB, err = sqlite_adapter.NewFlowDB(sqliteDB)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize sqlite flows db")
		}

		// init applications db
		connections.ApplicationsDB, err = sqlite_adapter.NewApplicationDB(sqliteDB)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize sqlite application db")
		}

		// init client session db
		connections.ClientSessionDB, err = sqlite_adapter.NewClientSessionDB(sqliteDB)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize sqlite client session db")
		}

		// init signing key db
		connections.SigningKeyDB, err = sqlite_adapter.NewSigningKeyDB(sqliteDB)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize sqlite signing key db")
		}

		// init auth session db
		connections.AuthSessionDB, err = sqlite_adapter.NewAuthSessionDB(sqliteDB)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize sqlite auth session db")
		}
	}

	if err != nil {
		log.Panic().Err(err).Msg("failed to initialize database")
	}

	// init assets
	err = auth.InitAssets()
	if err != nil {
		log.Panic().Err(err).Msg("failed to initialize assets")
	}

	return connections
}

// initPostgresDB initializes a PostgreSQL database connection
func initPostgresDB() (*pgxpool.Pool, error) {
	log := logger.GetLogger()
	log.Debug().Msg("initializing postgres database")
	db, err := postgres_adapter.Init(postgres_adapter.Config{
		Driver: "postgres",
		DSN:    config.ServerSettings.DBConnString,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// initSQLiteDB initializes a SQLite database connection
func initSQLiteDB() (*sql.DB, error) {
	log := logger.GetLogger()
	log.Debug().Msg("initializing sqlite database")
	db, err := sqlite_adapter.Init(sqlite_adapter.Config{
		Driver: "sqlite",
		DSN:    config.ServerSettings.DBConnString,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// initServices initializes all services and loads realm configurations
func initServices(dbConnections *service.DatabaseConnections) {
	// Initialize services
	services := service.InitServices(*dbConnections)

	// Use the static configuration service to load the realm configurations
	err := services.StaticConfigurationService.LoadConfigurationFromFiles(config.ServerSettings.RealmConfigurationFolder)
	if err != nil {
		log := logger.GetLogger()
		log.Panic().Err(err).Msg("failed to load static configuration")
	}

	log := logger.GetLogger()
	log.Debug().Msg("initialized services and realms")
}
