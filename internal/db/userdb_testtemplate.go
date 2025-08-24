package db

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func UserDBTests(t *testing.T, db UserDB) {
	t.Run("TestListUsersWithPagination", func(t *testing.T) {
		clearUserDB(t, db)
		TemplateTestListUsersWithPagination(t, db)
	})
	t.Run("TestGetUserStats", func(t *testing.T) {
		clearUserDB(t, db)
		TemplateTestGetUserStats(t, db)
	})
	t.Run("TestUserCRUD", func(t *testing.T) {
		clearUserDB(t, db)
		TemplateTestUserCRUD(t, db)
	})
	t.Run("TestUpdateUserDoesNotChangeOtherFields", func(t *testing.T) {
		clearUserDB(t, db)
		TemplateTestUpdateUserDoesNotChangeOtherFields(t, db)
	})
	t.Run("TestDeleteUser", func(t *testing.T) {
		clearUserDB(t, db)
		TemplateTestDeleteUser(t, db)
	})
	t.Run("TestGetUserByFederatedIdentifier", func(t *testing.T) {
		clearUserDB(t, db)
		TemplateTestGetUserByFederatedIdentifier(t, db)
	})
	t.Run("TestUpdateNonExistentUser", func(t *testing.T) {
		clearUserDB(t, db)
		TemplateTestUpdateNonExistentUser(t, db)
	})
}

func clearUserDB(t *testing.T, db UserDB) {

	users, err := db.ListUsers(context.Background(), "test-tenant", "test-realm")
	require.NoError(t, err)

	for _, user := range users {
		err := db.DeleteUser(context.Background(), "test-tenant", "test-realm", user.Username)
		require.NoError(t, err)
	}

}

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
			Username:  "active1-stats",
			Status:    "active",
			Email:     "active1@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "active2-stats",
			Status:    "active",
			Email:     "active2@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "inactive1-stats",
			Status:    "inactive",
			Email:     "inactive1@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Tenant:    testTenant,
			Realm:     testRealm,
			Username:  "locked1-stats",
			Status:    "locked",
			Email:     "locked1@example.com",
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
		Username:  "testuser-delete",
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

// This test checks that a user update does not have any other changes to the user than the changed fields
func TemplateTestUpdateUserDoesNotChangeOtherFields(t *testing.T, db UserDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create a userBefore
	user := model.User{
		Tenant:         testTenant,
		Realm:          testRealm,
		Username:       "testuser2",
		Status:         "active",
		TrustedDevices: []string{"device1", "device2"},
	}

	// Create the user
	err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	// Load user from db to ensure that id etc is initialized
	userBefore, err := db.GetUserByUsername(ctx, testTenant, testRealm, user.Username)
	require.NoError(t, err)

	fmt.Printf("userBefore updatedAt: %+v\n now=%+v\n", userBefore.UpdatedAt, time.Now())
	userBeforeUpdatedAt := userBefore.UpdatedAt
	userBefore.UpdatedAt = time.Time{}

	jsonBefore, _ := json.Marshal(userBefore)
	jsonBeforeString := string(jsonBefore)

	// Wait for 2 seconds to ensure that the updatedAt field is different
	time.Sleep(2 * time.Second)

	// Update the user
	err = db.UpdateUser(ctx, userBefore)
	require.NoError(t, err)

	userAfter, err := db.GetUserByUsername(ctx, testTenant, testRealm, userBefore.Username)
	require.NoError(t, err)

	fmt.Printf("userAfter  updatedAt: %+v\n now=%+v\n", userAfter.UpdatedAt, time.Now())

	// check that the updatedAt field is different but then changing it to nil to compare the json of the two user
	// to ensure nothing else changed
	assert.NotEqual(t, userBeforeUpdatedAt, userAfter.UpdatedAt)
	userAfter.UpdatedAt = time.Time{}

	jsonAfter, _ := json.Marshal(userAfter)
	jsonAfterString := string(jsonAfter)

	// Check that the json before and after are the same
	assert.Equal(t, jsonBeforeString, jsonAfterString)
}

// TemplateTestGetUserByFederatedIdentifier tests the GetUserByFederatedIdentifier functionality
func TemplateTestGetUserByFederatedIdentifier(t *testing.T, db UserDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create test user with federated identity
	testUser := model.User{
		Tenant:       testTenant,
		Realm:        testRealm,
		Username:     "federated-user",
		Status:       "active",
		Email:        "federated@example.com",
		FederatedIDP: stringPtr("google"),
		FederatedID:  stringPtr("google-123"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create the user
	err := db.CreateUser(ctx, testUser)
	require.NoError(t, err)

	t.Run("GetExistingFederatedUser", func(t *testing.T) {
		user, err := db.GetUserByFederatedIdentifier(ctx, testTenant, testRealm, "google", "google-123")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.Username, user.Username)
		assert.Equal(t, testUser.Email, user.Email)
		assert.Equal(t, testUser.FederatedIDP, user.FederatedIDP)
		assert.Equal(t, testUser.FederatedID, user.FederatedID)
	})

	t.Run("GetNonExistentFederatedUser", func(t *testing.T) {
		user, err := db.GetUserByFederatedIdentifier(ctx, testTenant, testRealm, "google", "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("GetUserWithDifferentProvider", func(t *testing.T) {
		user, err := db.GetUserByFederatedIdentifier(ctx, testTenant, testRealm, "github", "google-123")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

// TemplateTestUpdateNonExistentUser tests that an error is returned when trying to update a non-existent user
func TemplateTestUpdateNonExistentUser(t *testing.T, db UserDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create a non-existent user object (not saved to database)
	nonExistentUser := model.User{
		Tenant:          testTenant,
		Realm:           testRealm,
		Username:        "non-existent-user",
		Status:          "active",
		Email:           "nonexistent@example.com",
		LoginIdentifier: "nonexistent@example.com",
		TrustedDevices:  []string{"device1", "device2"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	t.Run("UpdateNonExistentUser", func(t *testing.T) {
		// Try to update a user that doesn't exist in the database
		err := db.UpdateUser(ctx, &nonExistentUser)
		assert.Error(t, err, "UpdateUser should return an error when updating a non-existent user")
	})

	t.Run("UpdateUserWithNonExistentID", func(t *testing.T) {
		// Create a user with a non-existent ID
		userWithNonExistentID := nonExistentUser
		userWithNonExistentID.ID = "non-existent-id-12345"

		// Try to update a user with a non-existent ID
		err := db.UpdateUser(ctx, &userWithNonExistentID)
		assert.Error(t, err, "UpdateUser should return an error when updating a user with non-existent ID")
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
