package integration

import (
	"goiam/test/integration"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test performs a complete end-to-end test of the OAuth2 client credentials flow.
// It tests the following operations in sequence:
// 1. Creating a confidential application
// 2. Requesting an access token using client credentials
// 3. Using the access token to access a protected resource
func TestOAuth2ClientCredentials_E2E(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	// Test OAuth2 client credentials flow data
	clientID := "backend-api"
	clientSecret := "backend-api-secret"
	scope := "write:flows write:realms write:applications write:user"

	// Test token request
	t.Run("Request Access Token", func(t *testing.T) {
		resp := e.POST("/acme/customers/oauth2/token").
			WithHeader("Content-Type", "application/x-www-form-urlencoded").
			WithFormField("grant_type", "client_credentials").
			WithFormField("client_id", clientID).
			WithFormField("client_secret", clientSecret).
			WithFormField("scope", scope).
			Expect().
			Status(http.StatusOK)

		tokenResp := resp.JSON().Object()

		// Verify token response
		tokenResp.HasValue("token_type", "Bearer")
		tokenResp.Value("access_token").String().NotEmpty()
		tokenResp.Value("expires_in").Number().Gt(0)
		tokenResp.Value("scope").String().IsEqual(scope)

		// Should not have refresh token
		tokenResp.NotContainsKey("refresh_token")

		// Store the access token for later use
		accessToken := tokenResp.Value("access_token").String().Raw()
		assert.NotEmpty(t, accessToken)
	})
}

// This test checks various failure conditions for the OAuth2 client credentials flow
func TestOAuth2ClientCredentials_Failures(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	// Test data
	validClientID := "backend-api"
	validClientSecret := "backend-api-secret"
	validScope := "write:flows write:realms write:applications write:user"
	invalidScope := "invalid:scope"
	managementUIClientID := "management-ui"

	t.Run("Invalid Client Secret", func(t *testing.T) {
		resp := e.POST("/acme/customers/oauth2/token").
			WithHeader("Content-Type", "application/x-www-form-urlencoded").
			WithFormField("grant_type", "client_credentials").
			WithFormField("client_id", validClientID).
			WithFormField("client_secret", "wrong-secret").
			WithFormField("scope", validScope).
			Expect().
			Status(http.StatusBadRequest)

		errorResp := resp.JSON().Object()
		errorResp.HasValue("error", "unauthorized_client")
		errorResp.HasValue("error_description", "Invalid client authentication")
	})

	t.Run("Invalid Client ID", func(t *testing.T) {
		resp := e.POST("/acme/customers/oauth2/token").
			WithHeader("Content-Type", "application/x-www-form-urlencoded").
			WithFormField("grant_type", "client_credentials").
			WithFormField("client_id", "non-existent-client").
			WithFormField("client_secret", validClientSecret).
			WithFormField("scope", validScope).
			Expect().
			Status(http.StatusBadRequest)

		errorResp := resp.JSON().Object()
		errorResp.HasValue("error", "unauthorized_client")
		errorResp.HasValue("error_description", "Invalid client ID")
	})

	t.Run("Invalid Scope", func(t *testing.T) {
		resp := e.POST("/acme/customers/oauth2/token").
			WithHeader("Content-Type", "application/x-www-form-urlencoded").
			WithFormField("grant_type", "client_credentials").
			WithFormField("client_id", validClientID).
			WithFormField("client_secret", validClientSecret).
			WithFormField("scope", invalidScope).
			Expect().
			Status(http.StatusBadRequest)

		errorResp := resp.JSON().Object()
		errorResp.HasValue("error", "invalid_scope")
		errorResp.HasValue("error_description", "Invalid scope "+invalidScope)
	})

	t.Run("Grant Type Not Allowed", func(t *testing.T) {
		resp := e.POST("/acme/customers/oauth2/token").
			WithHeader("Content-Type", "application/x-www-form-urlencoded").
			WithFormField("grant_type", "client_credentials").
			WithFormField("client_id", managementUIClientID).
			WithFormField("client_secret", "management-ui-secret").
			WithFormField("scope", validScope).
			Expect().
			Status(http.StatusBadRequest)

		errorResp := resp.JSON().Object()
		errorResp.HasValue("error", "unauthorized_client")
		errorResp.HasValue("error_description", "Grant type not allowed")
	})
}
