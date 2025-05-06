package integration

import (
	"goiam/test/integration"
	"net/http"
	"net/url"
	"testing"
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

	// Create the management-ui application
	t.Run("Create Management UI Application", func(t *testing.T) {
		appData := map[string]interface{}{
			"tenant":                       "acme",
			"realm":                        "customers",
			"client_id":                    clientID,
			"confidential":                 false,
			"consent_required":             false,
			"description":                  "Management UI Application",
			"allowed_scopes":               []string{"openid", "profile", "write:user", "write:flows", "write:realms", "write:applications"},
			"allowed_grants":               []string{"authorization_code_pkce"},
			"allowed_authentication_flows": []string{"login_or_register"},
			"access_token_lifetime":        600,
			"refresh_token_lifetime":       3600,
			"id_token_lifetime":            600,
			"access_token_type":            "session",
			"access_token_algorithm":       "",
			"access_token_mapping":         "",
			"id_token_algorithm":           "",
			"id_token_mapping":             "",
			"redirect_uris":                []string{redirectURI},
			"created_at":                   "2025-05-03T13:39:12+05:30",
			"updated_at":                   "2025-05-06T00:59:35+08:00",
		}

		e.POST("/admin/acme/customers/applications/" + clientID).
			WithJSON(appData).
			Expect().
			Status(http.StatusCreated)
	})

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
				Status(http.StatusOK).
				JSON().
				Object()

			// Verify token response
			resp.HasValue("token_type", "Bearer")
			resp.Value("id_token").String().NotEmpty()
			accessToken = resp.Value("access_token").String().NotEmpty().Raw()
			//resp.Value("refresh_token").String().NotEmpty()
		})

		// Test getting user info
		t.Run("Get User Info", func(t *testing.T) {

			e.GET("/acme/customers/oauth2/userinfo").
				WithHeader("Authorization", "Bearer "+accessToken).
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object().
				Value("sub").String().NotEmpty()
		})
	})
}
