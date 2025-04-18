package sqlite_adapter

import (
	"database/sql"
	"fmt"
	"goiam/internal/db/model"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Open(dsn string) (*sql.DB, error) {
	return sql.Open("sqlite", dsn)
}

func Migrate(db *sql.DB) error {
	migrationSQL, err := os.ReadFile("migrations/001_create_users.up.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(migrationSQL))
	return err
}

func setupTestDB(t *testing.T) *SQLiteUserDB {
	// print current pwd for debugging
	pwd, err := os.Getwd()
	fmt.Println("Current working directory:", pwd)

	// Setup test database
	db, err := Open(":memory:?_foreign_keys=on")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	err = Migrate(db)
	require.NoError(t, err)

	// Create user DB with test tenant and realm
	userDB, err := NewSQLiteUserDB(db)
	require.NoError(t, err)

	return userDB
}

func TestListUsersWithPagination(t *testing.T) {
	userDB := setupTestDB(t)
	model.TemplateTestUserCRUD(t, userDB)
}

func TestGetUserStats(t *testing.T) {
	userDB := setupTestDB(t)
	model.TemplateTestGetUserStats(t, userDB)
}

func TestUserCRUD(t *testing.T) {
	userDB := setupTestDB(t)
	model.TemplateTestUserCRUD(t, userDB)
}

func TestSQLiteUserDB_DeleteUser(t *testing.T) {

	db := setupTestDB(t)
	model.TemplateTestDeleteUser(t, db)
}
