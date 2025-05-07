package integration

import (
	"goiam/test/integration"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This test performs a complete end-to-end test of the OAuth2 PKCE flow.
// It tests the following operations in sequence:
// 1. Creating the management-ui application
// 2. Starting the OAuth2 authorization flow with PKCE
// 3. Authenticating the user
// 4. Completing the authorization
// 5. Exchanging the authorization code for tokens
// 6. Getting user information using the access token
// The test uses the management-ui client and tests the complete flow.

func TestOAuth2PKCE_E2E(t *testing.T) {
	// Get the base httpexpect instance with in-memory listener
	e := integration.SetupIntegrationTest(t, "")

	// Test OAuth2 PKCE flow data
	clientID := "management-ui"
	redirectURI := "http://localhost:3000"
	codeChallenge := "wbwbDjIP3vh12aP80FhFzWw1SB0pkjiemAQ-N-GuPCI"
	codeChallengeMethod := "S256"
	state := "0b1d7e0fd86540daa9b9000b8ccf2e5d"
	scope := "openid profile write:flows write:realms write:applications write:user"
	codeVerifier := "4f13a3d4d18440e490c0ebab1831820eba8d51cbfe4444e5b2d34423fdddbb6b00def2eb5b9e4712b453fe4f71230fa5"

	// Test starting OAuth2 authorization
	t.Run("Start OAuth2 Authorization", func(t *testing.T) {
		resp := e.GET("/acme/customers/oauth2/authorize").
			WithQuery("client_id", clientID).
			WithQuery("redirect_uri", redirectURI).
			WithQuery("response_type", "code").
			WithQuery("scope", scope).
			WithQuery("state", state).
			WithQuery("code_challenge", codeChallenge).
			WithQuery("code_challenge_method", codeChallengeMethod).
			Expect().
			Status(http.StatusFound)

		assert.Empty(t, resp.Header("Access-Control-Allow-Origin").Raw(), "CORS header should not exist")

		// Get session cookie
		sessionCookie := resp.Cookie("session_id")
		if sessionCookie == nil {
			t.Fatal("No session cookie found")
		}

		// Test user authentication
		t.Run("Authenticate User", func(t *testing.T) {

			e.GET("/acme/customers/auth/login-or-register").
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusOK)

			e.POST("/acme/customers/auth/login-or-register").
				WithHeader("Content-Type", "application/x-www-form-urlencoded").
				WithFormField("step", "askUsername").
				WithFormField("username", "foobar").
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusOK)

			e.POST("/acme/customers/auth/login-or-register").
				WithHeader("Content-Type", "application/x-www-form-urlencoded").
				WithFormField("step", "node_a1e9d8fa").
				WithFormField("confirmation", "true").
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusOK)

			e.POST("/acme/customers/auth/login-or-register").
				WithHeader("Content-Type", "application/x-www-form-urlencoded").
				WithFormField("step", "node_26e37459").
				WithFormField("password", "foobar").
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusFound).
				Header("Location").IsEqual("http://localhost:8080/acme/customers/oauth2/finishauthorize")

		})

		// Test completing authorization and capture the auth code
		var authCode string
		t.Run("Complete Authorization", func(t *testing.T) {
			resp := e.GET("/acme/customers/oauth2/finishauthorize").
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusFound).
				Header("Location")

			// Parse the redirect URL to get the authorization code
			redirectURL, err := url.Parse(resp.Raw())
			if err != nil {
				t.Fatalf("Failed to parse redirect URL: %v", err)
			}

			// Extract the code from the query parameters
			code := redirectURL.Query().Get("code")
			if code == "" {
				t.Fatal("No authorization code found in redirect URL")
			}
			authCode = code
		})

		var accessToken string
		var refreshToken string
		// Test token exchange
		t.Run("Exchange Code for Token", func(t *testing.T) {
			resp := e.POST("/acme/customers/oauth2/token").
				WithHeader("Content-Type", "application/x-www-form-urlencoded").
				WithFormField("grant_type", "authorization_code").
				WithFormField("redirect_uri", redirectURI).
				WithFormField("code", authCode).
				WithFormField("code_verifier", codeVerifier).
				WithFormField("client_id", clientID).
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusOK)

			// Check cors
			assert.NotEmpty(t, resp.Header("Access-Control-Allow-Origin").Raw(), "CORS header should exist")

			tokenResp := resp.JSON().Object()

			// Verify token response
			tokenResp.HasValue("token_type", "Bearer")
			tokenResp.Value("id_token").String().NotEmpty()
			accessToken = tokenResp.Value("access_token").String().NotEmpty().Raw()
			refreshToken = tokenResp.Value("refresh_token").String().NotEmpty().Raw()
		})

		// Test getting user info
		t.Run("Get User Info", func(t *testing.T) {

			// userinfo must support OPTIONS request and Access-Control-Allow-Headers: Content-Type, Authorization
			resp := e.OPTIONS("/acme/customers/oauth2/userinfo").
				WithHeader("Access-Control-Request-Headers", "Content-Type, Authorization").
				Expect().
				Status(http.StatusOK)

			assert.NotEmpty(t, resp.Header("Access-Control-Allow-Headers").Raw(), "CORS header should exist")

			resp = e.GET("/acme/customers/oauth2/userinfo").
				WithHeader("Authorization", "Bearer "+accessToken).
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusOK)

			// Check cors
			assert.NotEmpty(t, resp.Header("Access-Control-Allow-Origin").Raw(), "CORS header should exist")

			userInfoResp := resp.JSON().Object()

			userInfoResp.Value("sub").String().NotEmpty()

		})

		// Test refreshing the access token
		t.Run("Refresh Access Token", func(t *testing.T) {
			resp := e.POST("/acme/customers/oauth2/token").
				WithHeader("Content-Type", "application/x-www-form-urlencoded").
				WithFormField("grant_type", "refresh_token").
				WithFormField("refresh_token", refreshToken).
				WithFormField("client_id", clientID).
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object()

			// Verify token response
			resp.HasValue("token_type", "Bearer")

			// Check that the id token is empty when refreshing the access token
			resp.NotContainsKey("id_token")

			new_accessToken := resp.Value("access_token").String().NotEmpty().Raw()
			new_refreshToken := resp.Value("refresh_token").String().NotEmpty().Raw()

			// Check that the new access token is different from the old one
			if new_accessToken == accessToken {
				t.Fatal("New access token is the same as the old one")
			}

			// Check that the new refresh token is different from the old one
			if new_refreshToken == refreshToken {
				t.Fatal("New refresh token is the same as the old one")
			}

		})

	})
}
