package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // sqlite driver
)

var DB *sql.DB

type Config struct {
	Driver string // "sqlite" or "pgx"
	DSN    string
}

func Init(cfg Config) error {
	var err error
	DB, err = sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	return nil
}
