package init

import (
	"fmt"

	"github.com/Identityplane/GoAM/pkg/db"
)

type DBConnectionsFactory interface {
	NewUserDB() (db.UserDB, error)
	NewUserAttributeDB() (db.UserAttributeDB, error)
	NewRealmDB() (db.RealmDB, error)
	NewFlowDB() (db.FlowDB, error)
	NewApplicationsDB() (db.ApplicationDB, error)
	NewClientSessionDB() (db.ClientSessionDB, error)
	NewSigningKeyDB() (db.SigningKeyDB, error)
	NewAuthSessionDB() (db.AuthSessionDB, error)
}

// Singleton instance of the DBConnectionsFactory
var dbConnectionsFactory DBConnectionsFactory

func SetDBConnectionsFactory(factory DBConnectionsFactory) {
	dbConnectionsFactory = factory
}

func GetDBConnectionsFactory() DBConnectionsFactory {
	return dbConnectionsFactory
}

func GetDatabaseConnections() (*db.DatabaseConnections, error) {

	// Get the database connections factory
	factory := GetDBConnectionsFactory()
	if factory == nil {
		return nil, fmt.Errorf("database connections factory is not set")
	}

	var err error

	// Create the database connections
	connections := &db.DatabaseConnections{}

	// Init user db
	connections.UserDB, err = factory.NewUserDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres user db: %w", err)
	}

	// Init user attribute db
	connections.UserAttributeDB, err = factory.NewUserAttributeDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres user attribute db: %w", err)
	}

	// Init signing key db
	connections.SigningKeyDB, err = factory.NewSigningKeyDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres signing key db: %w", err)
	}

	// Init auth session db
	connections.AuthSessionDB, err = factory.NewAuthSessionDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres auth session db: %w", err)
	}

	// Init client session db
	connections.ClientSessionDB, err = factory.NewClientSessionDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres client session db: %w", err)
	}

	// Init realm db
	connections.RealmDB, err = factory.NewRealmDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres realm db: %w", err)
	}

	// Init flow db
	connections.FlowDB, err = factory.NewFlowDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres flow db: %w", err)
	}

	// Init application db
	connections.ApplicationsDB, err = factory.NewApplicationsDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres application db: %w", err)
	}

	return connections, nil
}
