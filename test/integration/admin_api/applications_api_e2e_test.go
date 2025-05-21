package integration_admin_api

import (
	"goiam/test/integration"
	"net/http"
	"testing"
)

// TestApplicationAPI_E2E performs a complete end-to-end test of the admin API application management functionality.
// It tests the following application operations in sequence:
// 1. Creating a new application
// 2. Listing all applications
// 3. Getting a specific application's details
// 4. Updating an application's information
// 5. Regenerating client secret
// 6. Deleting an application
// 7. Verifying application deletion
// The test uses a test tenant "acme" and realm "test_realm" for all operations.

func TestApplicationAPI_E2E(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	// Test application data
	tenant := "acme"
	realm := "test_realm"
	clientId := "test_app"
	minimalApp := map[string]interface{}{
		"client_id":   clientId,
		"realm":       realm,
		"tenant":      tenant,
		"description": "Test Application",
	}

	// Test creating an application
	t.Run("Create Application", func(t *testing.T) {
		e.POST("/admin/acme/test_realm/applications/test_app").
			WithJSON(minimalApp).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", realm).
			HasValue("client_id", clientId).
			HasValue("description", "Test Application")
	})

	// Test listing applications
	t.Run("List Applications", func(t *testing.T) {
		e.GET("/admin/acme/test_realm/applications").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			Length().
			Equal(1)
	})

	// Test getting a specific application
	t.Run("Get Application", func(t *testing.T) {
		e.GET("/admin/acme/test_realm/applications/test_app").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", realm).
			HasValue("client_id", clientId).
			HasValue("description", "Test Application")
	})

	// Test updating an application
	updatedName := "Updated Test Application"
	t.Run("Update Application", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"description": updatedName,
		}

		e.PUT("/admin/acme/test_realm/applications/test_app").
			WithJSON(updatePayload).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("description", updatedName)

		// Verify the update
		e.GET("/admin/acme/test_realm/applications/test_app").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("tenant", tenant).
			HasValue("realm", realm).
			HasValue("client_id", clientId).
			HasValue("description", updatedName)
	})

	// Test regenerating client secret
	t.Run("Regenerate Client Secret", func(t *testing.T) {
		response := e.POST("/admin/acme/test_realm/applications/test_app/regenerate-secret").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		// Verify we got a client secret
		response.Value("client_secret").String().NotEmpty()
	})

	// Test deleting an application
	t.Run("Delete Application", func(t *testing.T) {
		e.DELETE("/admin/acme/test_realm/applications/test_app").
			Expect().
			Status(http.StatusNoContent)

		// Verify application is deleted
		e.GET("/admin/acme/test_realm/applications/test_app").
			Expect().
			Status(http.StatusNotFound)
	})
}
