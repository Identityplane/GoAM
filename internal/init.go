package internal

import (
	"database/sql"
	"goiam/internal/config"
	"goiam/internal/db/postgres_adapter"
	"goiam/internal/db/sqlite_adapter"
	"goiam/internal/logger"
	"goiam/internal/service"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// All loaded realm configurations, indexed by "tenant/realm"
	LoadedRealms     = map[string]*service.LoadedRealm{}
	UserAdminService service.UserAdminService
	DBConnections    *service.DatabaseConnections
)

// Initialize loads all tenant/realm configurations at startup.
// Each realm must include its own flow configuration.
func Initialize() {

	config.InitConfiguration()

	// Print config path
	logger.DebugNoContext("Using config path: %s", config.ConfigPath)

	// Step 1: Initialize database connections
	dbConnections := initDatabase()

	// Step 2: Initialize services and realms
	initServices(dbConnections)

}

// initDatabase initializes all database connections based on the connection strings
func initDatabase() *service.DatabaseConnections {
	connections := &service.DatabaseConnections{}
	var err error

	if strings.HasPrefix(config.DBConnString, "postgres://") {

		// Init database connection
		postgresdb, err := initPostgresDB()
		if err != nil {
			logger.PanicNoContext("Failed to initialize postgres database: %v", err)
		}

		service.DbAdapters["postgres"] = postgresdb

		// Run migrations
		err = postgres_adapter.RunMigrations(postgresdb)
		if err != nil {
			logger.PanicNoContext("Failed to run postgres migrations: %v", err)
		}

		// Init user db
		connections.UserDB, err = postgres_adapter.NewPostgresUserDB(postgresdb)
		if err != nil {
			logger.PanicNoContext("Failed to initialize postgres user db: %v", err)
		}

		// Init signing key db
		connections.SigningKeyDB, err = postgres_adapter.NewPostgresSigningKeysDB(postgresdb)
		if err != nil {
			logger.PanicNoContext("Failed to initialize postgres signing key db: %v", err)
		}

		// Init auth session db
		connections.AuthSessionDB, err = postgres_adapter.NewPostgresAuthSessionDB(postgresdb)
		if err != nil {
			logger.PanicNoContext("Failed to initialize postgres auth session db: %v", err)
		}

		// Init client session db
		connections.ClientSessionDB, err = postgres_adapter.NewPostgresClientSessionDB(postgresdb)
		if err != nil {
			logger.PanicNoContext("Failed to initialize postgres client session db: %v", err)
		}

		// Init realm db
		connections.RealmDB, err = postgres_adapter.NewPostgresRealmDB(postgresdb)
		if err != nil {
			logger.PanicNoContext("Failed to initialize postgres realm db: %v", err)
		}

		// Init flow db
		connections.FlowDB, err = postgres_adapter.NewPostgresFlowDB(postgresdb)
		if err != nil {
			logger.PanicNoContext("Failed to initialize postgres flow db: %v", err)
		}

		// Init application db
		connections.ApplicationsDB, err = postgres_adapter.NewPostgresApplicationDB(postgresdb)
		if err != nil {
			logger.PanicNoContext("Failed to initialize postgres application db: %v", err)
		}

	} else {

		// init database connection
		sqliteDB, err := initSQLiteDB()
		if err != nil {
			logger.PanicNoContext("Failed to initialize sqlite database: %v", err)
		}

		// Migrate database, currently we only do this for sqlite
		err = sqlite_adapter.RunMigrations(sqliteDB)
		if err != nil {
			logger.PanicNoContext("Failed to migrate sqlite database: %v", err)
		}

		// init user db
		connections.UserDB, err = sqlite_adapter.NewUserDB(sqliteDB)
		if err != nil {
			logger.PanicNoContext("Failed to initialize sqlite user db: %v", err)
		}

		// init realms db
		connections.RealmDB, err = sqlite_adapter.NewRealmDB(sqliteDB)
		if err != nil {
			logger.PanicNoContext("Failed to initialize sqlite realm db: %v", err)
		}

		// init flows db
		connections.FlowDB, err = sqlite_adapter.NewFlowDB(sqliteDB)
		if err != nil {
			logger.PanicNoContext("Failed to initialize sqlite flows db: %v", err)
		}

		// init applications db
		connections.ApplicationsDB, err = sqlite_adapter.NewApplicationDB(sqliteDB)
		if err != nil {
			logger.PanicNoContext("Failed to initialize sqlite application db: %v", err)
		}

		// init client session db
		connections.ClientSessionDB, err = sqlite_adapter.NewClientSessionDB(sqliteDB)
		if err != nil {
			logger.PanicNoContext("Failed to initialize sqlite client session db: %v", err)
		}

		// init signing key db
		connections.SigningKeyDB, err = sqlite_adapter.NewSigningKeyDB(sqliteDB)
		if err != nil {
			logger.PanicNoContext("Failed to initialize sqlite signing key db: %v", err)
		}

		// init auth session db
		connections.AuthSessionDB, err = sqlite_adapter.NewAuthSessionDB(sqliteDB)
		if err != nil {
			logger.PanicNoContext("Failed to initialize sqlite auth session db: %v", err)
		}
	}

	if err != nil {
		logger.PanicNoContext("Failed to initialize database: %v", err)
	}

	return connections
}

// initPostgresDB initializes a PostgreSQL database connection
func initPostgresDB() (*pgxpool.Pool, error) {
	logger.DebugNoContext("Initializing postgres database")
	db, err := postgres_adapter.Init(postgres_adapter.Config{
		Driver: "postgres",
		DSN:    config.DBConnString,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// initSQLiteDB initializes a SQLite database connection
func initSQLiteDB() (*sql.DB, error) {
	logger.DebugNoContext("Initializing sqlite database")
	db, err := sqlite_adapter.Init(sqlite_adapter.Config{
		Driver: "sqlite",
		DSN:    config.DBConnString,
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
	err := services.StaticConfigurationService.LoadConfigurationFromFiles(config.ConfigPath)
	if err != nil {
		logger.PanicNoContext("Failed to load static configuration: %v", err)
	}

	logger.DebugNoContext("Initialized services and realms")
}
