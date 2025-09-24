package authhandler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
	"github.com/gavv/httpexpect/v2"
)

// JSON API request/response structures
type FlowRequest struct {
	ExecutionID string            `json:"executionId"`
	SessionID   string            `json:"sessionId"`
	CurrentNode string            `json:"currentNode"`
	Responses   map[string]string `json:"responses"`
}

type FlowResponse struct {
	ExecutionID string            `json:"executionId"`
	SessionID   string            `json:"sessionId"`
	CurrentNode string            `json:"currentNode"`
	Prompts     map[string]string `json:"prompts,omitempty"`
	Result      *FlowResult       `json:"result,omitempty"`
	Debug       any               `json:"debug,omitempty"`
}

type FlowResult struct {
	Status                string `json:"status"`
	Message               string `json:"message"`
	AccessToken           string `json:"access_token,omitempty"`
	TokenType             string `json:"token_type,omitempty"`
	RefreshToken          string `json:"refresh_token,omitempty"`
	ExpiresIn             int    `json:"expires_in,omitempty"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in,omitempty"`
	User                  *User  `json:"user,omitempty"`
}

type User struct {
	Sub string `json:"sub"`
}

func TestJSONFlow_MockSuccessFlow(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	t.Run("Mock Success Flow", func(t *testing.T) {
		// Test the mock flow that should complete immediately with OAuth2 tokens
		resp := e.GET("/acme/customers/api/v1/mock-success").
			WithQuery("client_id", "customers-app").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON()

		// Validate response structure - mock flow should complete immediately
		resp.Object().
			HasValue("currentNode", "successResult").
			Value("executionId").String().NotEmpty()

		resp.Object().Value("sessionId").String().NotEmpty()

		// Check if result contains OAuth2 tokens
		if resp.Object().Value("result").Raw() != nil {
			result := resp.Object().Value("result").Object()

			// Check for OAuth2 tokens
			result.Value("access_token").String().NotEmpty()
			result.Value("token_type").String().Equal("Bearer")
			result.Value("refresh_token").String().NotEmpty()
			result.Value("expires_in").Number().Equal(3600)
			result.Value("refresh_token_expires_in").Number().Equal(31536000)

			// Validate user object
			user := result.Value("user").Object()
			user.Value("sub").String().NotEmpty()
		}
	})
}

func TestJSONFlow_UsernamePasswordRegisterFlow(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	var executionID, sessionID string

	// Step 1: GET request - expect JSON with username prompt
	t.Run("Step 1: GET - Username Prompt", func(t *testing.T) {
		resp := e.GET("/acme/customers/api/v1/username-password-register").
			WithQuery("client_id", "customers-app").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON()

		// Validate response structure
		resp.Object().
			HasValue("currentNode", "askUsername").
			Value("executionId").String().NotEmpty()

		resp.Object().Value("sessionId").String().NotEmpty()

		resp.Object().Value("prompts").Object().
			HasValue("username", "text")

		// Extract IDs for next request
		executionID = resp.Object().Value("executionId").String().Raw()
		sessionID = resp.Object().Value("sessionId").String().Raw()
	})

	// Step 2: POST username - expect JSON with password prompt
	t.Run("Step 2: POST Username - Password Prompt", func(t *testing.T) {
		request := FlowRequest{
			ExecutionID: executionID,
			SessionID:   sessionID,
			CurrentNode: "askUsername",
			Responses: map[string]string{
				"username": "testuser",
			},
		}

		resp := e.POST("/acme/customers/api/v1/username-password-register").
			WithHeader("Content-Type", "application/json").
			WithJSON(request).
			Expect().
			Status(http.StatusOK).
			JSON()

		// Validate response structure
		resp.Object().
			HasValue("currentNode", "askPassword").
			Value("executionId").String().Equal(executionID)

		resp.Object().Value("sessionId").String().Equal(sessionID)

		resp.Object().Value("prompts").Object().
			HasValue("password", "password")
	})

	// Step 3: POST password - expect success result with OAuth2 tokens
	t.Run("Step 3: POST Password - Success Result with OAuth2 Tokens", func(t *testing.T) {
		request := FlowRequest{
			ExecutionID: executionID,
			SessionID:   sessionID,
			CurrentNode: "askPassword",
			Responses: map[string]string{
				"password": "testuser",
			},
		}

		resp := e.POST("/acme/customers/api/v1/username-password-register").
			WithQuery("client_id", "customers-app").
			WithHeader("Content-Type", "application/json").
			WithJSON(request).
			Expect().
			Status(http.StatusOK).
			JSON()

		// Validate success response
		resp.Object().
			HasValue("currentNode", "registerSuccess").
			Value("executionId").String().Equal(executionID)

		resp.Object().Value("sessionId").String().Equal(sessionID)

		// Check if result contains OAuth2 tokens
		if resp.Object().Value("result").Raw() != nil {
			result := resp.Object().Value("result").Object()

			// Check for OAuth2 tokens
			result.Value("access_token").String().NotEmpty()
			result.Value("token_type").String().Equal("Bearer")
			result.Value("refresh_token").String().NotEmpty()
			result.Value("expires_in").Number().Equal(3600)
			result.Value("refresh_token_expires_in").Number().Equal(31536000)

			// Validate user object
			user := result.Value("user").Object()
			user.Value("sub").String().NotEmpty()
		}
	})

}

func TestJSONFlow_ErrorHandling(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	t.Run("JSON API Error Handling", func(t *testing.T) {
		// Test missing client_id
		t.Run("Missing Client ID", func(t *testing.T) {
			e.GET("/acme/customers/api/v1/mock-success").
				WithHeader("Accept", "application/json").
				Expect().
				Status(http.StatusBadRequest).
				JSON().
				Object().
				Value("error").Object().
				HasValue("error", "INVALID_REQUEST").
				HasValue("error_description", "Client ID is required")
		})
	})
}

// Helper function to parse JSON response
func parseJSONResponse(t *testing.T, resp *httpexpect.Object) FlowResponse {
	var flowResp FlowResponse

	// Convert httpexpect.Object to JSON bytes
	rawData := resp.Raw()

	// Convert map to JSON bytes
	jsonBytes, err := json.Marshal(rawData)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	err = json.Unmarshal(jsonBytes, &flowResp)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	return flowResp
}
