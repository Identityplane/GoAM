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
		Index:  "test@example.com",
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
		assert.Equal(t, "test@example.com", attrs[0].Index)

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
			Index:  "primary@example.com",
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
			Index:  "work@example.com",
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
			indexes[attr.Index] = true
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
			Index:  "primary@example.com",
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
			Index:  "work@example.com",
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
			Index:  "+1234567890",
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
		assert.Equal(t, "+1234567890", phoneAttrs[0].Index)

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
			Index:  "primary@example.com",
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
			Index:  "work@example.com",
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
			Index:  "+1234567890",
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
				Index:  "primary@example.com",
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
				Index:  "work@example.com",
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
				Index:  "+1234567890",
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
		assert.Equal(t, "+1234567890", phoneAttrs[0].Index)

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

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}
