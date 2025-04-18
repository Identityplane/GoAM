package postgres_adapter

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"goiam/internal/db/model"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *pgx.Conn

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
	migrationSQL, err := os.ReadFile("migrations/001_create_users.up.sql")
	require.NoError(t, err)

	_, err = conn.Exec(ctx, string(migrationSQL))
	require.NoError(t, err)

	return conn, nil
}

func TestListUsersWithPagination(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close(context.Background())

	userDB, err := NewPostgresUserDB(conn)
	require.NoError(t, err)
	model.TemplateTestListUsersWithPagination(t, userDB)
}

func TestUserCRUD(t *testing.T) {
	// Setup test database
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close(context.Background())

	// Create user DB with test tenant and realm
	userDB, err := NewPostgresUserDB(conn)
	require.NoError(t, err)

	// Run the CRUD tests
	model.TemplateTestUserCRUD(t, userDB)
}

func TestGetUserStats(t *testing.T) {
	// Setup test database
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close(context.Background())

	// Create user DB with test tenant and realm
	userDB, err := NewPostgresUserDB(conn)
	require.NoError(t, err)

	// Run the stats test
	model.TemplateTestGetUserStats(t, userDB)
}

func TestPostgresUserDB_DeleteUser(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close(context.Background())

	userDB, err := NewPostgresUserDB(conn)
	require.NoError(t, err)

	model.TemplateTestDeleteUser(t, userDB)
}
