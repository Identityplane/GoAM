package node_github

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupGithubMockServer(t *testing.T, mockGithubUserId int64) *httptest.Server {
	githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/login/oauth/access_token":
			// Verify request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Failed to read request body: %v", err)
			}

			// Parse the request body
			var requestBody map[string]string
			if err := json.Unmarshal(body, &requestBody); err != nil {
				t.Fatalf("Failed to parse request body: %v", err)
			}

			// Verify required fields
			if requestBody["client_id"] != "test_client_id" ||
				requestBody["client_secret"] != "test_client_secret" ||
				requestBody["code"] != "test_code" {
				t.Fatalf("Invalid request body: %s", string(body))
			}

			// Mock token response
			response := githubAccessTokenResponse{
				AccessToken:  "mock_access_token",
				RefreshToken: "mock_refresh_token",
				TokenType:    "bearer",
				Scope:        "user",
			}
			json.NewEncoder(w).Encode(response)
		case "/user":
			// Verify Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "token mock_access_token" {
				t.Fatalf("Invalid Authorization header: %s", authHeader)
			}

			// Mock user data response
			response := GitHubUser{
				Login:     "testuser",
				ID:        mockGithubUserId,
				AvatarURL: "https://github.com/avatar.png",
				Email:     "test@example.com",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	return githubServer
}
