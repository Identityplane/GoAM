package postgres_adapter

import (
	"context"
	"embed"
	"fmt"
	"goiam/internal/logger"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

func setupTestDB(t *testing.T) (*pgx.Conn, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "goiam",
			"POSTGRES_PASSWORD": "secret123",
			"POSTGRES_DB":       "goiamdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get the container's host and port
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)
	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create the connection string
	dsn := fmt.Sprintf("postgres://goiam:secret123@%s:%s/goiamdb?sslmode=disable", host, port.Port())

	// Wait here for 1 sec to ensure the database is ready
	time.Sleep(1 * time.Second)

	// Connect to the database
	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)

	// Run migrations
	err = RunMigrations(conn)
	require.NoError(t, err)

	return conn, nil
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Uses embed fs to load migrations and run them over the database connection in order
func RunMigrations(conn *pgx.Conn) error {

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
			logger.DebugNoContext("Running migration: %s", migration.Name())

			// run migration
			_, err = conn.Exec(context.Background(), string(migrationFile))
			if err != nil {
				return fmt.Errorf("failed to run migration: %w", err)
			}
		}
	}

	return nil
}
