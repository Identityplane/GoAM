package integration_admin_api

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
)

// This test performs a complete end-to-end test of the admin API user management functionality.
// It tests the following user operations in sequence:
// 1. Creating a new user with all required fields
// 2. Retrieving user statistics
// 3. Listing users with pagination
// 4. Getting a specific user's details
// 5. Updating a user's information
// 6. Deleting a user and verifying deletion
// The test uses a test tenant "acme" and realm "customers" for all operations.

func TestUserAPI_E2E(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	// Test user data with UserFlat model fields
	testUser := map[string]interface{}{
		"id":                 "admin_test_user",
		"status":             "active",
		"tenant":             "acme",
		"realm":              "customers",
		"email":              "test@example.com",
		"email_verified":     true,
		"phone":              "+1234567890",
		"phone_verified":     false,
		"preferred_username": "testuser",
		"given_name":         "John",
		"family_name":        "Doe",
		"name":               "John Doe",
		"locale":             "en-US",
	}

	// Test creating a user with UserFlat model
	t.Run("Create User", func(t *testing.T) {
		e.POST("/admin/acme/customers/users/"+testUser["id"].(string)).
			WithJSON(testUser).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object().
			HasValue("id", testUser["id"]).
			HasValue("status", testUser["status"]).
			HasValue("tenant", testUser["tenant"]).
			HasValue("realm", testUser["realm"]).
			HasValue("email", testUser["email"]).
			HasValue("email_verified", testUser["email_verified"]).
			HasValue("phone", testUser["phone"]).
			HasValue("phone_verified", testUser["phone_verified"]).
			HasValue("preferred_username", testUser["preferred_username"]).
			HasValue("given_name", testUser["given_name"]).
			HasValue("family_name", testUser["family_name"]).
			HasValue("name", testUser["name"]).
			HasValue("locale", testUser["locale"])
	})

	// Test getting user stats
	t.Run("Get User Stats", func(t *testing.T) {
		e.GET("/admin/acme/customers/users/stats").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("total_users", 1).
			HasValue("active_users", 1)
	})

	// Test listing users with pagination
	t.Run("List Users", func(t *testing.T) {
		resp := e.GET("/admin/acme/customers/users").
			WithQuery("page", 1).
			WithQuery("page_size", 20).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		// Check pagination metadata
		resp.Value("pagination").
			Object().
			HasValue("page", 1).
			HasValue("page_size", 20).
			HasValue("total_items", 1).
			HasValue("total_pages", 1)

		// Check user data
		resp.Value("data").
			Array().
			Length().
			Equal(1)
	})

	// Test getting a specific user and verify UserFlat fields are returned
	t.Run("Get User", func(t *testing.T) {
		e.GET("/admin/acme/customers/users/"+testUser["id"].(string)).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("id", testUser["id"]).
			HasValue("status", testUser["status"]).
			HasValue("tenant", testUser["tenant"]).
			HasValue("realm", testUser["realm"]).
			HasValue("email", testUser["email"]).
			HasValue("email_verified", testUser["email_verified"]).
			HasValue("phone", testUser["phone"]).
			HasValue("phone_verified", testUser["phone_verified"]).
			HasValue("preferred_username", testUser["preferred_username"]).
			HasValue("given_name", testUser["given_name"]).
			HasValue("family_name", testUser["family_name"]).
			HasValue("name", testUser["name"]).
			HasValue("locale", testUser["locale"]).
			ContainsKey("url")
	})

	// Test updating a user with PUT (full update)
	t.Run("Update User", func(t *testing.T) {
		updatedUser := map[string]interface{}{
			"id":                 testUser["id"],
			"status":             "inactive",
			"tenant":             testUser["tenant"],
			"realm":              testUser["realm"],
			"email":              "updated@example.com",
			"email_verified":     false,
			"phone":              "+9876543210",
			"phone_verified":     true,
			"preferred_username": "updateduser",
			"given_name":         "Jane",
			"family_name":        "Smith",
			"name":               "Jane Smith",
			"locale":             "de-DE",
		}

		e.PUT("/admin/acme/customers/users/"+testUser["id"].(string)).
			WithJSON(updatedUser).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("id", updatedUser["id"]).
			HasValue("status", updatedUser["status"]).
			HasValue("email", updatedUser["email"]).
			HasValue("email_verified", updatedUser["email_verified"]).
			HasValue("phone", updatedUser["phone"]).
			HasValue("phone_verified", updatedUser["phone_verified"]).
			HasValue("preferred_username", updatedUser["preferred_username"]).
			HasValue("given_name", updatedUser["given_name"]).
			HasValue("family_name", updatedUser["family_name"]).
			HasValue("name", updatedUser["name"]).
			HasValue("locale", updatedUser["locale"]).
			ContainsKey("url")
	})

	// Test patching a user (partial update)
	t.Run("Patch User", func(t *testing.T) {
		patchData := map[string]interface{}{
			"preferred_username": "patcheduser",
			"given_name":         "Patched",
		}

		e.PATCH("/admin/acme/customers/users/"+testUser["id"].(string)).
			WithJSON(patchData).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("preferred_username", patchData["preferred_username"]).
			HasValue("given_name", patchData["given_name"]).
			// Verify other fields from previous update are still there
			HasValue("email", "updated@example.com").
			HasValue("family_name", "Smith").
			ContainsKey("url")
	})

	// Test deleting a user
	t.Run("Delete User", func(t *testing.T) {
		e.DELETE("/admin/acme/customers/users/" + testUser["id"].(string)).
			Expect().
			Status(http.StatusNoContent)

		// Verify user is deleted
		e.GET("/admin/acme/customers/users/" + testUser["id"].(string)).
			Expect().
			Status(http.StatusNotFound)
	})
}
