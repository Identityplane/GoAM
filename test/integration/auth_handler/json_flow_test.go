package authhandler

import (
	"net/http"
	"testing"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/test/integration"
	"github.com/google/uuid"
)

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

func TestJSONFlow_InvalidPassword(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	tenant := "acme"
	realm := "customers"
	username := "testuser-" + uuid.NewString()
	password := "correct-password"

	// Step 1: Register a user first so we can try to login
	t.Run("Register user", func(t *testing.T) {
		// Start registration flow to get session ID
		resp := e.GET("/" + tenant + "/" + realm + "/api/v1/username-password-register").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON()

		sessionID := resp.Object().Value("sessionId").String().Raw()

		// Submit first step (username)
		request := FlowRequest{
			SessionID:   sessionID,
			CurrentNode: "askUsername",
			Responses: map[string]string{
				"username": username,
			},
		}

		e.POST("/" + tenant + "/" + realm + "/api/v1/username-password-register").
			WithHeader("Content-Type", "application/json").
			WithJSON(request).
			Expect().
			Status(http.StatusOK)

		// Submit second step (password)
		request = FlowRequest{
			SessionID:   sessionID,
			CurrentNode: "askPassword",
			Responses: map[string]string{
				"password": password,
			},
		}

		e.POST("/" + tenant + "/" + realm + "/api/v1/username-password-register").
			WithHeader("Content-Type", "application/json").
			WithJSON(request).
			Expect().
			Status(http.StatusOK)
	})

	// Step 2: Try to login with wrong password
	t.Run("Login with wrong password", func(t *testing.T) {
		// Start login flow
		resp := e.GET("/" + tenant + "/" + realm + "/api/v1/login").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON()

		sessionID := resp.Object().Value("sessionId").String().Raw()

		// Submit username to get to password prompt
		resp = e.POST("/" + tenant + "/" + realm + "/api/v1/login").
			WithHeader("Content-Type", "application/json").
			WithJSON(FlowRequest{
				SessionID:   sessionID,
				CurrentNode: "askUsername",
				Responses: map[string]string{
					"username": username,
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON()

		resp.Object().HasValue("currentNode", "askPassword")

		// Submit wrong password
		request := FlowRequest{
			SessionID:   sessionID,
			CurrentNode: "askPassword",
			Responses: map[string]string{
				"password": "wrong-password",
			},
		}

		resp = e.POST("/" + tenant + "/" + realm + "/api/v1/login").
			WithHeader("Content-Type", "application/json").
			WithJSON(request).
			Expect().
			Status(http.StatusOK).
			JSON()

		// Should still be at the same node but with an error message
		resp.Object().HasValue("currentNode", "askPassword")
		resp.Object().Value("errorMessage").String().Contains("Invalid")
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

func TestJSONFlow_SessionResumption(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")
	flowRoute := "username-password-register"

	t.Run("init with no existing session", func(t *testing.T) {
		resp := e.GET("/acme/customers/api/v1/" + flowRoute).
			WithQuery("init", "true").
			Expect().
			Status(http.StatusOK)

		resp.JSON().Object().Value("sessionId").String().NotEmpty()

		// Verify cookie is set
		resp.Cookie("session_id").Value().NotEmpty()
	})

	t.Run("init with existing sessions", func(t *testing.T) {
		// 1. Create initial session
		resp1 := e.GET("/acme/customers/api/v1/" + flowRoute).
			Expect().
			Status(http.StatusOK)

		sessionID1 := resp1.JSON().Object().Value("sessionId").String().Raw()
		cookie1 := resp1.Cookie("session_id").Value().Raw()

		// 2. Call with init=true and existing cookie
		resp2 := e.GET("/acme/customers/api/v1/" + flowRoute).
			WithQuery("init", "true").
			WithCookie("session_id", cookie1).
			Expect().
			Status(http.StatusOK)

		sessionID2 := resp2.JSON().Object().Value("sessionId").String().Raw()
		if sessionID1 == sessionID2 {
			t.Errorf("Expected different session ID after init=true, got same: %s", sessionID1)
		}

		// The new cookie should also be different
		cookie2 := resp2.Cookie("session_id").Value().Raw()
		if cookie1 == cookie2 {
			t.Errorf("Expected different cookie value after init=true")
		}
	})

	t.Run("continue with no existing session", func(t *testing.T) {
		e.GET("/acme/customers/api/v1/" + flowRoute).
			WithQuery("continue", "true").
			Expect().
			Status(http.StatusNotFound).
			JSON().Object().Value("error").Object().Value("error").IsEqual("SESSION_NOT_FOUND")
	})

	t.Run("continue with existing session", func(t *testing.T) {
		// 1. Create initial session
		resp1 := e.GET("/acme/customers/api/v1/" + flowRoute).
			Expect().
			Status(http.StatusOK)

		sessionID1 := resp1.JSON().Object().Value("sessionId").String().Raw()
		cookie1 := resp1.Cookie("session_id").Value().Raw()

		// 2. Call with continue=true and existing cookie
		resp2 := e.GET("/acme/customers/api/v1/" + flowRoute).
			WithQuery("continue", "true").
			WithCookie("session_id", cookie1).
			Expect().
			Status(http.StatusOK)

		sessionID2 := resp2.JSON().Object().Value("sessionId").String().Raw()
		if sessionID1 != sessionID2 {
			t.Errorf("Expected same session ID after continue=true, got different: %s vs %s", sessionID1, sessionID2)
		}
	})
}

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
	Error       *model.AuthError  `json:"error,omitempty"`
	ErrorMessage *string          `json:"errorMessage,omitempty"`
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
