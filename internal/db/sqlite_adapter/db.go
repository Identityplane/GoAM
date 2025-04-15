package sqlite_adapter

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // sqlite driver
)

type Config struct {
	Driver string // "sqlite" or "pgx"
	DSN    string
}

func Init(cfg Config) (*sql.DB, error) {
	var err error
	database, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	if err = database.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return database, nil
}
