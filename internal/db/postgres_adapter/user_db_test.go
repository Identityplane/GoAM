package postgres_adapter

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"goiam/internal/db/model"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
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

func TestUserCRUD(t *testing.T) {
	// Setup test database
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close(context.Background())

	// Create user DB with test tenant and realm
	userDB, err := NewPostgresUserDB(conn)
	require.NoError(t, err)

	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create test user
	testUser := model.User{
		Tenant:    testTenant,
		Realm:     testRealm,
		Username:  "testuser",
		Status:    "active",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx := context.Background()

	t.Run("CreateUser", func(t *testing.T) {
		err := userDB.CreateUser(ctx, testUser)
		assert.NoError(t, err)
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		user, err := userDB.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.Username, user.Username)
		assert.Equal(t, testUser.Email, user.Email)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		user, err := userDB.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		require.NoError(t, err)
		require.NotNil(t, user)

		user.Email = "updated@example.com"
		err = userDB.UpdateUser(ctx, user)
		assert.NoError(t, err)

		updatedUser, err := userDB.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, "updated@example.com", updatedUser.Email)
	})

	t.Run("GetNonExistentUser", func(t *testing.T) {
		user, err := userDB.GetUserByUsername(ctx, "nonexistent", "nonexistent", "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}
