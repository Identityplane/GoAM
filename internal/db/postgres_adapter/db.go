package postgres_adapter

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	Driver string // "postgres" or "pgx"
	DSN    string
}

func Init(cfg Config) (*pgx.Conn, error) {
	config, err := pgx.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	conn, err := pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	// Test the connection
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	// Excetute statement to check database oid and log it
	rows, err := conn.Query(context.Background(), `
		SELECT oid, datname 
		FROM pg_database 
		WHERE datname = current_database();
	`)
	if err != nil {
		return nil, fmt.Errorf("check database: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var oid uint32
		var datname string
		if err := rows.Scan(&oid, &datname); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		fmt.Printf("Database: %s (oid: %d)\n", datname, oid)
	}

	return conn, nil
}
