package postgres_adapter

import (
	"testing"

	"goiam/internal/db"

	"github.com/stretchr/testify/require"
)

func TestListUsersWithPagination(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	userDB, err := NewPostgresUserDB(conn)
	require.NoError(t, err)
	db.TemplateTestListUsersWithPagination(t, userDB)
}

func TestUserCRUD(t *testing.T) {
	// Setup test database
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	// Create user DB with test tenant and realm
	userDB, err := NewPostgresUserDB(conn)
	require.NoError(t, err)

	// Run the CRUD tests
	db.TemplateTestUserCRUD(t, userDB)
}

func TestGetUserStats(t *testing.T) {
	// Setup test database
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	// Create user DB with test tenant and realm
	userDB, err := NewPostgresUserDB(conn)
	require.NoError(t, err)

	// Run the stats test
	db.TemplateTestGetUserStats(t, userDB)
}

func TestPostgresUserDB_DeleteUser(t *testing.T) {
	conn, err := setupTestDB(t)
	require.NoError(t, err)
	defer conn.Close()

	userDB, err := NewPostgresUserDB(conn)
	require.NoError(t, err)

	db.TemplateTestDeleteUser(t, userDB)
}
