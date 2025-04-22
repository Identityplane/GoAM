package db

import (
	"context"
	"testing"
	"time"

	"goiam/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test for user listing
func TemplateTestListUsersWithPagination(t *testing.T, db UserDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create test users
	users := []model.User{
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "active1",
			Status:    "active",
			Email:     "active1@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "active2",
			Status:    "active",
			Email:     "active2@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "inactive1",
			Status:    "inactive",
			Email:     "inactive1@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Create users
	for _, user := range users {
		err := db.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	// List users with pagination of 2
	listedUsers, err := db.ListUsersWithPagination(ctx, testTenant, testRealm, 0, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, len(listedUsers))
	assert.Equal(t, users[0].Username, listedUsers[0].Username)
	assert.Equal(t, users[1].Username, listedUsers[1].Username)

	// List users with pagination of 2, offset 2
	listedUsers, err = db.ListUsersWithPagination(ctx, testTenant, testRealm, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 1, len(listedUsers))
	assert.Equal(t, users[2].Username, listedUsers[0].Username)

}

// TestGetUserStats is a parameterized test that can be used with any database implementation
func TemplateTestGetUserStats(t *testing.T, db UserDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create test users with different statuses
	users := []model.User{
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "active1",
			Status:    "active",
			Email:     "active1@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "active2",
			Status:    "active",
			Email:     "active2@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "inactive1",
			Status:    "inactive",
			Email:     "inactive1@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "locked",
			Status:    "locked",
			Email:     "locked@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Create users
	for _, user := range users {
		err := db.CreateUser(ctx, user)
		require.NoError(t, err)
	}

	// Get stats
	stats, err := db.GetUserStats(ctx, testTenant, testRealm)
	require.NoError(t, err)

	// Verify stats
	assert.Equal(t, int64(4), stats.TotalUsers)
	assert.Equal(t, int64(2), stats.ActiveUsers)
	assert.Equal(t, int64(1), stats.InactiveUsers)
	assert.Equal(t, int64(1), stats.LockedUsers)
}

// TemplateTestUserCRUD is a parameterized test for basic CRUD operations
func TemplateTestUserCRUD(t *testing.T, db UserDB) {
	ctx := context.Background()
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

	t.Run("CreateUser", func(t *testing.T) {
		err := db.CreateUser(ctx, testUser)
		assert.NoError(t, err)
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		user, err := db.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.Username, user.Username)
		assert.Equal(t, testUser.Email, user.Email)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		user, err := db.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		require.NoError(t, err)
		require.NotNil(t, user)

		user.Email = "updated@example.com"
		err = db.UpdateUser(ctx, user)
		assert.NoError(t, err)

		updatedUser, err := db.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, "updated@example.com", updatedUser.Email)
	})

	t.Run("GetNonExistentUser", func(t *testing.T) {
		user, err := db.GetUserByUsername(ctx, "nonexistent", "nonexistent", "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

// TemplateTestDeleteUser is a parameterized test for user deletion
func TemplateTestDeleteUser(t *testing.T, db UserDB) {
	ctx := context.Background()
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

	// Create the user first
	err := db.CreateUser(ctx, testUser)
	require.NoError(t, err)

	t.Run("DeleteExistingUser", func(t *testing.T) {
		// Delete the user
		err := db.DeleteUser(ctx, testTenant, testRealm, testUser.Username)
		assert.NoError(t, err)

		// Verify user is deleted
		user, err := db.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("DeleteNonExistentUser", func(t *testing.T) {
		// Try to delete non-existent user
		err := db.DeleteUser(ctx, testTenant, testRealm, "nonexistent")
		assert.NoError(t, err) // Should be idempotent
	})
}
