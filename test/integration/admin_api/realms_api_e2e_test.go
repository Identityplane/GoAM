package integration_admin_api

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
)

// TestRealmAPI_E2E performs a complete end-to-end test of the admin API realm management functionality.
// It tests the following realm operations in sequence:
// 1. Creating a new realm
// 2. Listing all realms
// 3. Getting a specific realm's details
// 4. Updating a realm's information
// 5. Deleting a realm
// 6. Verifying realm deletion
// The test uses a test tenant "acme" and realm "test_realm" for all operations.

func TestRealmAPI_E2E(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	// Test realm data
	tenant := "acme"
	testRealm := map[string]interface{}{
		"realm":          "test_realm",
		"realm_name":     "Test Realm",
		"base_url":       "https://test.example.com",
		"realm_settings": map[string]string{"theme": "dark", "language": "en"},
	}

	// Test creating a realm
	t.Run("Create Realm", func(t *testing.T) {
		e.POST("/admin/acme/test_realm/").
			WithJSON(testRealm).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", testRealm["realm"]).
			HasValue("realm_name", testRealm["realm_name"]).
			HasValue("base_url", testRealm["base_url"]).
			HasValue("realm_settings", testRealm["realm_settings"])
	})

	// Test listing realms
	t.Run("List Realms", func(t *testing.T) {
		resp := e.GET("/admin/realms").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		// Verify the response contains our test realm
		resp.Length().IsEqual(2)
	})

	// Test getting a specific realm
	t.Run("Get Realm", func(t *testing.T) {
		e.GET("/admin/acme/test_realm/").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", testRealm["realm"]).
			HasValue("realm_name", testRealm["realm_name"]).
			HasValue("base_url", testRealm["base_url"]).
			HasValue("realm_settings", testRealm["realm_settings"])
	})

	// Test updating a realm
	t.Run("Update Realm", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"realm_name": "Updated Test Realm",
			"base_url":   "https://updated.example.com",
		}

		e.PATCH("/admin/acme/test_realm/").
			WithJSON(updatePayload).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("realm_name", updatePayload["realm_name"]).
			HasValue("base_url", updatePayload["base_url"])

		// Verify the update
		e.GET("/admin/acme/test_realm/").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", testRealm["realm"]).
			HasValue("realm_name", updatePayload["realm_name"]).
			HasValue("base_url", updatePayload["base_url"]).
			HasValue("realm_settings", testRealm["realm_settings"]) // Settings should remain unchanged
	})

	// Test updating only the realm settings
	t.Run("Update Realm Settings Only", func(t *testing.T) {
		updateSettingsPayload := map[string]interface{}{
			"realm_settings": map[string]string{
				"theme":     "light",
				"language":  "fr",
				"timezone":  "UTC",
				"new_field": "new_value",
			},
		}

		e.PATCH("/admin/acme/test_realm/").
			WithJSON(updateSettingsPayload).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("realm_settings", updateSettingsPayload["realm_settings"])

		// Verify the settings update
		e.GET("/admin/acme/test_realm/").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("realm_name", "Updated Test Realm").        // Should remain unchanged
			HasValue("base_url", "https://updated.example.com"). // Should remain unchanged
			HasValue("realm_settings", updateSettingsPayload["realm_settings"])
	})

	// Test creating a realm with empty settings
	t.Run("Create Realm With Empty Settings", func(t *testing.T) {
		emptySettingsRealm := map[string]interface{}{
			"realm":          "test_realm_empty_settings",
			"realm_name":     "Test Realm Empty Settings",
			"base_url":       "https://empty.example.com",
			"realm_settings": map[string]string{},
		}

		e.POST("/admin/acme/test_realm_empty_settings/").
			WithJSON(emptySettingsRealm).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", emptySettingsRealm["realm"]).
			HasValue("realm_name", emptySettingsRealm["realm_name"]).
			HasValue("base_url", emptySettingsRealm["base_url"]).
			HasValue("realm_settings", emptySettingsRealm["realm_settings"])

		// Verify the realm was created with empty settings
		e.GET("/admin/acme/test_realm_empty_settings/").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("realm_settings", map[string]string{})

		// Clean up the test realm
		e.DELETE("/admin/acme/test_realm_empty_settings/").
			Expect().
			Status(http.StatusNoContent)
	})

	// Test deleting a realm
	t.Run("Delete Realm", func(t *testing.T) {
		e.DELETE("/admin/acme/test_realm/").
			Expect().
			Status(http.StatusNoContent)

		// Verify realm is deleted
		e.GET("/admin/acme/test_realm/").
			Expect().
			Status(http.StatusNotFound)
	})
}
