package internal

import (
	"database/sql"
	"goiam/internal/config"
	"goiam/internal/db/postgres_adapter"
	"goiam/internal/db/sqlite_adapter"
	"goiam/internal/logger"
	"goiam/internal/service"
	"strings"

	"github.com/jackc/pgx/v5"
)

var (
	// All loaded realm configurations, indexed by "tenant/realm"
	LoadedRealms     = map[string]*service.LoadedRealm{}
	UserAdminService service.UserAdminService
)

// Initialize loads all tenant/realm configurations at startup.
// Each realm must include its own flow configuration.
func Initialize() {
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

		// Init user db
		connections.UserDB, err = postgres_adapter.NewPostgresUserDB(postgresdb)
		if err != nil {
			logger.PanicNoContext("Failed to initialize postgres user db: %v", err)
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
	}

	if err != nil {
		logger.PanicNoContext("Failed to initialize database: %v", err)
	}

	return connections
}

// initPostgresDB initializes a PostgreSQL database connection
func initPostgresDB() (*pgx.Conn, error) {
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

	// Initialize realms
	if err := services.RealmService.InitRealms(config.ConfigPath, dbConnections.UserDB); err != nil {
		logger.PanicNoContext("Failed to initialize realms: %v", err)
	}

	// Initialize flows
	if err := services.FlowService.InitFlows(); err != nil {
		logger.PanicNoContext("Failed to initialize flows: %v", err)
	}

	logger.DebugNoContext("Initialized services and realms")
}
