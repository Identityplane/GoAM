package sqlite_adapter

import (
	"context"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/require"
)

func TestUserAttributeDb(t *testing.T) {
	sqldb := setupTestDB(t)

	// Create both user DB and user attribute DB
	userDB, err := NewUserDB(sqldb)
	require.NoError(t, err)

	userAttributeDB, err := NewUserAttributeDB(sqldb)
	require.NoError(t, err)

	// Create a test user first
	ctx := context.Background()
	testUser := model.User{
		ID:        "123e4567-e89b-12d3-a456-426614174000",
		Tenant:    "test-tenant",
		Realm:     "test-realm",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = userDB.CreateUser(ctx, testUser)
	require.NoError(t, err)

	// Now run the user attribute tests
	db.UserAttributeDBTests(t, userAttributeDB)

	// Clean up the test user
	err = userDB.DeleteUser(ctx, "test-tenant", "test-realm", "test-user")
	require.NoError(t, err)
}
