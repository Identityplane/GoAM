package integration_admin_api

import (
	"context"
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/test/integration"

	"github.com/stretchr/testify/assert"
)

func TestAdminAuthzE2E(t *testing.T) {
	// Get the base httpexpect instance with in-memory listener
	e := integration.SetupIntegrationTest(t, "")
	config.ServerSettings.UnsafeDisableAdminAuth = false

	// An anonymous call use whoami should return an unauthorized error
	t.Run("Check Unauthorized Endpoints", func(t *testing.T) {
		e.GET("/admin/whoami").
			Expect().
			Status(http.StatusUnauthorized)

		// Same for list realms
		e.GET("/admin/realms").
			Expect().
			Status(http.StatusUnauthorized)
	})

	// Create test user in internal realm
	testUser := model.User{
		Tenant: "internal",
		Realm:  "internal",
		ID:     "testadmin",
		Status: "active",
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

		// Verify entitlements
		entitlements := resp.Value("entitlements").Array()
		entitlements.Length().IsEqual(0)
	})

	t.Run("Check List Realms Before Tenant Creation", func(t *testing.T) {
		e.GET("/admin/realms").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Array().Length().IsEqual(0)
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

		// Verify entitlements now include the new tenant
		entitlements := resp.Value("entitlements").Array()
		entitlements.Length().IsEqual(1)

		entitlement := entitlements.Element(0).Object()
		entitlement.Value("tenant").String().IsEqual("test-tenant")
		entitlement.Value("realm").String().IsEqual("*")
		entitlement.Value("scopes").Array().Length().IsEqual(1)
		entitlement.Value("scopes").Array().Element(0).String().IsEqual("*")
	})

	t.Run("Check List Realms After Tenant Creation", func(t *testing.T) {
		e.GET("/admin/realms").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Array().Length().IsEqual(1)
	})
}
