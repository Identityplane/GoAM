package init

import (
	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/db/postgres_adapter"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConnectionsFactory struct {
	pool *pgxpool.Pool
}

func NewPostgresDBConnectionsFactory() (DBConnectionsFactory, error) {

	// Init database connection
	pool, err := initPostgresDB()
	if err != nil {
		return nil, err
	}

	// Run migrations
	if config.ServerSettings.RunDBMigrations {

		err = postgres_adapter.RunMigrations(pool)
		if err != nil {
			return nil, err
		}
	}

	// Create factory
	return &PostgresConnectionsFactory{pool: pool}, nil
}

func (f *PostgresConnectionsFactory) NewUserDB() (db.UserDB, error) {
	return postgres_adapter.NewPostgresUserDB(f.pool)
}

func (f *PostgresConnectionsFactory) NewUserAttributeDB() (db.UserAttributeDB, error) {
	return postgres_adapter.NewPostgresUserAttributeDB(f.pool)
}

func (f *PostgresConnectionsFactory) NewSigningKeyDB() (db.SigningKeyDB, error) {
	return postgres_adapter.NewPostgresSigningKeysDB(f.pool)
}

func (f *PostgresConnectionsFactory) NewAuthSessionDB() (db.AuthSessionDB, error) {
	return postgres_adapter.NewPostgresAuthSessionDB(f.pool)
}

func (f *PostgresConnectionsFactory) NewClientSessionDB() (db.ClientSessionDB, error) {
	return postgres_adapter.NewPostgresClientSessionDB(f.pool)
}

func (f *PostgresConnectionsFactory) NewApplicationsDB() (db.ApplicationDB, error) {
	return postgres_adapter.NewPostgresApplicationDB(f.pool)
}

func (f *PostgresConnectionsFactory) NewFlowDB() (db.FlowDB, error) {
	return postgres_adapter.NewPostgresFlowDB(f.pool)
}

func (f *PostgresConnectionsFactory) NewRealmDB() (db.RealmDB, error) {
	return postgres_adapter.NewPostgresRealmDB(f.pool)
}

// initPostgresDB initializes a PostgreSQL database connection
func initPostgresDB() (*pgxpool.Pool, error) {
	log := logger.GetGoamLogger()
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
