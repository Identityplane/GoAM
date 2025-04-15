package sqlite_adapter

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"goiam/internal/db/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

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

func TestUserCRUD(t *testing.T) {
	// print current pwd for debugging
	pwd, err := os.Getwd()
	fmt.Println("Current working directory:", pwd)

	// Setup test database
	db, err := Open(":memory:?_foreign_keys=on")
	require.NoError(t, err)
	defer db.Close()

	err = Migrate(db)
	require.NoError(t, err)

	// Create user DB with test tenant and realm
	userDB := NewSQLiteUserDB(db)
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
