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
			Tenant:            testTenant,
			Realm:             testRealm,
			Username:          "active1",
			Status:            "active",
			Email:             "active1@example.com",
			LoginIdentifier:   "active1@example.com",
			ProfilePictureURI: "https://example.com/active1.jpg",
			Entitlements:      []string{"read:users", "write:users"},
			Consent:           []string{"marketing", "analytics", "cookies"},
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			Tenant:            testTenant,
			Realm:             testRealm,
			Username:          "active2",
			Status:            "active",
			Email:             "active2@example.com",
			LoginIdentifier:   "active2@example.com",
			ProfilePictureURI: "https://example.com/active2.jpg",
			Entitlements:      []string{"read:users"},
			Consent:           []string{"marketing"},
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			Tenant:            testTenant,
			Realm:             testRealm,
			Username:          "inactive1",
			Status:            "inactive",
			Email:             "inactive1@example.com",
			LoginIdentifier:   "inactive1@example.com",
			ProfilePictureURI: "https://example.com/inactive1.jpg",
			Entitlements:      []string{},
			Consent:           []string{},
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
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
	assert.Equal(t, users[0].LoginIdentifier, listedUsers[0].LoginIdentifier)
	assert.Equal(t, users[0].ProfilePictureURI, listedUsers[0].ProfilePictureURI)
	assert.Equal(t, users[0].Entitlements, listedUsers[0].Entitlements)
	assert.Equal(t, users[0].Consent, listedUsers[0].Consent)
	assert.Equal(t, users[1].Username, listedUsers[1].Username)
	assert.Equal(t, users[1].LoginIdentifier, listedUsers[1].LoginIdentifier)
	assert.Equal(t, users[1].ProfilePictureURI, listedUsers[1].ProfilePictureURI)
	assert.Equal(t, users[1].Entitlements, listedUsers[1].Entitlements)
	assert.Equal(t, users[1].Consent, listedUsers[1].Consent)

	// List users with pagination of 2, offset 2
	listedUsers, err = db.ListUsersWithPagination(ctx, testTenant, testRealm, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 1, len(listedUsers))
	assert.Equal(t, users[2].Username, listedUsers[0].Username)
	assert.Equal(t, users[2].LoginIdentifier, listedUsers[0].LoginIdentifier)
	assert.Equal(t, users[2].ProfilePictureURI, listedUsers[0].ProfilePictureURI)
	assert.Equal(t, users[2].Entitlements, listedUsers[0].Entitlements)
	assert.Equal(t, users[2].Consent, listedUsers[0].Consent)
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
		Tenant:            testTenant,
		Realm:             testRealm,
		Username:          "testuser",
		Status:            "active",
		Email:             "test@example.com",
		LoginIdentifier:   "test@example.com",
		ProfilePictureURI: "https://example.com/test.jpg",
		Entitlements:      []string{"read:users", "write:users"},
		Consent:           []string{"marketing", "analytics"},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
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
		assert.Equal(t, testUser.LoginIdentifier, user.LoginIdentifier)
		assert.Equal(t, testUser.ProfilePictureURI, user.ProfilePictureURI)
		assert.Equal(t, testUser.Entitlements, user.Entitlements)
		assert.Equal(t, testUser.Consent, user.Consent)

		// Db should set a userid
		assert.NotEmpty(t, user.ID)
		// update id
		testUser.ID = user.ID
	})

	t.Run("GetUserByID", func(t *testing.T) {
		user, err := db.GetUserByID(ctx, testTenant, testRealm, testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.Username, user.Username)
		assert.Equal(t, testUser.LoginIdentifier, user.LoginIdentifier)
		assert.Equal(t, testUser.ProfilePictureURI, user.ProfilePictureURI)
		assert.Equal(t, testUser.Entitlements, user.Entitlements)
		assert.Equal(t, testUser.Consent, user.Consent)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		user, err := db.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		require.NoError(t, err)
		require.NotNil(t, user)

		// Update multiple fields
		user.Email = "updated@example.com"
		user.LoginIdentifier = "updated@example.com"
		user.ProfilePictureURI = "https://example.com/updated.jpg"
		user.Entitlements = []string{"read:users", "write:users", "admin:users"}
		user.Consent = []string{"marketing", "analytics", "cookies", "third-party"}

		err = db.UpdateUser(ctx, user)
		assert.NoError(t, err)

		updatedUser, err := db.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, "updated@example.com", updatedUser.Email)
		assert.Equal(t, "updated@example.com", updatedUser.LoginIdentifier)
		assert.Equal(t, "https://example.com/updated.jpg", updatedUser.ProfilePictureURI)
		assert.Equal(t, []string{"read:users", "write:users", "admin:users"}, updatedUser.Entitlements)
		assert.Equal(t, []string{"marketing", "analytics", "cookies", "third-party"}, updatedUser.Consent)
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		// First get the user to ensure we have the latest data
		user, err := db.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		require.NoError(t, err)
		require.NotNil(t, user)

		// Now test GetUserByEmail
		userByEmail, err := db.GetUserByEmail(ctx, testTenant, testRealm, user.Email)
		assert.NoError(t, err)
		assert.NotNil(t, userByEmail)
		assert.Equal(t, user.Username, userByEmail.Username)
		assert.Equal(t, user.Email, userByEmail.Email)
		assert.Equal(t, user.LoginIdentifier, userByEmail.LoginIdentifier)
		assert.Equal(t, user.ProfilePictureURI, userByEmail.ProfilePictureURI)
		assert.Equal(t, user.Entitlements, userByEmail.Entitlements)
		assert.Equal(t, user.Consent, userByEmail.Consent)
	})

	t.Run("GetUserByLoginIdentifier", func(t *testing.T) {
		// First get the user to ensure we have the latest data
		user, err := db.GetUserByUsername(ctx, testTenant, testRealm, testUser.Username)
		require.NoError(t, err)
		require.NotNil(t, user)

		// Now test GetUserByLoginIdentifier
		userByLoginID, err := db.GetUserByLoginIdentifier(ctx, testTenant, testRealm, user.LoginIdentifier)
		assert.NoError(t, err)
		assert.NotNil(t, userByLoginID)
		assert.Equal(t, user.Username, userByLoginID.Username)
		assert.Equal(t, user.Email, userByLoginID.Email)
		assert.Equal(t, user.LoginIdentifier, userByLoginID.LoginIdentifier)
		assert.Equal(t, user.ProfilePictureURI, userByLoginID.ProfilePictureURI)
		assert.Equal(t, user.Entitlements, userByLoginID.Entitlements)
		assert.Equal(t, user.Consent, userByLoginID.Consent)
	})

	t.Run("GetNonExistentUserByEmail", func(t *testing.T) {
		user, err := db.GetUserByEmail(ctx, testTenant, testRealm, "nonexistent@example.com")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("GetNonExistentUserByLoginIdentifier", func(t *testing.T) {
		user, err := db.GetUserByLoginIdentifier(ctx, testTenant, testRealm, "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, user)
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
