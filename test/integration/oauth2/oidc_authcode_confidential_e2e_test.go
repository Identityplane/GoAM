package integration

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"

	"github.com/stretchr/testify/assert"
)

// This test performs a complete end-to-end test of the OAuth2 Authorization Code flow for a confidential client (backend-api).
// It tests the following operations in sequence:
// 1. Starting the OAuth2 authorization flow
// 2. Authenticating the user
// 3. Completing the authorization
// 4. Exchanging the authorization code for tokens (with client_secret)
// 5. Getting user information using the access token
// 6. Introspecting the access token
func TestOAuth2AuthCodeConfidential_E2E(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	clientID := "backend-api"
	clientSecret := "backend-api-secret"
	redirectURI := "http://localhost:3000"
	state := "0b1d7e0fd86540daa9b9000b8ccf2e5d"
	scope := "openid write:flows write:realms write:applications write:user"

	t.Run("Start OAuth2 Authorization", func(t *testing.T) {
		resp := e.GET("/acme/customers/oauth2/authorize").
			WithQuery("client_id", clientID).
			WithQuery("redirect_uri", redirectURI).
			WithQuery("response_type", "code").
			WithQuery("scope", scope).
			WithQuery("state", state).
			Expect().
			Status(http.StatusSeeOther)

		sessionCookie := resp.Cookie("session_id")
		if sessionCookie == nil {
			t.Fatal("No session cookie found")
		}
		t.Run("Authenticate User", func(t *testing.T) {
			cookieValue := sessionCookie.Value().Raw()

			// Resume the OAuth2-created auth session via JSON API
			getResp := e.GET("/acme/customers/api/v1/login-or-register").
				WithQuery("continue", "true").
				WithCookie("session_id", cookieValue).
				WithHeader("Accept", "application/json").
				Expect().
				Status(http.StatusOK).
				JSON()

			sessionID := getResp.Object().Value("sessionId").String().Raw()
			currentNode := getResp.Object().Value("currentNode").String().Raw()
			if sessionID == "" {
				t.Fatal("No sessionId found in JSON response")
			}
			assert.Equal(t, "askUsername", currentNode)

			// Submit username
			usernameResp := e.POST("/acme/customers/api/v1/login-or-register").
				WithHeader("Content-Type", "application/json").
				WithJSON(map[string]any{
					"sessionId":   sessionID,
					"currentNode": "askUsername",
					"responses": map[string]string{
						"username": "foobar",
					},
				}).
				WithCookie("session_id", cookieValue).
				Expect().
				Status(http.StatusOK).
				JSON()

			nextNode := usernameResp.Object().Value("currentNode").String().Raw()
			if prompts, ok := usernameResp.Object().Value("prompts").Raw().(map[string]any); ok {
				if _, hasConfirmation := prompts["confirmation"]; hasConfirmation {
					confirmResp := e.POST("/acme/customers/api/v1/login-or-register").
						WithHeader("Content-Type", "application/json").
						WithJSON(map[string]any{
							"sessionId":   sessionID,
							"currentNode": nextNode,
							"responses": map[string]string{
								"confirmation": "true",
							},
						}).
						WithCookie("session_id", cookieValue).
						Expect().
						Status(http.StatusOK).
						JSON()
					nextNode = confirmResp.Object().Value("currentNode").String().Raw()
				}
			}

			// Submit password (should finish and set redirect to finishauthorize)
			passwordResp := e.POST("/acme/customers/api/v1/login-or-register").
				WithHeader("Content-Type", "application/json").
				WithJSON(map[string]any{
					"sessionId":   sessionID,
					"currentNode": nextNode,
					"responses": map[string]string{
						"password": "foobar",
					},
				}).
				WithCookie("session_id", cookieValue).
				Expect().
				Status(http.StatusOK).
				JSON()

			passwordResp.Object().Value("result").Object().
				Value("redirect").String().Contains("/acme/customers/oauth2/finishauthorize")
		})

		var authCode string
		t.Run("Complete Authorization", func(t *testing.T) {
			resp := e.GET("/acme/customers/oauth2/finishauthorize").
				WithCookie("session_id", sessionCookie.Value().Raw()).
				Expect().
				Status(http.StatusSeeOther).
				Header("Location")

			redirectURL, err := url.Parse(resp.Raw())
			if err != nil {
				t.Fatalf("Failed to parse redirect URL: %v", err)
			}

			code := redirectURL.Query().Get("code")
			if code == "" {
				t.Fatal("No authorization code found in redirect URL")
			}
			authCode = code
		})

		var accessToken string
		t.Run("Exchange Code for Token", func(t *testing.T) {
			resp := e.POST("/acme/customers/oauth2/token").
				WithHeader("Content-Type", "application/x-www-form-urlencoded").
				WithFormField("grant_type", "authorization_code").
				WithFormField("redirect_uri", redirectURI).
				WithFormField("code", authCode).
				WithFormField("client_id", clientID).
				WithFormField("client_secret", clientSecret).
				Expect().
				Status(http.StatusOK)

			assert.NotEmpty(t, resp.Header("Access-Control-Allow-Origin").Raw(), "CORS header should exist")

			tokenResp := resp.JSON().Object()

			tokenResp.HasValue("token_type", "Bearer")
			tokenResp.Value("id_token").String().NotEmpty()
			accessToken = tokenResp.Value("access_token").String().NotEmpty().Raw()
		})

		t.Run("Get User Info", func(t *testing.T) {
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

			assert.NotEmpty(t, resp.Header("Access-Control-Allow-Origin").Raw(), "CORS header should exist")

			userInfoResp := resp.JSON().Object()
			userInfoResp.Value("sub").String().NotEmpty()
		})

		t.Run("Introspect Access Token", func(t *testing.T) {
			resp := e.POST("/acme/customers/oauth2/introspect").
				WithHeader("Content-Type", "application/x-www-form-urlencoded").
				WithFormField("token", accessToken).
				WithFormField("token_type_hint", "access_token").
				WithFormField("client_id", clientID).
				WithFormField("client_secret", clientSecret).
				Expect().
				Status(http.StatusOK)

			assert.NotEmpty(t, resp.Header("Access-Control-Allow-Origin").Raw(), "CORS header should exist")

			introspectionResp := resp.JSON().Object()
			introspectionResp.HasValue("active", true)
			introspectionResp.HasValue("token_type", "Bearer")
			introspectionResp.HasValue("client_id", clientID)
			introspectionResp.HasValue("scope", scope)
			introspectionResp.Value("exp").Number().Gt(0)
			introspectionResp.Value("iat").Number().Gt(0)
			introspectionResp.Value("nbf").Number().Gt(0)
			introspectionResp.Value("jti").String().NotEmpty()
		})
	})
}
