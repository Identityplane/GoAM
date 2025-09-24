package authhandler

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/test/integration"
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

	tenant := "acme"
	realm := "customers"
	clientID := "customers-app"
	flowID := "mock_success"
	application, ok := service.GetServices().ApplicationService.GetApplication(tenant, realm, clientID)
	if !ok {
		t.Fatalf("Application not found")
	}
	flow, ok := service.GetServices().FlowService.GetFlowById(tenant, realm, flowID)
	if !ok {
		t.Fatalf("Flow not found")
	}

	t.Run("Mock Success Flow", func(t *testing.T) {
		// Test the mock flow that should complete immediately with OAuth2 tokens
		resp := e.GET("/acme/customers/api/v1/"+flow.Route).
			WithQuery("client_id", application.ClientId).
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
			result.Value("success").Boolean().IsEqual(true)
			result.Value("access_token").String().NotEmpty()
			result.Value("token_type").String().IsEqual("Bearer")
			result.Value("refresh_token").String().NotEmpty()
			result.Value("expires_in").Number().IsEqual(application.AccessTokenLifetime)
			result.Value("refresh_token_expires_in").Number().IsEqual(application.RefreshTokenLifetime)

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
	t.Run("Username Password Register Flow Step 1: GET - Username Prompt", func(t *testing.T) {
		resp := e.GET("/acme/customers/api/v1/username-password-register").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON()

		// Validate response structure
		resp.Object().HasValue("currentNode", "askUsername").Value("executionId").String().NotEmpty()

		resp.Object().Value("sessionId").String().NotEmpty()

		resp.Object().Value("prompts").Object().HasValue("username", "text")

		// Extract IDs for next request
		executionID = resp.Object().Value("executionId").String().Raw()
		sessionID = resp.Object().Value("sessionId").String().Raw()
	})

	t.Run("Username Password Register Flow Step 2: POST Username - Password Prompt", func(t *testing.T) {
		request := FlowRequest{
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
			Value("executionId").String().IsEqual(executionID)

		resp.Object().Value("sessionId").String().IsEqual(sessionID)

		resp.Object().Value("prompts").Object().HasValue("password", "password")
	})

	t.Run("Username Password Register Flow Step 3: POST Password - Success Result", func(t *testing.T) {
		request := FlowRequest{
			SessionID:   sessionID,
			CurrentNode: "askPassword",
			Responses: map[string]string{
				"password": "testuser",
			},
		}

		resp := e.POST("/acme/customers/api/v1/username-password-register").
			WithHeader("Content-Type", "application/json").
			WithJSON(request).
			Expect().
			Status(http.StatusOK).
			JSON()

		// Validate success response
		resp.Object().
			HasValue("currentNode", "registerSuccess").
			Value("executionId").String().IsEqual(executionID)

		resp.Object().Value("sessionId").String().IsEqual(sessionID)

		// Check if result contains a sucessful response
		resp.Object().Value("result").IsObject().Object().
			HasValue("success", true)
	})

}

func TestJSONFlow_FlowWithoutApplication(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	t.Run("Flow with no application and with success result should result in a success=true response without tokens", func(t *testing.T) {

		e.GET("/acme/customers/api/v1/mock-success").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			Value("result").Object().
			HasValue("success", true).
			NotContainsKey("access_token")
	})
}

func TestJSONFlow_FlowWithFailureResult(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	t.Run("Flow with failure result should result in a success=false response without tokens", func(t *testing.T) {

		e.GET("/acme/customers/api/v1/mock-failure").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			Value("result").Object().
			HasValue("success", false).
			NotContainsKey("access_token")
	})
}
