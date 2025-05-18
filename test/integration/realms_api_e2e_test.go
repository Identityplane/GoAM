package integration

import (
	"net/http"
	"testing"
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
	e := SetupIntegrationTest(t, "")

	// Test realm data
	tenant := "acme"
	testRealm := map[string]interface{}{
		"realm":      "test_realm",
		"realm_name": "Test Realm",
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
			HasValue("realm_name", testRealm["realm_name"])
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
			HasValue("realm_name", testRealm["realm_name"])
	})

	// Test updating a realm
	t.Run("Update Realm", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"realm_name": "Updated Test Realm",
		}

		e.PATCH("/admin/acme/test_realm/").
			WithJSON(updatePayload).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("realm_name", updatePayload["realm_name"])

		// Verify the update
		e.GET("/admin/acme/test_realm/").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", testRealm["realm"]).
			HasValue("realm_name", updatePayload["realm_name"])
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
