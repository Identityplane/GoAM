package integration_admin_api

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
)

// This test performs a complete end-to-end test of the admin API user attributes management functionality.
// It tests the following user attribute operations in sequence:
// 1. Creating a test user to attach attributes to
// 2. Creating a new email attribute for the user
// 3. Creating a second email attribute (testing multiple attributes of same type)
// 4. Listing all attributes for the user
// 5. Getting a specific attribute's details
// 6. Updating an attribute's information
// 7. Deleting specific attributes
// 8. Cleaning up the test user
// The test uses a test tenant "acme" and realm "customers" for all operations.

func TestUserAttributesAPI_E2E(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	// First create a test user to attach attributes to
	testUser := map[string]interface{}{
		"username":     "attr_test_user",
		"display_name": "Attribute Test User",
		"given_name":   "Attribute",
		"family_name":  "Test",
		"email":        "attr_test@example.com",
		"status":       "active",
	}

	var testUserID string

	// Create the test user
	t.Run("Create Test User", func(t *testing.T) {
		resp := e.POST("/admin/acme/customers/users/" + testUser["username"].(string)).
			WithJSON(testUser).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		resp.HasValue("username", testUser["username"])
		testUserID = resp.Value("id").String().Raw()
	})

	var primaryEmailAttributeID string
	var workEmailAttributeID string

	// Test creating a primary email attribute
	t.Run("Create Primary Email Attribute", func(t *testing.T) {
		emailAttribute := map[string]interface{}{
			"type":  "email",
			"index": "primary@example.com",
			"value": map[string]interface{}{
				"email":    "primary@example.com",
				"verified": true,
			},
		}

		resp := e.POST("/admin/acme/customers/users/" + testUserID + "/attributes").
			WithJSON(emailAttribute).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		resp.HasValue("type", "email").
			HasValue("index", "primary@example.com").
			Value("value").Object().HasValue("email", "primary@example.com").
			HasValue("verified", true)

		primaryEmailAttributeID = resp.Value("id").String().Raw()
	})

	// Test creating a work email attribute (multiple attributes of same type)
	t.Run("Create Work Email Attribute", func(t *testing.T) {
		emailAttribute := map[string]interface{}{
			"type":  "email",
			"index": "work@company.com",
			"value": map[string]interface{}{
				"email":    "work@company.com",
				"verified": false,
			},
		}

		resp := e.POST("/admin/acme/customers/users/" + testUserID + "/attributes").
			WithJSON(emailAttribute).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		resp.HasValue("type", "email").
			HasValue("index", "work@company.com").
			Value("value").Object().HasValue("email", "work@company.com").
			HasValue("verified", false)

		workEmailAttributeID = resp.Value("id").String().Raw()
	})

	// Test listing all user attributes
	t.Run("List User Attributes", func(t *testing.T) {
		resp := e.GET("/admin/acme/customers/users/" + testUserID + "/attributes").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().Equal(2)

		// Check that both attributes are present
		attr1 := resp.Element(0).Object()
		attr2 := resp.Element(1).Object()

		// Verify we have both email attributes (order may vary)
		indexes := []string{
			attr1.Value("index").String().Raw(),
			attr2.Value("index").String().Raw(),
		}

		// Check that we have both expected indexes
		hasPrimary := false
		hasWork := false
		for _, index := range indexes {
			if index == "primary@example.com" {
				hasPrimary = true
			} else if index == "work@company.com" {
				hasWork = true
			}
		}

		if !hasPrimary || !hasWork {
			t.Errorf("Expected to find both primary@example.com and work@company.com attributes")
		}
	})

	// Test getting a specific attribute
	t.Run("Get Specific Attribute", func(t *testing.T) {
		e.GET("/admin/acme/customers/users/"+testUserID+"/attributes/email/"+primaryEmailAttributeID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("id", primaryEmailAttributeID).
			HasValue("type", "email").
			HasValue("index", "primary@example.com").
			Value("value").Object().HasValue("email", "primary@example.com").
			HasValue("verified", true)
	})

	// Test updating an attribute (verify the work email)
	t.Run("Update Work Email Attribute", func(t *testing.T) {
		updatedAttribute := map[string]interface{}{
			"type":  "email",
			"index": "work@company.com",
			"value": map[string]interface{}{
				"email":    "work@company.com",
				"verified": true, // Changed from false to true
			},
		}

		e.PATCH("/admin/acme/customers/users/"+testUserID+"/attributes/email/"+workEmailAttributeID).
			WithJSON(updatedAttribute).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("id", workEmailAttributeID).
			HasValue("type", "email").
			HasValue("index", "work@company.com").
			Value("value").Object().HasValue("email", "work@company.com").
			HasValue("verified", true) // Should now be verified
	})

	// Test deleting a specific attribute
	t.Run("Delete Work Email Attribute", func(t *testing.T) {
		e.DELETE("/admin/acme/customers/users/" + testUserID + "/attributes/email/" + workEmailAttributeID).
			Expect().
			Status(http.StatusNoContent)

		// Verify the attribute is deleted
		e.GET("/admin/acme/customers/users/" + testUserID + "/attributes/email/" + workEmailAttributeID).
			Expect().
			Status(http.StatusNotFound)
	})

	// Test that listing attributes now shows only one
	t.Run("List Attributes After Deletion", func(t *testing.T) {
		resp := e.GET("/admin/acme/customers/users/" + testUserID + "/attributes").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().Equal(1)

		// Verify it's the primary email attribute
		resp.Element(0).Object().
			HasValue("index", "primary@example.com").
			HasValue("type", "email")
	})

	// Clean up: delete the remaining attribute
	t.Run("Delete Remaining Attribute", func(t *testing.T) {
		e.DELETE("/admin/acme/customers/users/" + testUserID + "/attributes/email/" + primaryEmailAttributeID).
			Expect().
			Status(http.StatusNoContent)
	})

	// Verify no attributes remain
	t.Run("Verify No Attributes Remain", func(t *testing.T) {
		e.GET("/admin/acme/customers/users/" + testUserID + "/attributes").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			Length().Equal(0)
	})

	// Clean up: delete the test user
	t.Run("Delete Test User", func(t *testing.T) {
		e.DELETE("/admin/acme/customers/users/" + testUser["username"].(string)).
			Expect().
			Status(http.StatusNoContent)

		// Verify user is deleted
		e.GET("/admin/acme/customers/users/" + testUser["username"].(string)).
			Expect().
			Status(http.StatusNotFound)
	})
}
