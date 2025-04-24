package sqlite_adapter

import (
	"database/sql"
	"fmt"
	"goiam/internal/db"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Open(dsn string) (*sql.DB, error) {
	return sql.Open("sqlite", dsn)
}

func setupTestDB(t *testing.T) *sql.DB {
	// print current pwd for debugging
	pwd, err := os.Getwd()
	fmt.Println("Current working directory:", pwd)

	// Setup test database
	db, err := Open(":memory:?_foreign_keys=on")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	err = RunMigrations(db)
	require.NoError(t, err)

	return db
}

func TestListUsersWithPagination(t *testing.T) {
	sqldb := setupTestDB(t)
	userDB, err := NewUserDB(sqldb)
	require.NoError(t, err)

	db.TemplateTestListUsersWithPagination(t, userDB)
}

func TestGetUserStats(t *testing.T) {
	sqldb := setupTestDB(t)
	userDB, err := NewUserDB(sqldb)
	require.NoError(t, err)

	db.TemplateTestGetUserStats(t, userDB)
}

func TestUserCRUD(t *testing.T) {
	sqldb := setupTestDB(t)
	userDB, err := NewUserDB(sqldb)
	require.NoError(t, err)

	db.TemplateTestUserCRUD(t, userDB)
}

func TestSQLiteUserDB_DeleteUser(t *testing.T) {
	sqldb := setupTestDB(t)
	userDB, err := NewUserDB(sqldb)
	require.NoError(t, err)

	db.TemplateTestDeleteUser(t, userDB)
}
