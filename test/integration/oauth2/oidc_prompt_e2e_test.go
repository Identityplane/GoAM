package integration

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/test/integration"
)

// This test performs end-to-end tests of the OIDC prompt parameter functionality.
// It tests various combinations of prompt parameter values and flow configurations:
// 1. prompt=none with noprompt flow -> Should directly get authorization code
// 2. prompt=login with prompt flow -> Should fail with login_required
// 3. prompt=login with noprompt flow -> Should fail with login_required
// 4. prompt=login with prompt flow -> Should redirect to login page

func TestOIDCPrompt_E2E(t *testing.T) {
	// Get the base httpexpect instance with in-memory listener
	e := integration.SetupIntegrationTest(t, "")

	// Common test data
	clientID := "management-ui"
	redirectURI := "http://localhost:3000"
	codeChallenge := "wbwbDjIP3vh12aP80FhFzWw1SB0pkjiemAQ-N-GuPCI"
	codeChallengeMethod := "S256"
	state := "0b1d7e0fd86540daa9b9000b8ccf2e5d"
	scope := "openid profile write:flows write:realms write:applications write:user"

	// Flow names
	const (
		flowNoPrompt = "noprompt"
		flowPrompt   = "login_or_register"
	)

	// Create a user for the test
	service.GetServices().UserService.CreateUser(context.Background(), "acme", "customers", model.User{
		ID: "admin",
	})

	// Test Case 1: prompt=none with noprompt flow
	t.Run("Prompt None with NoPrompt Flow", func(t *testing.T) {
		resp := e.GET("/acme/customers/oauth2/authorize").
			WithQuery("client_id", clientID).
			WithQuery("redirect_uri", redirectURI).
			WithQuery("response_type", "code").
			WithQuery("scope", scope).
			WithQuery("state", state).
			WithQuery("code_challenge", codeChallenge).
			WithQuery("code_challenge_method", codeChallengeMethod).
			WithQuery("prompt", "none").
			WithQuery("flow", flowNoPrompt).
			Expect().
			Status(http.StatusSeeOther)

		// Verify we get a valid authorization code
		redirectURL, err := url.Parse(resp.Header("Location").Raw())
		if err != nil {
			t.Fatalf("Failed to parse redirect URL: %v", err)
		}

		code := redirectURL.Query().Get("code")
		if code == "" {
			t.Fatal("Expected authorization code but got none")
		}
	})

	// Test Case 2: prompt=login with prompt flow
	t.Run("Prompt Login with Prompt Flow - Should redirect to login page", func(t *testing.T) {
		resp := e.GET("/acme/customers/oauth2/authorize").
			WithQuery("client_id", clientID).
			WithQuery("redirect_uri", redirectURI).
			WithQuery("response_type", "code").
			WithQuery("scope", scope).
			WithQuery("state", state).
			WithQuery("code_challenge", codeChallenge).
			WithQuery("code_challenge_method", codeChallengeMethod).
			WithQuery("prompt", "login").
			WithQuery("flow", flowPrompt).
			Expect().
			Status(http.StatusSeeOther)

		// Verify we get a login_required error
		redirectURL, err := url.Parse(resp.Header("Location").Raw())
		if err != nil {
			t.Fatalf("Failed to parse redirect URL: %v", err)
		}

		error := redirectURL.Query().Get("error")
		if error != "" {
			t.Fatalf("received an Oauth error but exptected none '%s'", error)
		}
	})

	// Test Case 3: prompt=login with noprompt flow
	t.Run("Prompt Login with NoPrompt Flow - Should Fail", func(t *testing.T) {
		resp := e.GET("/acme/customers/oauth2/authorize").
			WithQuery("client_id", clientID).
			WithQuery("redirect_uri", redirectURI).
			WithQuery("response_type", "code").
			WithQuery("scope", scope).
			WithQuery("state", state).
			WithQuery("code_challenge", codeChallenge).
			WithQuery("code_challenge_method", codeChallengeMethod).
			WithQuery("prompt", "login").
			WithQuery("flow", flowNoPrompt).
			Expect().
			Status(http.StatusSeeOther)

		// Verify we get a login_required error
		redirectURL, err := url.Parse(resp.Header("Location").Raw())
		if err != nil {
			t.Fatalf("Failed to parse redirect URL: %v", err)
		}

		error := redirectURL.Query().Get("error")
		if error != "server_error" {
			t.Fatalf("Expected error 'server_error' but got '%s'", error)
		}
	})

	// Test Case 4: prompt=login with prompt flow
	t.Run("Prompt Login with Prompt Flow - Should Redirect to Login", func(t *testing.T) {
		resp := e.GET("/acme/customers/oauth2/authorize").
			WithQuery("client_id", clientID).
			WithQuery("redirect_uri", redirectURI).
			WithQuery("response_type", "code").
			WithQuery("scope", scope).
			WithQuery("state", state).
			WithQuery("code_challenge", codeChallenge).
			WithQuery("code_challenge_method", codeChallengeMethod).
			WithQuery("prompt", "login").
			WithQuery("flow", flowPrompt).
			Expect().
			Status(http.StatusSeeOther)

		// Get session cookie
		sessionCookie := resp.Cookie("session_id")
		if sessionCookie == nil {
			t.Fatal("No session cookie found")
		}

		// Verify we get redirected to the login page
		location := resp.Header("Location").Raw()
		if !strings.Contains(location, "/acme/customers/auth/login-or-register") {
			t.Fatalf("Expected redirect to login page but got: %s", location)
		}
	})
}
