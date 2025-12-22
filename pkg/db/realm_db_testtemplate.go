package db

import (
	"context"
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TemplateTestRealmCRUD is a parameterized test for basic CRUD operations on realms
func TemplateTestRealmCRUD(t *testing.T, db RealmDB) {
	ctx := context.Background()
	testTenant := "test-tenant"

	// Create test realm
	testRealm := model.Realm{
		Realm:         "test-realm",
		RealmName:     "Test Realm",
		Tenant:        testTenant,
		BaseUrl:       "https://test.example.com",
		RealmSettings: map[string]string{"theme": "dark", "language": "en"},
	}

	t.Run("CreateRealm", func(t *testing.T) {
		err := db.CreateRealm(ctx, testRealm)
		assert.NoError(t, err)
	})

	t.Run("GetRealm", func(t *testing.T) {
		realm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		assert.NoError(t, err)
		assert.NotNil(t, realm)
		assert.Equal(t, testRealm.Realm, realm.Realm)
		assert.Equal(t, testRealm.RealmName, realm.RealmName)
		assert.Equal(t, testRealm.BaseUrl, realm.BaseUrl)
		assert.Equal(t, testRealm.RealmSettings, realm.RealmSettings)
		assert.Equal(t, "dark", realm.RealmSettings["theme"])
		assert.Equal(t, "en", realm.RealmSettings["language"])
	})

	t.Run("UpdateRealm", func(t *testing.T) {
		realm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		require.NoError(t, err)
		require.NotNil(t, realm)

		realm.RealmName = "Updated Test Realm"
		realm.BaseUrl = "https://updated.example.com"
		realm.RealmSettings["theme"] = "light"
		realm.RealmSettings["new_setting"] = "value"
		err = db.UpdateRealm(ctx, realm)
		assert.NoError(t, err)

		updatedRealm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Test Realm", updatedRealm.RealmName)
		assert.Equal(t, "https://updated.example.com", updatedRealm.BaseUrl)
		assert.Equal(t, "light", updatedRealm.RealmSettings["theme"])
		assert.Equal(t, "en", updatedRealm.RealmSettings["language"])
		assert.Equal(t, "value", updatedRealm.RealmSettings["new_setting"])
	})

	// Test updating only the RealmSettings field
	t.Run("UpdateRealmSettingsOnly", func(t *testing.T) {
		realm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		require.NoError(t, err)
		require.NotNil(t, realm)

		// Update only the realm settings
		realm.RealmSettings = map[string]string{
			"theme":     "blue",
			"language":  "fr",
			"timezone":  "UTC",
			"new_field": "new_value",
		}
		err = db.UpdateRealm(ctx, realm)
		assert.NoError(t, err)

		// Verify the update
		updatedRealm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Test Realm", updatedRealm.RealmName)        // Should remain unchanged
		assert.Equal(t, "https://updated.example.com", updatedRealm.BaseUrl) // Should remain unchanged
		assert.Equal(t, "blue", updatedRealm.RealmSettings["theme"])
		assert.Equal(t, "fr", updatedRealm.RealmSettings["language"])
		assert.Equal(t, "UTC", updatedRealm.RealmSettings["timezone"])
		assert.Equal(t, "new_value", updatedRealm.RealmSettings["new_field"])
		// Old values should be gone
		assert.Equal(t, "", updatedRealm.RealmSettings["en"])
		assert.Equal(t, "", updatedRealm.RealmSettings["value"])
	})

	// Test creating a realm with empty RealmSettings
	t.Run("CreateRealmWithEmptySettings", func(t *testing.T) {
		emptySettingsRealm := model.Realm{
			Realm:         "test-realm-empty-settings",
			RealmName:     "Test Realm Empty Settings",
			Tenant:        testTenant,
			BaseUrl:       "https://empty.example.com",
			RealmSettings: map[string]string{},
		}

		err := db.CreateRealm(ctx, emptySettingsRealm)
		assert.NoError(t, err)

		// Verify the realm was created with empty settings
		createdRealm, err := db.GetRealm(ctx, testTenant, emptySettingsRealm.Realm)
		assert.NoError(t, err)
		assert.NotNil(t, createdRealm)
		assert.Equal(t, emptySettingsRealm.Realm, createdRealm.Realm)
		assert.Equal(t, emptySettingsRealm.RealmName, createdRealm.RealmName)
		assert.Equal(t, emptySettingsRealm.BaseUrl, createdRealm.BaseUrl)
		assert.NotNil(t, createdRealm.RealmSettings)
		assert.Len(t, createdRealm.RealmSettings, 0)

		// Clean up
		err = db.DeleteRealm(ctx, testTenant, emptySettingsRealm.Realm)
		assert.NoError(t, err)
	})

	t.Run("ListRealms", func(t *testing.T) {
		realms, err := db.ListRealms(ctx, testTenant)
		assert.NoError(t, err)
		assert.Len(t, realms, 1)
		assert.Equal(t, testRealm.Realm, realms[0].Realm)
		assert.Equal(t, "https://updated.example.com", realms[0].BaseUrl)
		assert.Equal(t, "blue", realms[0].RealmSettings["theme"])
		assert.Equal(t, "fr", realms[0].RealmSettings["language"])
		assert.Equal(t, "UTC", realms[0].RealmSettings["timezone"])
		assert.Equal(t, "new_value", realms[0].RealmSettings["new_field"])
	})

	t.Run("DeleteRealm", func(t *testing.T) {
		// First check if realm is empty
		err := db.DeleteRealm(ctx, testTenant, testRealm.Realm)
		assert.NoError(t, err)

		realm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		assert.NoError(t, err)
		assert.Nil(t, realm)
	})
}
