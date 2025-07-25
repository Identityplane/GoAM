package sqlite_adapter

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/Identityplane/GoAM/internal/db"

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

func TestUserDb(t *testing.T) {
	sqldb := setupTestDB(t)
	userDB, err := NewUserDB(sqldb)
	require.NoError(t, err)

	db.UserDBTests(t, userDB)
}
