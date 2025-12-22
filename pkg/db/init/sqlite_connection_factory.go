package init

import (
	"database/sql"

	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/db/sqlite_adapter"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/db"
)

type SQLiteConnectionsFactory struct {
	db *sql.DB
}

func NewSQLiteDBConnectionsFactory() (DBConnectionsFactory, error) {

	// Init database connection
	db, err := initSQLiteDB()
	if err != nil {
		return nil, err
	}

	// Run migrations
	if config.ServerSettings.RunDBMigrations {

		err = sqlite_adapter.RunMigrations(db)
		if err != nil {
			return nil, err
		}
	}

	// Create factory
	return &SQLiteConnectionsFactory{db: db}, nil
}

func (f *SQLiteConnectionsFactory) NewUserDB() (db.UserDB, error) {
	return sqlite_adapter.NewUserDB(f.db)
}

func (f *SQLiteConnectionsFactory) NewUserAttributeDB() (db.UserAttributeDB, error) {
	return sqlite_adapter.NewUserAttributeDB(f.db)
}

func (f *SQLiteConnectionsFactory) NewSigningKeyDB() (db.SigningKeyDB, error) {
	return sqlite_adapter.NewSigningKeyDB(f.db)
}

func (f *SQLiteConnectionsFactory) NewAuthSessionDB() (db.AuthSessionDB, error) {
	return sqlite_adapter.NewAuthSessionDB(f.db)
}

func (f *SQLiteConnectionsFactory) NewClientSessionDB() (db.ClientSessionDB, error) {
	return sqlite_adapter.NewClientSessionDB(f.db)
}

func (f *SQLiteConnectionsFactory) NewApplicationsDB() (db.ApplicationDB, error) {
	return sqlite_adapter.NewApplicationDB(f.db)
}

func (f *SQLiteConnectionsFactory) NewFlowDB() (db.FlowDB, error) {
	return sqlite_adapter.NewFlowDB(f.db)
}

func (f *SQLiteConnectionsFactory) NewRealmDB() (db.RealmDB, error) {
	return sqlite_adapter.NewRealmDB(f.db)
}

// initSQLiteDB initializes a SQLite database connection
func initSQLiteDB() (*sql.DB, error) {
	log := logger.GetGoamLogger()
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
