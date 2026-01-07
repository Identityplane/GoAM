package sqlite_adapter

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/Identityplane/GoAM/internal/logger"

	_ "modernc.org/sqlite" // sqlite driver
)

type Config struct {
	Driver string // "sqlite" or "pgx"
	DSN    string
}

var log = logger.GetGoamLogger()

func Init(cfg Config) (*sql.DB, error) {
	var err error
	// Add connection parameters for better concurrency
	// _busy_timeout sets how long SQLite will wait for a lock (in milliseconds)
	// _journal_mode=WAL enables Write-Ahead Logging for better concurrency
	dsn := cfg.DSN
	if !strings.Contains(dsn, "_busy_timeout") {
		if strings.Contains(dsn, "?") {
			dsn += "&_busy_timeout=5000&_journal_mode=WAL"
		} else {
			dsn += "?_busy_timeout=5000&_journal_mode=WAL"
		}
	}

	database, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	// Set connection pool settings
	database.SetMaxOpenConns(1) // SQLite works best with a single connection
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(0) // Keep connections alive

	// Set busy timeout using PRAGMA (works with modernc.org/sqlite driver)
	_, err = database.Exec("PRAGMA busy_timeout = 5000")
	if err != nil {
		log.Warn().Err(err).Msg("failed to set busy_timeout pragma")
	}

	// Enable WAL mode for better concurrency
	_, err = database.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		log.Warn().Err(err).Msg("failed to set WAL mode")
	}

	if err = database.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return database, nil
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Uses embed fs to load migrations and run them over the database connection in order
func RunMigrations(db *sql.DB) error {

	// Open migrations folder
	migrations, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// go over all files in migrations folder
	for _, migration := range migrations {
		// read file
		migrationFile, err := migrationsFS.ReadFile("migrations/" + migration.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration: %w", err)
		}

		// Only expecute up migrations
		if strings.Contains(migration.Name(), "up.sql") {
			// Log name of migration
			log.Debug().
				Str("migration", migration.Name()).
				Msg("running migration")

			// run migration
			_, err = db.Exec(string(migrationFile))
			if err != nil {
				// Check if this is a "duplicate column" error (SQLite error code 1)
				// If so, log it as a warning and continue
				if strings.Contains(err.Error(), "duplicate column name") ||
					strings.Contains(err.Error(), "duplicate column") ||
					strings.Contains(err.Error(), "already exists") {
					log.Warn().
						Str("migration", migration.Name()).
						Err(err).
						Msg("migration skipped - column/table already exists")
					continue
				}
				return fmt.Errorf("failed to run migration %s: %w", migration.Name(), err)
			}
		}
	}

	return nil
}
