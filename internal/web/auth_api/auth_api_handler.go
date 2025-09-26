package auth_api

import (
	"encoding/json"
	"fmt"

	"github.com/Identityplane/GoAM/internal/auth/graph"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/internal/web/auth"
	"github.com/Identityplane/GoAM/internal/web/webutils"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/valyala/fasthttp"
)

// JSON API Request/Response Structures

// FlowRequest represents a JSON API request for flow processing
type FlowRequest struct {
	SessionID   string            `json:"sessionId"`
	CurrentNode string            `json:"currentNode"`
	Responses   map[string]string `json:"responses"`
}

// FlowResponse represents a JSON API response for flow processing
type FlowResponse struct {
	RunId       string                    `json:"executionId,omitempty"`
	SessionID   string                    `json:"sessionId,omitempty"`
	CurrentNode string                    `json:"currentNode,omitempty"`
	Prompts     map[string]string         `json:"prompts,omitempty"`
	Result      *model.SimpleAuthResponse `json:"result,omitempty"`
	Error       *model.AuthError          `json:"error,omitempty"`
	Debug       any                       `json:"debug,omitempty"`
}

// FlowResult represents the final result of a successful flow
type FlowResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	UserID  string `json:"userId,omitempty"`
}

// FlowError represents an error in flow processing
type FlowError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// HandleJSONAuthRequest processes JSON authentication requests
func HandleJSONAuthRequest(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flowPath := ctx.UserValue("path").(string)

	// Set JSON content type
	ctx.SetContentType("application/json")

	// Load realm
	loadedRealm, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		sendErrorResponse(ctx, fasthttp.StatusNotFound, "REALM_NOT_FOUND", "Realm not found", "")
		return
	}

	// Load the flow
	flow, ok := service.GetServices().FlowService.GetFlowForExecution(flowPath, loadedRealm)
	if !ok {
		sendErrorResponse(ctx, fasthttp.StatusNotFound, "FLOW_NOT_FOUND", "Flow not found", "")
		return
	}

	// Handle GET request - start/continue flow
	if string(ctx.Method()) == "GET" {
		handleJSONGetRequest(ctx, loadedRealm.Config, flow)
		return
	}

	// Handle POST request - submit responses
	if string(ctx.Method()) == "POST" {
		handleJSONPostRequest(ctx, loadedRealm.Config, flow)
		return
	}

	// Method not allowed
	sendErrorResponse(ctx, fasthttp.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
}

// handleJSONGetRequest handles GET requests to start or continue a flow
func handleJSONGetRequest(ctx *fasthttp.RequestCtx, realm *model.Realm, flow *model.Flow) {

	queryArgs := ctx.QueryArgs()
	// Check if query contains a debug param (any value)
	debug := queryArgs.Has("debug")

	// Create new session for GET requests (starting a new flow)
	session, sessionId, err := createNewJSONSession(ctx, realm, flow, debug)
	if err != nil {
		sendErrorResponse(ctx, fasthttp.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Could not create session", "")
		return
	}

	// Get the client ID from the query parameters
	shouldReturn := initializeSimpleFlow(queryArgs, realm.Tenant, realm.Realm, ctx, flow, session)
	if shouldReturn {
		return
	}

	// Process the flow to get current state
	newSession, err := processJSONFlow(ctx, flow, *session)
	if err != nil {
		sendErrorResponse(ctx, fasthttp.StatusBadRequest, "FLOW_ERROR", err.Error(), "")
		return
	}

	// Save updated session
	service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(ctx, realm.Tenant, realm.Realm, *newSession)

	// Send response
	sendFlowResponse(ctx, newSession, flow, realm, sessionId)
}

func initializeSimpleFlow(queryArgs *fasthttp.Args, tenant string, realm string, ctx *fasthttp.RequestCtx, flow *model.Flow, session *model.AuthenticationSession) bool {

	err := auth.CreateSimpleAuthSession(ctx, flow, session, model.GRANT_SIMPLE_AUTH_BODY)
	if err != nil {
		sendErrorResponse(ctx, fasthttp.StatusBadRequest, err.Error, err.ErrorDescription, "")
		return true
	}

	return false
}

// handleJSONPostRequest handles POST requests to submit flow responses
func handleJSONPostRequest(ctx *fasthttp.RequestCtx, realm *model.Realm, flow *model.Flow) {
	// Parse request body
	var req FlowRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		sendErrorResponse(ctx, fasthttp.StatusBadRequest, "INVALID_JSON", "Invalid JSON request", "")
		return
	}

	// Get existing session using both IDs
	session, ok := getJSONSessionByIDs(ctx, realm.Tenant, realm.Realm, req.SessionID)
	if !ok {
		sendErrorResponse(ctx, fasthttp.StatusBadRequest, "INVALID_IDS", "Invalid session ID", "")
		return
	}

	// Validate current node matches
	if session.Current != req.CurrentNode {
		sendErrorResponse(ctx, fasthttp.StatusBadRequest, "INVALID_NODE", "Current node mismatch", "")
		return
	}

	// Process the flow with user responses
	newSession, err := processJSONFlowWithResponses(ctx, flow, *session, req.Responses)
	if err != nil {
		sendErrorResponse(ctx, fasthttp.StatusBadRequest, "FLOW_ERROR", err.Error(), "")
		return
	}

	// Save updated session
	service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(ctx, realm.Tenant, realm.Realm, *newSession)

	// Send response
	sendFlowResponse(ctx, newSession, flow, realm, req.SessionID)
}

// Helper functions
func createNewJSONSession(ctx *fasthttp.RequestCtx, realm *model.Realm, flow *model.Flow, debug bool) (*model.AuthenticationSession, string, error) {

	// Create new session
	realmUrl := webutils.GetUrlForRealm(ctx, realm)
	loginUri := realmUrl + "/api/v1/" + flow.Route

	session, sessionId := service.GetServices().SessionsService.CreateAuthSessionObject(realm.Tenant, realm.Realm, flow.Id, loginUri)

	// If allowed we add the debug flag
	if debug && flow.DebugAllowed {
		session.Debug = true
	}

	return session, sessionId, nil
}

func getJSONSessionByIDs(ctx *fasthttp.RequestCtx, tenant, realm, sessionID string) (*model.AuthenticationSession, bool) {
	// Get session by session ID
	sessionIDHash := lib.HashString(sessionID)
	session, ok := service.GetServices().SessionsService.GetAuthenticationSession(ctx, tenant, realm, sessionIDHash)
	if !ok {
		return nil, false
	}

	return session, true
}

func processJSONFlow(ctx *fasthttp.RequestCtx, flow *model.Flow, session model.AuthenticationSession) (*model.AuthenticationSession, error) {

	// Load realm
	loadedRealm, ok := service.GetServices().RealmService.GetRealm(flow.Tenant, flow.Realm)
	if !ok {
		return nil, fmt.Errorf("realm not found")
	}

	// Run flow engine without input (GET request)
	newSession, err := graph.Run(flow.Definition, &session, nil, loadedRealm.Repositories)
	if err != nil {
		return newSession, err
	}

	return newSession, nil
}

func processJSONFlowWithResponses(ctx *fasthttp.RequestCtx, flow *model.Flow, session model.AuthenticationSession, responses map[string]string) (*model.AuthenticationSession, error) {
	// Load realm
	loadedRealm, ok := service.GetServices().RealmService.GetRealm(flow.Tenant, flow.Realm)
	if !ok {
		return nil, fmt.Errorf("realm not found")
	}

	// Run flow engine with user responses
	newSession, err := graph.Run(flow.Definition, &session, responses, loadedRealm.Repositories)
	if err != nil {
		return newSession, err
	}

	return newSession, nil
}

func sendFlowResponse(ctx *fasthttp.RequestCtx, session *model.AuthenticationSession, flow *model.Flow, realm *model.Realm, sessionId string) {

	response := FlowResponse{
		RunId:       session.RunID,
		SessionID:   sessionId, // Sensitive session id
		CurrentNode: session.Current,
	}

	if session.Debug {
		response.Debug = session
	}

	// If there are prompts, add them
	if len(session.Prompts) > 0 {
		response.Prompts = session.Prompts
	}

	if session.Finished() {

		if session.SimpleAuthSessionInformation != nil {
			authResult, authError := auth.FinishSimpleAuthFlow(ctx, session, realm)
			if authError != nil {
				sendErrorResponse(ctx, fasthttp.StatusInternalServerError, authError.Error, authError.ErrorDescription, "")
				return
			}
			if authResult != nil && session.SimpleAuthSessionInformation.Request.Grant == model.GRANT_SIMPLE_AUTH_BODY {

				response.Result = authResult
			}
		} else {
			response.Result = &model.SimpleAuthResponse{
				Success: session.DidResultAuthenticated(),
			}
		}
	}

	// Send JSON response
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(response)
}

func sendErrorResponse(ctx *fasthttp.RequestCtx, statusCode int, code, message, field string) {

	ctx.SetStatusCode(statusCode)
	errorResp := FlowResponse{
		Error: &model.AuthError{
			Error:            code,
			ErrorDescription: message,
		},
	}
	json.NewEncoder(ctx).Encode(errorResp)
}
