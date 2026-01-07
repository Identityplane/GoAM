package integration_admin_api

import (
	"context"
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/test/integration"
)

// This test performs a complete end-to-end test of the admin API user attributes management functionality.
// It tests the following user attribute operations:
// 1. Creating attributes of different types (email, phone, oidc, username)
// 2. Listing all attributes and filtering by type
// 3. Getting specific attributes
// 4. Updating attributes (including type immutability)
// 5. Deleting attributes
// 6. Error cases (404, 400, invalid JSON, etc.)
// 7. Edge cases (null index, sensitive indices, etc.)
// The test uses a test tenant "acme" and realm "customers" for all operations.

func TestUserAttributesAPI_E2E(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")
	ctx := context.Background()

	// Arrange
	tenant := "acme"
	realm := "customers"
	testUserID := "testuser"

	// Create a new realm
	service.GetServices().RealmService.CreateRealm(&model.Realm{
		Tenant: tenant,
		Realm:  realm,
	})

	// Create an empty user
	service.GetServices().UserService.CreateUser(ctx, tenant, realm, model.User{
		ID: testUserID,
	})

	var primaryEmailAttributeID string
	var workEmailAttributeID string
	var phoneAttributeID string
	var oidcAttributeID string
	var usernameAttributeID string

	// Test creating a primary email attribute
	t.Run("Create Primary Email Attribute", func(t *testing.T) {
		emailAttribute := map[string]interface{}{
			"type": model.AttributeTypeEmail,
			"value": map[string]interface{}{
				"email":    "primary@example.com",
				"verified": true,
			},
		}

		resp := e.POST("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			WithJSON(emailAttribute).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		resp.HasValue("type", model.AttributeTypeEmail).
			HasValue("index", "primary@example.com").
			Value("value").Object().HasValue("email", "primary@example.com").
			HasValue("verified", true)

		primaryEmailAttributeID = resp.Value("id").String().Raw()
	})

	// Test creating a work email attribute (multiple attributes of same type)
	t.Run("Create Work Email Attribute", func(t *testing.T) {
		emailAttribute := map[string]interface{}{
			"type": model.AttributeTypeEmail,
			"value": map[string]interface{}{
				"email":    "work@company.com",
				"verified": false,
			},
		}

		resp := e.POST("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			WithJSON(emailAttribute).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		resp.HasValue("type", model.AttributeTypeEmail).
			HasValue("index", "work@company.com").
			Value("value").Object().HasValue("email", "work@company.com").
			HasValue("verified", false)

		workEmailAttributeID = resp.Value("id").String().Raw()
	})

	// Test creating a phone attribute
	t.Run("Create Phone Attribute", func(t *testing.T) {
		phoneAttribute := map[string]interface{}{
			"type": model.AttributeTypePhone,
			"value": map[string]interface{}{
				"phone":    "+1234567890",
				"verified": true,
			},
		}

		resp := e.POST("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			WithJSON(phoneAttribute).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		resp.HasValue("type", model.AttributeTypePhone).
			HasValue("index", "+1234567890").
			Value("value").Object().HasValue("phone", "+1234567890").
			HasValue("verified", true)

		phoneAttributeID = resp.Value("id").String().Raw()
	})

	// Test creating an OIDC attribute
	t.Run("Create OIDC Attribute", func(t *testing.T) {
		oidcAttribute := map[string]interface{}{
			"type": model.AttributeTypeOidc,
			"value": map[string]interface{}{
				"issuer":       "https://accounts.google.com",
				"client_id":    "1234567890",
				"sub":          "google-user-123",
				"access_token": "token123",
			},
		}

		resp := e.POST("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			WithJSON(oidcAttribute).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		resp.HasValue("type", model.AttributeTypeOidc).
			HasValue("index", "https://accounts.google.com/google-user-123").
			Value("value").Object().
			HasValue("issuer", "https://accounts.google.com").
			HasValue("sub", "google-user-123")

		oidcAttributeID = resp.Value("id").String().Raw()
	})

	// Test creating a username attribute
	t.Run("Create Username Attribute", func(t *testing.T) {
		usernameAttribute := map[string]interface{}{
			"type": model.AttributeTypeUsername,
			"value": map[string]interface{}{
				"preferred_username": "johndoe",
				"given_name":         "John",
				"family_name":        "Doe",
			},
		}

		resp := e.POST("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			WithJSON(usernameAttribute).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		resp.HasValue("type", model.AttributeTypeUsername).
			Value("value").Object().
			HasValue("preferred_username", "johndoe").
			HasValue("given_name", "John").
			HasValue("family_name", "Doe")

		usernameAttributeID = resp.Value("id").String().Raw()
	})

	// Test listing all user attributes
	t.Run("List All User Attributes", func(t *testing.T) {
		resp := e.GET("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().Equal(5) // 2 emails, 1 phone, 1 oidc, 1 username
	})

	// Test filtering attributes by type
	t.Run("List Attributes Filtered By Type - Email", func(t *testing.T) {
		resp := e.GET("/admin/"+tenant+"/"+realm+"/users/"+testUserID+"/attributes").
			WithQuery("type", model.AttributeTypeEmail).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().Equal(2)

		// Verify all returned attributes are of type email
		for i := 0; i < 2; i++ {
			resp.Element(i).Object().HasValue("type", model.AttributeTypeEmail)
		}
	})

	t.Run("List Attributes Filtered By Type - Phone", func(t *testing.T) {
		resp := e.GET("/admin/"+tenant+"/"+realm+"/users/"+testUserID+"/attributes").
			WithQuery("type", model.AttributeTypePhone).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().Equal(1)
		resp.Element(0).Object().HasValue("type", model.AttributeTypePhone)
	})

	t.Run("List Attributes Filtered By Type - Non-existent", func(t *testing.T) {
		resp := e.GET("/admin/"+tenant+"/"+realm+"/users/"+testUserID+"/attributes").
			WithQuery("type", "nonexistent").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().Equal(0)
	})

	// Test getting a specific attribute
	t.Run("Get Specific Attribute By ID", func(t *testing.T) {
		e.GET("/admin/"+tenant+"/"+realm+"/users/"+testUserID+"/attributes/"+primaryEmailAttributeID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("id", primaryEmailAttributeID).
			HasValue("type", model.AttributeTypeEmail).
			HasValue("index", "primary@example.com").
			Value("value").Object().HasValue("email", "primary@example.com").
			HasValue("verified", true)
	})

	// Test getting non-existent attribute
	t.Run("Get Non-existent Attribute - 404", func(t *testing.T) {
		e.GET("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/nonexistent-id").
			Expect().
			Status(http.StatusNotFound)
	})

	// Test updating an attribute
	t.Run("Update Email Attribute", func(t *testing.T) {
		updatedAttribute := map[string]interface{}{
			"value": map[string]interface{}{
				"email":    "work@company.com",
				"verified": true, // Changed from false to true
			},
		}

		resp := e.PATCH("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + workEmailAttributeID).
			WithJSON(updatedAttribute).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		resp.HasValue("id", workEmailAttributeID).
			HasValue("type", model.AttributeTypeEmail). // Type should remain unchanged
			HasValue("index", "work@company.com").
			Value("value").Object().HasValue("email", "work@company.com").
			HasValue("verified", true) // Should now be verified
	})

	// Test that type cannot be changed in update
	t.Run("Update Attribute - Type Cannot Be Changed", func(t *testing.T) {
		updatedAttribute := map[string]interface{}{
			"type": model.AttributeTypePhone, // Try to change type
			"value": map[string]interface{}{
				"email":    "work@company.com",
				"verified": true,
			},
		}

		// The type in the request should be ignored, original type should be preserved
		resp := e.PATCH("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + workEmailAttributeID).
			WithJSON(updatedAttribute).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		// Type should still be email, not phone
		resp.HasValue("type", model.AttributeTypeEmail)
	})

	// Test updating with invalid JSON
	t.Run("Update Attribute - Invalid JSON", func(t *testing.T) {
		e.PATCH("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + workEmailAttributeID).
			WithBytes([]byte("invalid json")).
			Expect().
			Status(http.StatusBadRequest)
	})

	// Test creating attribute with invalid JSON
	t.Run("Create Attribute - Invalid JSON", func(t *testing.T) {
		e.POST("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			WithBytes([]byte("invalid json")).
			Expect().
			Status(http.StatusBadRequest)
	})

	// Test creating attribute for non-existent user
	t.Run("Create Attribute - Non-existent User", func(t *testing.T) {
		emailAttribute := map[string]interface{}{
			"type": model.AttributeTypeEmail,
			"value": map[string]interface{}{
				"email": "test@example.com",
			},
		}

		e.POST("/admin/" + tenant + "/" + realm + "/users/nonexistent-user/attributes").
			WithJSON(emailAttribute).
			Expect().
			Status(http.StatusNotFound)
	})

	// Test listing attributes for non-existent user
	t.Run("List Attributes - Non-existent User", func(t *testing.T) {
		e.GET("/admin/" + tenant + "/" + realm + "/users/nonexistent-user/attributes").
			Expect().
			Status(http.StatusNotFound)
	})

	// Test deleting a specific attribute
	t.Run("Delete Phone Attribute", func(t *testing.T) {
		e.DELETE("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + phoneAttributeID).
			Expect().
			Status(http.StatusNoContent)

		// Verify the attribute is deleted
		e.GET("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + phoneAttributeID).
			Expect().
			Status(http.StatusNotFound)
	})

	// Test deleting non-existent attribute
	t.Run("Delete Non-existent Attribute", func(t *testing.T) {
		e.DELETE("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/nonexistent-id").
			Expect().
			Status(http.StatusNoContent) // DELETE is idempotent
	})

	// Test that listing attributes after deletion shows correct count
	t.Run("List Attributes After Deletion", func(t *testing.T) {
		resp := e.GET("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().Equal(4) // 2 emails, 1 oidc, 1 username (phone was deleted)
	})

	// Test filtering after deletion
	t.Run("List Email Attributes After Deletion", func(t *testing.T) {
		resp := e.GET("/admin/"+tenant+"/"+realm+"/users/"+testUserID+"/attributes").
			WithQuery("type", model.AttributeTypeEmail).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().Equal(2)
	})

	// Test empty list when no attributes match filter
	t.Run("List Attributes - Empty Filter Result", func(t *testing.T) {
		resp := e.GET("/admin/"+tenant+"/"+realm+"/users/"+testUserID+"/attributes").
			WithQuery("type", model.AttributeTypePhone).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().Equal(0)
	})

	// Clean up: delete all remaining attributes
	t.Run("Cleanup - Delete All Attributes", func(t *testing.T) {
		e.DELETE("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + primaryEmailAttributeID).
			Expect().
			Status(http.StatusNoContent)

		e.DELETE("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + workEmailAttributeID).
			Expect().
			Status(http.StatusNoContent)

		e.DELETE("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + oidcAttributeID).
			Expect().
			Status(http.StatusNoContent)

		e.DELETE("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + usernameAttributeID).
			Expect().
			Status(http.StatusNoContent)
	})

	// Verify no attributes remain
	t.Run("Verify No Attributes Remain", func(t *testing.T) {
		resp := e.GET("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		resp.Length().IsEqual(0)
	})

	// Test creating attribute with null index (backup email)
	t.Run("Create Attribute With Null Index", func(t *testing.T) {
		backupEmailAttribute := map[string]interface{}{
			"type":  model.AttributeTypeEmail,
			"index": nil, // No index for backup email
			"value": map[string]interface{}{
				"email":    "backup@example.com",
				"verified": false,
			},
		}

		resp := e.POST("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes").
			WithJSON(backupEmailAttribute).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		resp.HasValue("type", model.AttributeTypeEmail).
			Value("value").Object().HasValue("email", "backup@example.com")

		// Index should be omitted from response if null
		resp.NotContainsKey("index")

		backupID := resp.Value("id").String().Raw()

		// Clean up
		e.DELETE("/admin/" + tenant + "/" + realm + "/users/" + testUserID + "/attributes/" + backupID).
			Expect().
			Status(http.StatusNoContent)
	})
}
