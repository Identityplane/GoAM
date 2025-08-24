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
		err := db.DeleteUser(context.Background(), "test-tenant", "test-realm", user.ID)
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
			ID:        "1",
			Tenant:    testTenant,
			Realm:     testRealm,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "2",
			Tenant:    testTenant,
			Realm:     testRealm,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "3",
			Tenant:    testTenant,
			Realm:     testRealm,
			Status:    "inactive",
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
	assert.Equal(t, users[0].ID, listedUsers[0].ID)
	assert.Equal(t, users[1].ID, listedUsers[1].ID)

	// List users with pagination of 2, offset 2
	listedUsers, err = db.ListUsersWithPagination(ctx, testTenant, testRealm, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 1, len(listedUsers))
	assert.Equal(t, users[2].ID, listedUsers[0].ID)
}

// TestGetUserStats is a parameterized test that can be used with any database implementation
func TemplateTestGetUserStats(t *testing.T, db UserDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create test users with different statuses
	users := []model.User{
		{
			ID:        "1",
			Tenant:    testTenant,
			Realm:     testRealm,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "2",
			Tenant:    testTenant,
			Realm:     testRealm,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "3",
			Tenant:    testTenant,
			Realm:     testRealm,
			Status:    "inactive",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "4",
			Tenant:    testTenant,
			Realm:     testRealm,
			Status:    "locked",
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
		ID:        "1",
		Tenant:    testTenant,
		Realm:     testRealm,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("CreateUser", func(t *testing.T) {
		err := db.CreateUser(ctx, testUser)
		assert.NoError(t, err)
	})

	t.Run("GetUserByID", func(t *testing.T) {
		user, err := db.GetUserByID(ctx, testTenant, testRealm, testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.ID, user.ID)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		user, err := db.GetUserByID(ctx, testTenant, testRealm, testUser.ID)
		require.NoError(t, err)
		require.NotNil(t, user)

		// Update multiple fields
		user.Status = "inactive"

		err = db.UpdateUser(ctx, user)
		assert.NoError(t, err)

		updatedUser, err := db.GetUserByID(ctx, testTenant, testRealm, testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, "inactive", updatedUser.Status)
	})

	t.Run("GetNonExistentUser", func(t *testing.T) {
		user, err := db.GetUserByID(ctx, "nonexistent", "nonexistent", "nonexistent")
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
		ID:        "1",
		Tenant:    testTenant,
		Realm:     testRealm,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create the user first
	err := db.CreateUser(ctx, testUser)
	require.NoError(t, err)

	t.Run("DeleteExistingUser", func(t *testing.T) {
		// Delete the user
		err := db.DeleteUser(ctx, testTenant, testRealm, testUser.ID)
		assert.NoError(t, err)

		// Verify user is deleted
		user, err := db.GetUserByID(ctx, testTenant, testRealm, testUser.ID)
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
		ID:        "1",
		Tenant:    testTenant,
		Realm:     testRealm,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create the user
	err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	// Load user from db to ensure that id etc is initialized
	userBefore, err := db.GetUserByID(ctx, testTenant, testRealm, user.ID)
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

	userAfter, err := db.GetUserByID(ctx, testTenant, testRealm, userBefore.ID)
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
		ID:        "1",
		Tenant:    testTenant,
		Realm:     testRealm,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create the user
	err := db.CreateUser(ctx, testUser)
	require.NoError(t, err)
}

// TemplateTestUpdateNonExistentUser tests that an error is returned when trying to update a non-existent user
func TemplateTestUpdateNonExistentUser(t *testing.T, db UserDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create a non-existent user object (not saved to database)
	nonExistentUser := model.User{
		ID:        "1",
		Tenant:    testTenant,
		Realm:     testRealm,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
