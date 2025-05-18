package integration

import (
	"context"
	"goiam/internal/config"
	"goiam/internal/model"
	"goiam/internal/service"
	"goiam/test/integration"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminAuthzE2E(t *testing.T) {
	// Get the base httpexpect instance with in-memory listener
	e := integration.SetupIntegrationTest(t, "")
	config.UnsafeDisableAdminAuthzCheck = false

	// Create test user in internal realm
	testUser := model.User{
		Tenant:          "internal",
		Realm:           "internal",
		ID:              "testadmin",
		Username:        "testadmin",
		DisplayName:     "Test Admin",
		Email:           "testadmin@example.com",
		EmailVerified:   true,
		Status:          "active",
		LoginIdentifier: "testadmin@example.com",
		Entitlements:    []string{}, // Give access to internal realm
	}

	services := service.GetServices()
	services.UserService.CreateUser(context.Background(), "internal", "internal", testUser)

	accessToken := integration.CreateAccessTokenSession(t, testUser)

	// Test whoami endpoint before tenant creation
	t.Run("Check WhoAmI Before Tenant Creation", func(t *testing.T) {
		resp := e.GET("/admin/whoami").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		// Verify user info
		user := resp.Value("user").Object()
		user.Value("id").String().IsEqual(testUser.ID)
		user.Value("display_name").String().IsEqual(testUser.DisplayName)
		user.Value("email").String().IsEqual(testUser.Email)

		// Verify entitlements
		entitlements := resp.Value("entitlements").Array()
		entitlements.Length().IsEqual(0)
	})

	// Create a new tenant
	t.Run("Create New Tenant", func(t *testing.T) {
		req := struct {
			TenantSlug string `json:"tenant_slug"`
			TenantName string `json:"tenant_name"`
		}{
			TenantSlug: "test-tenant",
			TenantName: "Test Tenant",
		}

		resp := e.POST("/admin/tenants").
			WithHeader("Authorization", "Bearer "+accessToken).
			WithJSON(req).
			Expect().
			Status(http.StatusCreated)

		// Verify response
		resp.JSON().Object().HasValue("tenant_slug", "test-tenant")

		// Verify that realm default was created
		realm, found := services.RealmService.GetRealm("test-tenant", "default")
		assert.True(t, found)
		assert.NotNil(t, realm)
	})

	// Test whoami endpoint after tenant creation
	t.Run("Check WhoAmI After Tenant Creation", func(t *testing.T) {
		resp := e.GET("/admin/whoami").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		// Verify user info
		user := resp.Value("user").Object()
		user.Value("id").String().IsEqual(testUser.ID)
		user.Value("display_name").String().IsEqual(testUser.DisplayName)
		user.Value("email").String().IsEqual(testUser.Email)

		// Verify entitlements now include the new tenant
		entitlements := resp.Value("entitlements").Array()
		entitlements.Length().IsEqual(1)

		entitlement := entitlements.Element(0).Object()
		entitlement.Value("tenant").String().IsEqual("test-tenant")
		entitlement.Value("realm").String().IsEqual("*")
		entitlement.Value("scopes").Array().Length().IsEqual(1)
		entitlement.Value("scopes").Array().Element(0).String().IsEqual("*")
	})
}
