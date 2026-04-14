package auth_api

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	RunId             string                    `json:"executionId"`
	IsSecondaryDevice bool                      `json:"isSecondaryDevice"`
	SessionID         string                    `json:"sessionId,omitempty"`
	CurrentNode       string                    `json:"currentNode"`
	CurrentNodeType   string                    `json:"currentNodeType"`
	Prompts           map[string]string         `json:"prompts,omitempty"`
	Result            *model.SimpleAuthResponse `json:"result,omitempty"`
	Error             *model.AuthError          `json:"error,omitempty"`
	ErrorMessage      *string                   `json:"errorMessage,omitempty"`
	Debug             any                       `json:"debug,omitempty"`
	Flow              string                    `json:"flow,omitempty"`
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
		sendErrorResponse(ctx, nil, fasthttp.StatusNotFound, "REALM_NOT_FOUND", "Realm not found", "")
		return
	}

	// Load the flow
	flow, ok := service.GetServices().FlowService.GetFlowForExecution(flowPath, loadedRealm)
	if !ok {
		sendErrorResponse(ctx, nil, fasthttp.StatusNotFound, "FLOW_NOT_FOUND", "Flow not found", "")
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
	sendErrorResponse(ctx, nil, fasthttp.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
}

// handleJSONGetRequest handles GET requests to start or continue a flow
func handleJSONGetRequest(ctx *fasthttp.RequestCtx, realm *model.Realm, flow *model.Flow) {

	queryArgs := ctx.QueryArgs()
	// Check if query contains a debug param (any value)
	debug := queryArgs.Has("debug")

	var session *model.AuthenticationSession
	var sessionId string
	var err error

	isContinue := queryArgs.Has("continue")
	isInit := queryArgs.Has("init")
	isPoll := queryArgs.Has("poll")

	// Try to get existing session first from cookies
	session, ok := auth.GetAuthenticationSession(ctx, realm.Tenant, realm.Realm)

	// Determine if we should resume or create new.
	// poll=true may reuse the cookie even when Finished() is true: the secondary branch can
	// complete first (successResult) while the primary graph pointer still waits on qrMobileToWeb.
	shouldResume := ok && session != nil && session.FlowId == flow.Id && (!session.Finished() || isPoll)

	if isInit {
		// If init is requested, we force a new session
		if ok && session != nil {
			service.GetServices().SessionsService.DeleteAuthenticationSession(ctx, realm.Tenant, realm.Realm, session.SessionIdHash)
		}
		shouldResume = false
	}

	// If we should resume, we use the existing session
	if shouldResume {
		sessionId = session.PrimarySecretSessionID
	} else if isContinue {
		// If continue is set but no session found, return error
		sendErrorResponse(ctx, nil, fasthttp.StatusNotFound, "SESSION_NOT_FOUND", "No active session found to continue", "")
		return
	} else {
		// Otherwise create new session for GET requests (starting a new flow)
		session, sessionId, err = createNewJSONSession(ctx, realm, flow, debug)
		if err != nil {
			sendErrorResponse(ctx, nil, fasthttp.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Could not create session", "")
			return
		}

		// Set the session cookie for the new session
		cookie, authErr := auth.GetCookieForSessionId(ctx, sessionId, realm)
		if authErr == nil {
			ctx.Response.Header.SetCookie(cookie)
		}

		// Get the client ID from the query parameters
		shouldReturn := initializeSimpleFlow(queryArgs, realm.Tenant, realm.Realm, ctx, flow, session)
		if shouldReturn {
			return
		}
	}

	// Process the flow to get current state
	setHttpAuthContext(ctx, session)

	var newSession *model.AuthenticationSession
	// If we are resuming (shouldResume) or explicitly continuing (isContinue)
	// and we already have a current node, we just return the state as is.
	// This prevents the flow engine from running multiple times for the same state.
	// poll=true re-runs the engine so clients can observe cross-device context updates (e.g. QR wait).
	if (shouldResume || isContinue) && activeGraphNodeName(session) != "" && !isPoll {
		newSession = session
	} else {
		var err error
		newSession, err = processJSONFlow(ctx, flow, *session)
		if err != nil {
			sendErrorResponse(ctx, newSession, fasthttp.StatusBadRequest, "FLOW_ERROR", err.Error(), "")
			return
		}

		// Save updated session
		service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(ctx, realm.Tenant, realm.Realm, *newSession)
	}

	// Apply any response modifications (headers/cookies)
	auth.SetHttpAuthContextToResponse(newSession, ctx, realm)

	// Send response
	sendFlowResponse(ctx, newSession, flow, realm, sessionId)
}

func initializeSimpleFlow(queryArgs *fasthttp.Args, tenant string, realm string, ctx *fasthttp.RequestCtx, flow *model.Flow, session *model.AuthenticationSession) bool {

	err := auth.CreateSimpleAuthSession(ctx, flow, session, model.GRANT_SIMPLE_AUTH_BODY)
	if err != nil {
		sendErrorResponse(ctx, session, fasthttp.StatusBadRequest, err.Error, err.ErrorDescription, "")
		return true
	}

	return false
}

// HandleResumeSession resumes an existing session
func HandleResumeSession(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realmName := ctx.UserValue("realm").(string)
	evaluate := ctx.QueryArgs().Has("evaluate")

	// Set JSON content type
	ctx.SetContentType("application/json")

	// Parse request body for sessionId
	var req FlowRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		sendErrorResponse(ctx, nil, fasthttp.StatusBadRequest, "INVALID_JSON", "Invalid JSON request", "")
		return
	}

	if req.SessionID == "" {
		sendErrorResponse(ctx, nil, fasthttp.StatusBadRequest, "MISSING_SESSION_ID", "Session ID is required", "")
		return
	}

	// Load realm
	loadedRealm, ok := service.GetServices().RealmService.GetRealm(tenant, realmName)
	if !ok {
		sendErrorResponse(ctx, nil, fasthttp.StatusNotFound, "REALM_NOT_FOUND", "Realm not found", "")
		return
	}

	// Load session
	session, ok := getJSONSessionByIDs(ctx, tenant, realmName, req.SessionID)
	if !ok {
		sendErrorResponse(ctx, nil, fasthttp.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", "")
		return
	}

	// Find the flow. session.FlowId is the ID, not the route.
	// Must use GetFlowById so the flow definition YAML is loaded (ListFlows does not).
	foundFlow, ok := service.GetServices().FlowService.GetFlowById(tenant, realmName, session.FlowId)
	if !ok || foundFlow == nil {
		sendErrorResponse(ctx, session, fasthttp.StatusNotFound, "FLOW_NOT_FOUND", "Flow not found", "")
		return
	}

	responseSession := session
	// Resume with secondary secret: run engine so we land on the secondary branch (e.g. askUserID).
	lookupHash := lib.HashString(req.SessionID)
	secondaryResume := session.SecondarySessionIDHash != "" && lookupHash == session.SecondarySessionIDHash
	if secondaryResume {
		session.IsSecondaryDevice = true
	}

	if evaluate || secondaryResume {

		newSession, err := processJSONFlow(ctx, foundFlow, *session)
		if err != nil {
			sendErrorResponse(ctx, newSession, fasthttp.StatusBadRequest, "FLOW_ERROR", err.Error(), "")
			return
		}

		service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(ctx, tenant, realmName, *newSession)
		responseSession = newSession
	}

	// Apply any response modifications (headers/cookies)
	auth.SetHttpAuthContextToResponse(responseSession, ctx, loadedRealm.Config)

	// Send response
	sendFlowResponse(ctx, responseSession, foundFlow, loadedRealm.Config, req.SessionID)
}

// handleJSONPostRequest handles POST requests to submit flow responses
func handleJSONPostRequest(ctx *fasthttp.RequestCtx, realm *model.Realm, flow *model.Flow) {
	var req FlowRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		sendErrorResponse(ctx, nil, fasthttp.StatusBadRequest, "INVALID_JSON", "Invalid JSON request", "")
		return
	}

	session, ok := getJSONSessionByIDs(ctx, realm.Tenant, realm.Realm, req.SessionID)
	if !ok {
		sendErrorResponse(ctx, nil, fasthttp.StatusBadRequest, "INVALID_IDS", "Invalid session ID", "")
		return
	}

	// Validate current node matches (use active graph node: primary Current or secondary CurrentOnSecondaryDevice)
	if session.GetCurrent() != req.CurrentNode {
		sendErrorResponse(ctx, session, fasthttp.StatusBadRequest, "INVALID_NODE", "Current node mismatch", "")
		return
	}

	// Process the flow with user responses
	setHttpAuthContext(ctx, session)
	newSession, err := processJSONFlowWithResponses(ctx, flow, *session, req.Responses)
	if err != nil {
		sendErrorResponse(ctx, newSession, fasthttp.StatusBadRequest, "FLOW_ERROR", err.Error(), "")
		return
	}

	// Save updated session
	service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(ctx, realm.Tenant, realm.Realm, *newSession)

	// Apply any response modifications (headers/cookies)
	auth.SetHttpAuthContextToResponse(newSession, ctx, realm)

	// Send response
	sendFlowResponse(ctx, newSession, flow, realm, req.SessionID)
}

// Helper functions
func createNewJSONSession(ctx *fasthttp.RequestCtx, realm *model.Realm, flow *model.Flow, debug bool) (*model.AuthenticationSession, string, error) {

	// Create new session
	realmUrl := webutils.GetUrlForRealm(ctx, realm)
	loginUri := realmUrl + "/authui/" + flow.Route

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

	reconcileSecondaryDeviceFromLookup(session, sessionID)
	return session, true
}

// reconcileSecondaryDeviceFromLookup sets IsSecondaryDevice from which secret was used to load the session.
// json:"-" on IsSecondaryDevice and cache paths can leave the flag wrong; graph and API rely on it.
func reconcileSecondaryDeviceFromLookup(s *model.AuthenticationSession, rawSessionID string) {
	if s == nil || rawSessionID == "" {
		return
	}
	h := lib.HashString(rawSessionID)
	if s.SecondarySessionIDHash != "" && h == s.SecondarySessionIDHash {
		s.IsSecondaryDevice = true
		return
	}
	if s.SessionIdHash != "" && h == s.SessionIdHash {
		s.IsSecondaryDevice = false
	}
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
		RunId:             session.RunID,
		SessionID:         sessionId, // Sensitive session id
		CurrentNode:       activeGraphNodeName(session),
		CurrentNodeType:   session.CurrentType,
		Flow:              flow.Route,
		IsSecondaryDevice: session.IsSecondaryDevice,
	}

	if session.Debug {
		response.Debug = session
	}

	response.ErrorMessage = session.Error

	// If there are prompts, add them
	if len(session.Prompts) > 0 {
		response.Prompts = session.Prompts
	}

	if session.Finished() {

		if session.SimpleAuthSessionInformation != nil {
			authResult, authError := auth.FinishSimpleAuthFlow(ctx, session, realm)
			if authError != nil {
				sendErrorResponse(ctx, session, fasthttp.StatusInternalServerError, authError.Error, authError.ErrorDescription, "")
				return
			}
			if authResult != nil && session.SimpleAuthSessionInformation.Request.Grant == model.GRANT_SIMPLE_AUTH_BODY {
				response.Result = authResult
			}
		} else {
			result := &model.SimpleAuthResponse{
				Success:  session.DidResultAuthenticated(),
				Redirect: session.FinishUri,
			}
			if session.Result != nil {
				result.UserID = session.Result.UserID
			}
			response.Result = result
		}
	}

	// Always ensure debug info is included if session is in debug mode
	if session.Debug {
		response.Debug = session
	}

	// Send JSON response
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(response)
}

// activeGraphNodeName returns the node name the client should send as currentNode (primary Current or secondary branch).
func activeGraphNodeName(s *model.AuthenticationSession) string {
	if s == nil {
		return ""
	}
	if n := s.GetCurrent(); n != "" {
		return n
	}
	return s.Current
}

func sendErrorResponse(ctx *fasthttp.RequestCtx, session *model.AuthenticationSession, statusCode int, code, message, field string) {

	ctx.SetStatusCode(statusCode)
	errorResp := FlowResponse{
		Error: &model.AuthError{
			Error:            code,
			ErrorDescription: message,
		},
	}

	if session != nil {
		errorResp.RunId = session.RunID
		errorResp.CurrentNode = activeGraphNodeName(session)
		if session.Debug {
			errorResp.Debug = session
		}
	}

	json.NewEncoder(ctx).Encode(errorResp)
}

func setHttpAuthContext(ctx *fasthttp.RequestCtx, session *model.AuthenticationSession) {
	if session.HttpAuthContext == nil {
		session.HttpAuthContext = &model.HttpAuthContext{
			RequestIP:                 ctx.RemoteIP().String(),
			RequestHeaders:            webutils.GetRequestHeaders(ctx),
			RequestCookies:            webutils.GetRequestCookies(ctx),
			AdditionalResponseHeaders: make(map[string]string),
			AdditionalResponseCookies: make(map[string]http.Cookie),
		}
	} else {
		// Update IP each time
		session.HttpAuthContext.RequestIP = ctx.RemoteIP().String()
		// We might want to refresh headers/cookies here too if they changed
		session.HttpAuthContext.RequestHeaders = webutils.GetRequestHeaders(ctx)
		session.HttpAuthContext.RequestCookies = webutils.GetRequestCookies(ctx)

		// Ensure response maps are initialized if they were somehow serialized as nil
		if session.HttpAuthContext.AdditionalResponseHeaders == nil {
			session.HttpAuthContext.AdditionalResponseHeaders = make(map[string]string)
		}
		if session.HttpAuthContext.AdditionalResponseCookies == nil {
			session.HttpAuthContext.AdditionalResponseCookies = make(map[string]http.Cookie)
		}
	}
}
