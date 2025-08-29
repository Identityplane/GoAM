package db

import (
	"context"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func UserAttributeDBTests(t *testing.T, db UserAttributeDB) {
	t.Run("TestUserAttributeCRUD", func(t *testing.T) {
		clearUserAttributeDB(t, db)
		TemplateTestUserAttributeCRUD(t, db)
	})
	t.Run("TestMultipleAttributesOfSameType", func(t *testing.T) {
		clearUserAttributeDB(t, db)
		TemplateTestMultipleAttributesOfSameType(t, db)
	})
	t.Run("TestGetUserWithAttributes", func(t *testing.T) {
		clearUserAttributeDB(t, db)
		TemplateTestGetUserWithAttributes(t, db)
	})
	t.Run("TestGetUserByAttributeIndexWithAttributes", func(t *testing.T) {
		clearUserAttributeDB(t, db)
		TemplateTestGetUserByAttributeIndexWithAttributes(t, db)
	})
	t.Run("TestCreateUserWithAttributes", func(t *testing.T) {
		clearUserAttributeDB(t, db)
		TemplateTestCreateUserWithAttributes(t, db)
	})
	t.Run("TestUpdateUserWithAttributes", func(t *testing.T) {
		clearUserAttributeDB(t, db)
		TemplateTestUpdateUserWithAttributes(t, db)
	})
	t.Run("TestIndexUniqueConstraint", func(t *testing.T) {
		clearUserAttributeDB(t, db)
		TemplateTestIndexUniqueConstraint(t, db)
	})
}

func clearUserAttributeDB(t *testing.T, db UserAttributeDB) {
	// This function would need to be implemented based on the specific database implementation
	// For now, we'll assume the test database is clean or provide a way to clean it
	// In a real implementation, you might want to add a ClearAll method to the interface
}

// TemplateTestUserAttributeCRUD is a parameterized test for basic CRUD operations
func TemplateTestUserAttributeCRUD(t *testing.T, db UserAttributeDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	testUserID := "123e4567-e89b-12d3-a456-426614174000"

	// Create test user attribute
	testAttribute := model.UserAttribute{
		UserID: testUserID,
		Tenant: testTenant,
		Realm:  testRealm,
		Index:  stringPtr("test@example.com"),
		Type:   "email",
		Value: model.EmailAttributeValue{
			Email:      "test@example.com",
			Verified:   false,
			VerifiedAt: nil,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("CreateUserAttribute", func(t *testing.T) {
		err := db.CreateUserAttribute(ctx, testAttribute)
		assert.NoError(t, err)
	})

	t.Run("ListUserAttributes", func(t *testing.T) {
		attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
		assert.NoError(t, err)
		assert.Len(t, attrs, 1)
		assert.Equal(t, "email", attrs[0].Type)
		if attrs[0].Index != nil {
			assert.Equal(t, "test@example.com", *attrs[0].Index)
		}

		// Update the test attribute with the ID from the database
		testAttribute.ID = attrs[0].ID
	})

	t.Run("GetUserAttributeByID", func(t *testing.T) {
		// First get the attribute to get its ID
		attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
		require.NoError(t, err)
		require.Len(t, attrs, 1)

		attr, err := db.GetUserAttributeByID(ctx, testTenant, testRealm, attrs[0].ID)
		assert.NoError(t, err)
		assert.NotNil(t, attr)
		assert.Equal(t, testAttribute.UserID, attr.UserID)
		assert.Equal(t, testAttribute.Index, attr.Index)
		assert.Equal(t, testAttribute.Type, attr.Type)
	})

	t.Run("UpdateUserAttribute", func(t *testing.T) {
		// Get the attribute first
		attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
		require.NoError(t, err)
		require.Len(t, attrs, 1)

		attr, err := db.GetUserAttributeByID(ctx, testTenant, testRealm, attrs[0].ID)
		require.NoError(t, err)
		require.NotNil(t, attr)

		// Update the attribute
		attr.Value = model.EmailAttributeValue{
			Email:      "test@example.com",
			Verified:   true,
			VerifiedAt: timePtr(time.Now()),
		}

		err = db.UpdateUserAttribute(ctx, attr)
		assert.NoError(t, err)

		// Verify the update
		updatedAttr, err := db.GetUserAttributeByID(ctx, testTenant, testRealm, attrs[0].ID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedAttr)

		if emailValue, ok := updatedAttr.Value.(map[string]interface{}); ok {
			assert.Equal(t, true, emailValue["verified"])
			assert.NotNil(t, emailValue["verified_at"])
		} else {
			t.Errorf("Failed to type assert Value to map[string]interface{}")
		}
	})

	t.Run("DeleteUserAttribute", func(t *testing.T) {
		// Get the attribute first to get its ID
		attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
		require.NoError(t, err)
		require.Len(t, attrs, 1)

		// Delete the attribute
		err = db.DeleteUserAttribute(ctx, testTenant, testRealm, attrs[0].ID)
		assert.NoError(t, err)

		// Verify it's deleted
		deletedAttr, err := db.GetUserAttributeByID(ctx, testTenant, testRealm, attrs[0].ID)
		assert.NoError(t, err)
		assert.Nil(t, deletedAttr)

		// Verify it's not in the list
		remainingAttrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
		assert.NoError(t, err)
		assert.Len(t, remainingAttrs, 0)
	})
}

// TemplateTestMultipleAttributesOfSameType tests multiple attributes of the same type
func TemplateTestMultipleAttributesOfSameType(t *testing.T, db UserAttributeDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	testUserID := "123e4567-e89b-12d3-a456-426614174000"

	// Create multiple email attributes
	emailAttrs := []model.UserAttribute{
		{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("primary@example.com"),
			Type:   "email",
			Value: model.EmailAttributeValue{
				Email:    "primary@example.com",
				Verified: true,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("work@example.com"),
			Type:   "email",
			Value: model.EmailAttributeValue{
				Email:    "work@example.com",
				Verified: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Create all attributes
	for _, attr := range emailAttrs {
		err := db.CreateUserAttribute(ctx, attr)
		require.NoError(t, err)
	}

	t.Run("MultipleAttributesOfSameType", func(t *testing.T) {
		attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
		assert.NoError(t, err)
		assert.Len(t, attrs, 2)

		// Verify both attributes exist
		indexes := make(map[string]bool)
		for _, attr := range attrs {
			if attr.Index != nil {
				indexes[*attr.Index] = true
			}
		}
		assert.True(t, indexes["primary@example.com"])
		assert.True(t, indexes["work@example.com"])
	})

	// Clean up
	attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
	require.NoError(t, err)
	for _, attr := range attrs {
		db.DeleteUserAttribute(ctx, testTenant, testRealm, attr.ID)
	}
}

// TemplateTestGetUserWithAttributes tests getting a user with all their attributes
func TemplateTestGetUserWithAttributes(t *testing.T, db UserAttributeDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	testUserID := "123e4567-e89b-12d3-a456-426614174000"

	// Create multiple attributes for the user
	attributes := []model.UserAttribute{
		{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("primary@example.com"),
			Type:   "email",
			Value: model.EmailAttributeValue{
				Email:    "primary@example.com",
				Verified: true,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("work@example.com"),
			Type:   "email",
			Value: model.EmailAttributeValue{
				Email:    "work@example.com",
				Verified: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("+1234567890"),
			Type:   "phone",
			Value: model.PhoneAttributeValue{
				Phone:    "+1234567890",
				Verified: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Create all attributes
	for _, attr := range attributes {
		err := db.CreateUserAttribute(ctx, attr)
		require.NoError(t, err)
	}

	t.Run("GetUserWithAttributes", func(t *testing.T) {
		user, err := db.GetUserWithAttributes(ctx, testTenant, testRealm, testUserID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUserID, user.ID)
		assert.Equal(t, testTenant, user.Tenant)
		assert.Equal(t, testRealm, user.Realm)

		// Verify all attributes are loaded
		assert.Len(t, user.UserAttributes, 3)

		// Verify email attributes using the helper method
		emailAttrs := user.GetAttributesByType("email")
		assert.Len(t, emailAttrs, 2)

		// Verify phone attribute using the helper method
		phoneAttrs := user.GetAttributesByType("phone")
		assert.Len(t, phoneAttrs, 1)
		if phoneAttrs[0].Index != nil {
			assert.Equal(t, "+1234567890", *phoneAttrs[0].Index)
		}

		// Test the generic GetAttribute method
		emailAttr, _, err := model.GetAttribute[model.EmailAttributeValue](user, "email")
		assert.Error(t, err)
		assert.Nil(t, emailAttr)
		assert.Contains(t, err.Error(), "multiple attributes of type 'email' found")
		assert.Contains(t, err.Error(), "use GetAttributesByType instead")

		// Test the new GetAttributes method for multiple email attributes
		emailValues, _, err := model.GetAttributes[model.EmailAttributeValue](user, "email")
		assert.NoError(t, err)
		assert.Len(t, emailValues, 2)

		// Verify the converted values
		assert.Equal(t, "primary@example.com", emailValues[0].Email)
		assert.True(t, emailValues[0].Verified)
		assert.Equal(t, "work@example.com", emailValues[1].Email)
		assert.False(t, emailValues[1].Verified)

		// Test GetAttributes for phone (single attribute)
		phoneValues, _, err := model.GetAttributes[model.PhoneAttributeValue](user, "phone")
		assert.NoError(t, err)
		assert.Len(t, phoneValues, 1)
		assert.Equal(t, "+1234567890", phoneValues[0].Phone)
		assert.False(t, phoneValues[0].Verified)

		// Test GetAttributes for non-existent type
		nonexistentValues, _, err := model.GetAttributes[model.EmailAttributeValue](user, "nonexistent")
		assert.NoError(t, err)
		assert.Len(t, nonexistentValues, 0)
	})

	// Clean up
	attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
	require.NoError(t, err)
	for _, attr := range attrs {
		db.DeleteUserAttribute(ctx, testTenant, testRealm, attr.ID)
	}
}

// TemplateTestGetUserByAttributeIndexWithAttributes tests finding a user by attribute index with all attributes
func TemplateTestGetUserByAttributeIndexWithAttributes(t *testing.T, db UserAttributeDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	testUserID := "123e4567-e89b-12d3-a456-426614174000"

	// Create multiple attributes for the user
	attributes := []model.UserAttribute{
		{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("primary@example.com"),
			Type:   "email",
			Value: model.EmailAttributeValue{
				Email:    "primary@example.com",
				Verified: true,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("work@example.com"),
			Type:   "email",
			Value: model.EmailAttributeValue{
				Email:    "work@example.com",
				Verified: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("+1234567890"),
			Type:   "phone",
			Value: model.PhoneAttributeValue{
				Phone:    "+1234567890",
				Verified: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Create all attributes
	for _, attr := range attributes {
		err := db.CreateUserAttribute(ctx, attr)
		require.NoError(t, err)
	}

	t.Run("GetUserByEmailIndex", func(t *testing.T) {
		// Find user by primary email index
		user, err := db.GetUserByAttributeIndexWithAttributes(ctx, testTenant, testRealm, "email", "primary@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUserID, user.ID)
		assert.Equal(t, testTenant, user.Tenant)
		assert.Equal(t, testRealm, user.Realm)

		// Verify all attributes are loaded (not just email)
		assert.Len(t, user.UserAttributes, 3)

		// Verify we can find attributes of different types using helper methods
		emailAttrs := user.GetAttributesByType("email")
		assert.Len(t, emailAttrs, 2)

		phoneAttrs := user.GetAttributesByType("phone")
		assert.Len(t, phoneAttrs, 1)

		// Test the new GetAttributes method for multiple email attributes
		emailValues, _, err := model.GetAttributes[model.EmailAttributeValue](user, "email")
		assert.NoError(t, err)
		assert.Len(t, emailValues, 2)

		// Verify the converted values
		assert.Equal(t, "primary@example.com", emailValues[0].Email)
		assert.True(t, emailValues[0].Verified)
		assert.Equal(t, "work@example.com", emailValues[1].Email)
		assert.False(t, emailValues[1].Verified)

		// Test GetAttributes for phone (single attribute)
		phoneValues, _, err := model.GetAttributes[model.PhoneAttributeValue](user, "phone")
		assert.NoError(t, err)
		assert.Len(t, phoneValues, 1)
		assert.Equal(t, "+1234567890", phoneValues[0].Phone)
		assert.False(t, phoneValues[0].Verified)

		// Test GetAttributes for non-existent type
		nonexistentValues, _, err := model.GetAttributes[model.EmailAttributeValue](user, "nonexistent")
		assert.NoError(t, err)
		assert.Len(t, nonexistentValues, 0)

		// Test that GetAttribute fails when there are multiple attributes of the same type
		emailAttr, _, err := model.GetAttribute[model.EmailAttributeValue](user, "email")
		assert.Error(t, err)
		assert.Nil(t, emailAttr)
		assert.Contains(t, err.Error(), "multiple attributes of type 'email' found")
		assert.Contains(t, err.Error(), "use GetAttributesByType instead")

		// Test that GetAttribute works for single attributes
		phoneAttr, _, err := model.GetAttribute[model.PhoneAttributeValue](user, "phone")
		assert.NoError(t, err)
		assert.NotNil(t, phoneAttr)
		assert.Equal(t, "+1234567890", phoneAttr.Phone)
	})

	t.Run("GetUserByPhoneIndex", func(t *testing.T) {
		// Find user by phone index
		user, err := db.GetUserByAttributeIndexWithAttributes(ctx, testTenant, testRealm, "phone", "+1234567890")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUserID, user.ID)

		// Verify all attributes are loaded
		assert.Len(t, user.UserAttributes, 3)
	})

	t.Run("GetUserByNonExistentIndex", func(t *testing.T) {
		// Try to find user by non-existent index
		user, err := db.GetUserByAttributeIndexWithAttributes(ctx, testTenant, testRealm, "email", "nonexistent@example.com")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	// Clean up
	attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
	require.NoError(t, err)
	for _, attr := range attrs {
		db.DeleteUserAttribute(ctx, testTenant, testRealm, attr.ID)
	}
}

// TemplateTestCreateUserWithAttributes tests creating a user with multiple attributes
func TemplateTestCreateUserWithAttributes(t *testing.T, db UserAttributeDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	testUserID := "123e4567-e89b-12d3-a456-426614174001" // Use a different UUID from other tests

	// Create a test user with multiple attributes
	testUser := &model.User{
		ID:     testUserID,
		Tenant: testTenant,
		Realm:  testRealm,
		Status: "active",
		UserAttributes: []model.UserAttribute{
			{
				UserID: testUserID,
				Tenant: testTenant,
				Realm:  testRealm,
				Index:  stringPtr("primary@example.com"),
				Type:   "email",
				Value: model.EmailAttributeValue{
					Email:    "primary@example.com",
					Verified: true,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				UserID: testUserID,
				Tenant: testTenant,
				Realm:  testRealm,
				Index:  stringPtr("work@example.com"),
				Type:   "email",
				Value: model.EmailAttributeValue{
					Email:    "work@example.com",
					Verified: false,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				UserID: testUserID,
				Tenant: testTenant,
				Realm:  testRealm,
				Index:  stringPtr("+1234567890"),
				Type:   "phone",
				Value: model.PhoneAttributeValue{
					Phone:    "+1234567890",
					Verified: false,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Create the user and all attributes
	err := db.CreateUserWithAttributes(ctx, testUser)
	require.NoError(t, err, "Failed to create user with attributes")

	t.Run("GetUserWithAttributes", func(t *testing.T) {
		user, err := db.GetUserWithAttributes(ctx, testTenant, testRealm, testUserID)
		require.NoError(t, err, "Failed to get user with attributes")
		require.NotNil(t, user, "User should not be nil")
		assert.Equal(t, testUserID, user.ID)
		assert.Equal(t, testTenant, user.Tenant)
		assert.Equal(t, testRealm, user.Realm)

		// Verify all attributes are loaded
		assert.Len(t, user.UserAttributes, 3)

		// Verify email attributes using the helper method
		emailAttrs := user.GetAttributesByType("email")
		assert.Len(t, emailAttrs, 2)

		// Verify phone attribute using the helper method
		phoneAttrs := user.GetAttributesByType("phone")
		assert.Len(t, phoneAttrs, 1)
		if phoneAttrs[0].Index != nil {
			assert.Equal(t, "+1234567890", *phoneAttrs[0].Index)
		}

		// Test the generic GetAttribute method
		emailAttr, _, err := model.GetAttribute[model.EmailAttributeValue](user, "email")
		assert.Error(t, err)
		assert.Nil(t, emailAttr)
		assert.Contains(t, err.Error(), "multiple attributes of type 'email' found")
		assert.Contains(t, err.Error(), "use GetAttributesByType instead")

		// Test the new GetAttributes method for multiple email attributes
		emailValues, _, err := model.GetAttributes[model.EmailAttributeValue](user, "email")
		assert.NoError(t, err)
		assert.Len(t, emailValues, 2)

		// Verify the converted values
		assert.Equal(t, "primary@example.com", emailValues[0].Email)
		assert.True(t, emailValues[0].Verified)
		assert.Equal(t, "work@example.com", emailValues[1].Email)
		assert.False(t, emailValues[1].Verified)

		// Test GetAttributes for phone (single attribute)
		phoneValues, _, err := model.GetAttributes[model.PhoneAttributeValue](user, "phone")
		assert.NoError(t, err)
		assert.Len(t, phoneValues, 1)
		assert.Equal(t, "+1234567890", phoneValues[0].Phone)
		assert.False(t, phoneValues[0].Verified)

		// Test GetAttributes for non-existent type
		nonexistentValues, _, err := model.GetAttributes[model.EmailAttributeValue](user, "nonexistent")
		assert.NoError(t, err)
		assert.Len(t, nonexistentValues, 0)
	})

	// Clean up
	attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
	require.NoError(t, err)
	for _, attr := range attrs {
		db.DeleteUserAttribute(ctx, testTenant, testRealm, attr.ID)
	}
}

// TemplateTestUpdateUserWithAttributes tests updating a user with modified attributes
func TemplateTestUpdateUserWithAttributes(t *testing.T, db UserAttributeDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	testUserID := "123e4567-e89b-12d3-a456-426614174003" // Use a different UUID from other tests

	// Create a test user with multiple attributes
	testUser := &model.User{
		ID:     testUserID,
		Tenant: testTenant,
		Realm:  testRealm,
		Status: "active",
		UserAttributes: []model.UserAttribute{
			{
				UserID: testUserID,
				Tenant: testTenant,
				Realm:  testRealm,
				Index:  stringPtr("primary@example.com"),
				Type:   "email",
				Value: model.EmailAttributeValue{
					Email:    "primary@example.com",
					Verified: false,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				UserID: testUserID,
				Tenant: testTenant,
				Realm:  testRealm,
				Index:  stringPtr("+1234567890"),
				Type:   "phone",
				Value: model.PhoneAttributeValue{
					Phone:    "+1234567890",
					Verified: false,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Create the user and all attributes
	err := db.CreateUserWithAttributes(ctx, testUser)
	require.NoError(t, err, "Failed to create user with attributes")

	// Get the created user to have the actual timestamps and IDs
	createdUser, err := db.GetUserWithAttributes(ctx, testTenant, testRealm, testUserID)
	require.NoError(t, err, "Failed to get created user")
	require.Len(t, createdUser.UserAttributes, 2, "User should have 2 attributes")

	// Store original timestamps for comparison
	originalEmailUpdatedAt := createdUser.UserAttributes[0].UpdatedAt
	originalPhoneUpdatedAt := createdUser.UserAttributes[1].UpdatedAt

	// Wait a bit to ensure timestamp differences
	time.Sleep(1 * time.Second)

	// Prepare the updated user
	updatedUser := &model.User{
		ID:     testUserID,
		Tenant: testTenant,
		Realm:  testRealm,
		Status: "suspended", // Change user status
		UserAttributes: []model.UserAttribute{
			// Update existing email attribute
			{
				ID:     createdUser.UserAttributes[0].ID,
				UserID: testUserID,
				Tenant: testTenant,
				Realm:  testRealm,
				Index:  stringPtr("primary@example.com"),
				Type:   "email",
				Value: model.EmailAttributeValue{
					Email:      "primary@example.com",
					Verified:   true, // Change verification status
					VerifiedAt: timePtr(time.Now()),
				},
				CreatedAt: createdUser.UserAttributes[0].CreatedAt,
				// Remove UpdatedAt - let the database function set it
			},
			// Keep phone attribute unchanged
			{
				ID:     createdUser.UserAttributes[1].ID,
				UserID: testUserID,
				Tenant: testTenant,
				Realm:  testRealm,
				Index:  stringPtr("+1234567890"),
				Type:   "phone",
				Value: model.PhoneAttributeValue{
					Phone:    "+1234567890",
					Verified: false,
				},
				CreatedAt: createdUser.UserAttributes[1].CreatedAt,
				// Remove UpdatedAt - let the database function set it
			},
			// Add new attribute
			{
				UserID: testUserID,
				Tenant: testTenant,
				Realm:  testRealm,
				Index:  stringPtr("work@example.com"),
				Type:   "email",
				Value: model.EmailAttributeValue{
					Email:    "work@example.com",
					Verified: false,
				},
				// Remove CreatedAt and UpdatedAt - let the database function set them
			},
		},
	}

	t.Run("UpdateUserWithAttributes", func(t *testing.T) {
		// Act: Update the user with attributes
		err := db.UpdateUserWithAttributes(ctx, updatedUser)
		require.NoError(t, err, "Failed to update user with attributes")

		// Assert: Retrieve the updated user and verify changes
		updatedUserFromDB, err := db.GetUserWithAttributes(ctx, testTenant, testRealm, testUserID)
		require.NoError(t, err, "Failed to get updated user")
		require.NotNil(t, updatedUserFromDB, "Updated user should not be nil")

		// Check user status was updated
		assert.Equal(t, "suspended", updatedUserFromDB.Status, "User status should be updated")

		// Check that we now have 3 attributes
		assert.Len(t, updatedUserFromDB.UserAttributes, 3, "User should have 3 attributes after update")

		// Find attributes by type for easier testing
		emailAttrs := updatedUserFromDB.GetAttributesByType("email")
		phoneAttrs := updatedUserFromDB.GetAttributesByType("phone")

		assert.Len(t, emailAttrs, 2, "Should have 2 email attributes")
		assert.Len(t, phoneAttrs, 1, "Should have 1 phone attribute")

		// Check primary email attribute was updated
		var primaryEmailAttr *model.UserAttribute
		for _, attr := range emailAttrs {
			if attr.Index != nil && *attr.Index == "primary@example.com" {
				primaryEmailAttr = &attr
				break
			}
		}
		require.NotNil(t, primaryEmailAttr, "Primary email attribute should exist")

		// Verify email attribute was updated
		if emailValue, ok := primaryEmailAttr.Value.(map[string]interface{}); ok {
			assert.Equal(t, true, emailValue["verified"], "Email verification status should be updated")
			assert.NotNil(t, emailValue["verified_at"], "Email verification timestamp should be set")
		} else {
			t.Errorf("Failed to type assert primary email Value to map[string]interface{}")
		}

		// Check that email attribute updated timestamp changed
		assert.NotEqual(t, originalEmailUpdatedAt, primaryEmailAttr.UpdatedAt,
			"Primary email attribute should have updated timestamp")

		// Check phone attribute remained unchanged
		var phoneAttr *model.UserAttribute
		for _, attr := range phoneAttrs {
			if attr.Index != nil && *attr.Index == "+1234567890" {
				phoneAttr = &attr
				break
			}
		}
		require.NotNil(t, phoneAttr, "Phone attribute should exist")

		// Verify phone attribute was not changed
		if phoneValue, ok := phoneAttr.Value.(map[string]interface{}); ok {
			assert.Equal(t, false, phoneValue["verified"], "Phone verification status should remain unchanged")
		} else {
			t.Errorf("Failed to type assert phone Value to map[string]interface{}")
		}

		// Check that phone attribute timestamp did not change
		assert.Equal(t, originalPhoneUpdatedAt, phoneAttr.UpdatedAt,
			"Phone attribute timestamp should remain unchanged")

		// Check new work email attribute was added
		var workEmailAttr *model.UserAttribute
		for _, attr := range emailAttrs {
			if attr.Index != nil && *attr.Index == "work@example.com" {
				workEmailAttr = &attr
				break
			}
		}
		require.NotNil(t, workEmailAttr, "Work email attribute should exist")

		// Verify new work email attribute
		if emailValue, ok := workEmailAttr.Value.(map[string]interface{}); ok {
			assert.Equal(t, "work@example.com", emailValue["email"], "Work email should be correct")
			assert.Equal(t, false, emailValue["verified"], "Work email should be unverified")
		} else {
			t.Errorf("Failed to type assert work email Value to map[string]interface{}")
		}

		// Verify new attribute has recent timestamps (should be different from original)
		assert.NotEqual(t, originalEmailUpdatedAt, workEmailAttr.CreatedAt,
			"New work email should have different created timestamp")
		assert.NotEqual(t, originalEmailUpdatedAt, workEmailAttr.UpdatedAt,
			"New work email should have different updated timestamp")
	})

	t.Run("TestUpdateNonExistentAttribute", func(t *testing.T) {
		// Test updating a non-existent attribute should return an error
		nonExistentAttribute := model.UserAttribute{
			ID:        "non-existent-id",
			UserID:    "non-existent-user",
			Tenant:    testTenant,
			Realm:     testRealm,
			Type:      "email",
			Value:     map[string]interface{}{"email": "test@example.com"},
			UpdatedAt: time.Now(),
		}

		err := db.UpdateUserAttribute(ctx, &nonExistentAttribute)
		assert.Error(t, err, "UpdateUserAttribute should return an error for non-existent attribute")
		assert.Contains(t, err.Error(), "expected 1 row to be affected", "Error should mention rows affected")
	})

	// Clean up
	attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
	require.NoError(t, err)
	for _, attr := range attrs {
		db.DeleteUserAttribute(ctx, testTenant, testRealm, attr.ID)
	}
}

// TemplateTestIndexUniqueConstraint tests the index unique constraint
func TemplateTestIndexUniqueConstraint(t *testing.T, db UserAttributeDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	testUserID := "123e4567-e89b-12d3-a456-426614174002" // Use a different UUID for this test

	// Create a user with a unique email attribute
	uniqueEmailAttr := model.UserAttribute{
		UserID: testUserID,
		Tenant: testTenant,
		Realm:  testRealm,
		Index:  stringPtr("unique@example.com"),
		Type:   "email",
		Value: model.EmailAttributeValue{
			Email:    "unique@example.com",
			Verified: true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	passwordAttr := model.UserAttribute{
		UserID: testUserID,
		Tenant: testTenant,
		Realm:  testRealm,
		Index:  nil, // Password attributes don't need an index
		Type:   "password",
		Value: model.PasswordAttributeValue{
			PasswordHash: "supersecret",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := db.CreateUserWithAttributes(ctx, &model.User{
		ID:     testUserID,
		Tenant: testTenant,
		Realm:  testRealm,
		Status: "active",
		UserAttributes: []model.UserAttribute{
			uniqueEmailAttr,
			passwordAttr,
		},
	})
	require.NoError(t, err)

	t.Run("UniqueEmailIndex", func(t *testing.T) {

		// Try to create a user with a duplicate email index
		duplicateEmailAttr := model.UserAttribute{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("unique@example.com"),
			Type:   "email",
			Value: model.EmailAttributeValue{
				Email:    "duplicate@example.com",
				Verified: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = db.CreateUserWithAttributes(ctx, &model.User{
			ID:     testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Status: "active",
			UserAttributes: []model.UserAttribute{
				duplicateEmailAttr,
			},
		})
		require.Error(t, err, "Failed to create user with attributes for unique constraint test")

	})

	t.Run("NullIndex Dublication Allowed", func(t *testing.T) {

		passwordAttr := model.UserAttribute{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  nil, // Password attributes don't need an index
			Type:   "password",
			Value: model.PasswordAttributeValue{
				PasswordHash: "supersecret",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		anotherEmailAttr := model.UserAttribute{
			UserID: testUserID,
			Tenant: testTenant,
			Realm:  testRealm,
			Index:  stringPtr("another@example.com"),
			Type:   "email",
			Value: model.EmailAttributeValue{
				Email:    "another@example.com",
				Verified: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Try to create a user with the same password
		err := db.CreateUserWithAttributes(ctx, &model.User{
			ID:     "123e4567-e89b-12d3-a456-426614174004", // Use a different UUID
			Tenant: testTenant,
			Realm:  testRealm,
			Status: "active",
			UserAttributes: []model.UserAttribute{
				passwordAttr,
				anotherEmailAttr,
			},
		})
		assert.NoError(t, err)
	})

	// Clean up
	attrs, err := db.ListUserAttributes(ctx, testTenant, testRealm, testUserID)
	require.NoError(t, err)
	for _, attr := range attrs {
		db.DeleteUserAttribute(ctx, testTenant, testRealm, attr.ID)
	}
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}
