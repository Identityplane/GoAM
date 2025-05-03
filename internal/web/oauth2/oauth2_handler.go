package oauth2

import (
	"context"
	"encoding/json"
	"goiam/internal/auth/graph"
	"goiam/internal/lib"
	"goiam/internal/model"
	"goiam/internal/service"
	"goiam/internal/web/auth"
	"net/url"
	"slices"
	"strings"

	"github.com/valyala/fasthttp"
)

// AuthorizationResponse represents the OAuth2 authorization response
type AuthorizationResponse struct {
	Code  string `json:"code"`  // REQUIRED. The authorization code
	State string `json:"state"` // REQUIRED if state was present in the request
	Iss   string `json:"iss"`   // OPTIONAL. The identifier of the authorization server
}

// OAuth2Error represents an OAuth2 error response
type OAuth2Error struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

// Valid OAuth2 error codes as defined in RFC 6749
const (
	ErrorInvalidRequest          = "invalid_request"           // RFC
	ErrorUnauthorizedClient      = "unauthorized_client"       // RFC
	ErrorAccessDenies            = "access_denied"             // RFC
	ErrorUnsupportedResponseType = "unsupported_response_type" // RFC
	ErrorInvalidScope            = "invalid_scope"             // RFC
	ErrorServerError             = "server_error"              // Not in rfc
	ErrorTemporarilyUnavailable  = "temporarily_unavailable"   // RFC
)

// HandleAuthorize handles the OAuth2 authorization endpoint
// @Summary OAuth2 Authorization Endpoint
// @Description Handles the OAuth2 authorization request and redirects to the client's redirect URI
// @Tags OAuth2
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param redirect_uri query string true "Redirect URI"
// @Param response_type query string true "Response Type"
// @Param scope query string true "Scope"
// @Param state query string true "State"
// @Param code_challenge query string true "Code Challenge"
// @Param code_challenge_method query string true "Code Challenge Method"
// @Success 302 {string} string "Redirect to client's redirect URI"
// @Failure 400 {string} string "Invalid request parameters"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/oauth2/authorize [get]
func HandleAuthorize(ctx *fasthttp.RequestCtx) {

	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flowId := string(ctx.QueryArgs().Peek("flow_id"))

	// Parse URL parameters
	oauth2request := model.AuthorizeRequest{
		ClientID:            string(ctx.QueryArgs().Peek("client_id")),
		RedirectURI:         string(ctx.QueryArgs().Peek("redirect_uri")),
		ResponseType:        string(ctx.QueryArgs().Peek("response_type")),
		Scope:               strings.Split(string(ctx.QueryArgs().Peek("scope")), " "),
		State:               string(ctx.QueryArgs().Peek("state")),
		CodeChallenge:       string(ctx.QueryArgs().Peek("code_challenge")),
		CodeChallengeMethod: string(ctx.QueryArgs().Peek("code_challenge_method")),
	}

	// Load the client from the database
	application, ok := service.GetServices().ApplicationService.GetApplication(tenant, realm, oauth2request.ClientID)
	if !ok {
		ReturnOAuth2Error(ctx, fasthttp.StatusNotFound, ErrorUnauthorizedClient, "Invalid client ID")
		return
	}

	// Check which flow is requested, we differenciate between authorization_code and authorization_code_pkce, and client_credentials
	// If we have a code challenge and grant type code it is a pkce flow
	var oauth2_flow lib.OAuth2GrantType = lib.Oauth2_InvalidFlow
	if oauth2request.CodeChallenge != "" && oauth2request.ResponseType == "code" {

		oauth2_flow = lib.Oauth2_AuthorizationCodePKCE
	} else if oauth2request.ResponseType == "code" {

		oauth2_flow = lib.Oauth2_AuthorizationCode
	} else {
		// Return invalid flow
		ReturnOAuth2Error(ctx, fasthttp.StatusBadRequest, ErrorUnsupportedResponseType, "Unsupported response type, authroization endpoint only supports code grant")
		return
	}

	// Check if the grant type is allowed
	if !slices.Contains(application.AllowedGrants, string(oauth2_flow)) {
		ReturnOAuth2Error(ctx, fasthttp.StatusBadRequest, ErrorUnauthorizedClient, "Grant type not allowed")
		return
	}

	// if the application is public is must have a code_challenge and code_challenge_method for the pkce flow
	if !application.Confidential && (oauth2request.CodeChallenge == "" || oauth2request.CodeChallengeMethod == "") {
		ReturnOAuth2Error(ctx, fasthttp.StatusBadRequest, ErrorInvalidRequest, "Code challenge and code challenge method are required for public applications")
		return
	}

	// validate if the redirect_uir is in the list of allowed redirect uris
	if oauth2request.RedirectURI != "" && !slices.Contains(application.RedirectUris, oauth2request.RedirectURI) {
		ReturnOAuth2Error(ctx, fasthttp.StatusBadRequest, ErrorInvalidRequest, "Invalid redirect URI")
		return
	}

	// Check if all requested scopes are allowed for each request scope
	for _, scope := range oauth2request.Scope {
		if !slices.Contains(application.AllowedScopes, scope) {
			ReturnOAuth2Error(ctx, fasthttp.StatusBadRequest, ErrorInvalidScope, "Invalid scope "+scope)
			return
		}
	}

	// If CodeChallengeMethod is provided it must be S256
	if oauth2request.CodeChallengeMethod != "" && oauth2request.CodeChallengeMethod != "S256" {
		ReturnOAuth2Error(ctx, fasthttp.StatusBadRequest, ErrorInvalidRequest, "Code challenge method must be S256")
		return
	}

	// If there is no allowed authentication flow we fail with a server error
	if len(application.AllowedAuthenticationFlows) == 0 {
		ReturnOAuth2Error(ctx, fasthttp.StatusInternalServerError, ErrorServerError, "Internal server error. No allowed authentication flows")
		return
	}

	// If the flow id is not set as an additional paramater we default to the first flow of the allowed flows
	if flowId == "" {
		flowId = application.AllowedAuthenticationFlows[0]
	}

	// Check if the flow is allowed for the application
	if !slices.Contains(application.AllowedAuthenticationFlows, flowId) {
		ReturnOAuth2Error(ctx, fasthttp.StatusBadRequest, ErrorUnauthorizedClient, "Flow not allowed")
		return
	}

	// Load the flow
	flow, ok := service.GetServices().FlowService.GetFlowById(tenant, realm, flowId)
	if !ok {
		ReturnOAuth2Error(ctx, fasthttp.StatusInternalServerError, ErrorServerError, "Internal server error. Flow not found: "+flowId)
		return
	}

	// We create a new session for this auth request
	var session *model.AuthenticationSession
	var err error
	session, err = auth.CreateNewAuthenticationSession(ctx, tenant, realm)
	if err != nil {
		ReturnOAuth2Error(ctx, fasthttp.StatusInternalServerError, ErrorServerError, "Internal server error. Cannot create session")
		return
	}

	// Set the oauth2 context to the session
	session.Oauth2SessionInformation.AuthorizeRequest = oauth2request

	// All checks passed, we can now start the authentication flow
	session, err = auth.ProcessAuthRequest(ctx, flow)

	if err != nil {
		ReturnOAuth2Error(ctx, fasthttp.StatusInternalServerError, ErrorServerError, err.Error())
		return
	}

	// If there is no result yet the ProcessAuthRequest rendered a resonse and we can return here
	if session.Result == nil {
		return
	}

	// If the result node is a failure result we return an oauth2 error
	if session.Current == graph.FailureResultNode.Name {
		ReturnOAuth2Error(ctx, fasthttp.StatusUnauthorized, ErrorAccessDenies, "Authentication Failed")
		return
	}

	// If the result node is a success result we redirect to the redirect uri
	if session.Current != graph.SuccessResultNode.Name {
		ReturnOAuth2Error(ctx, fasthttp.StatusInternalServerError, ErrorServerError, "Internal server error. Unexpected result node")
	}

	// In order to set a access token the result auth level must be at least 1
	if session.Result.AuthLevel == model.AuthLevelUnauthenticated {
		ReturnOAuth2Error(ctx, fasthttp.StatusUnauthorized, ErrorAccessDenies, "Authentication level unauthenticated")
		return
	}

	// If all ok we create a client session and issue an auth code
	authCode, err := service.GetServices().SessionsService.CreateAuthCodeSession(context.Background(), tenant, realm, application.ClientId, session.Result.UserID, oauth2request.Scope, "authorization_code")
	if err != nil {
		ReturnOAuth2Error(ctx, fasthttp.StatusUnauthorized, ErrorServerError, "Internal server error. Could not create session")
		return
	}

	// Create the authorization response
	response := AuthorizationResponse{
		Code:  authCode,
		State: oauth2request.State,
		Iss:   "nil",
	}

	// Build the redirect URL with the response parameters
	redirectURL := oauth2request.RedirectURI + "?" + response.ToQueryString()

	// Set the response headers
	ctx.SetStatusCode(fasthttp.StatusFound)
	ctx.Response.Header.Set("Location", redirectURL)
	ctx.Response.Header.Set("Cache-Control", "no-store")
	ctx.Response.Header.Set("Pragma", "no-cache")
}

// ToQueryString converts the AuthorizationResponse to a URL query string
func (r *AuthorizationResponse) ToQueryString() string {
	params := url.Values{}
	params.Add("code", r.Code)
	if r.State != "" {
		params.Add("state", r.State)
	}
	if r.Iss != "" {
		params.Add("iss", r.Iss)
	}
	return params.Encode()
}

// ReturnOAuth2Error sends an OAuth2 error response
func ReturnOAuth2Error(ctx *fasthttp.RequestCtx, statusCode int, errorCode string, errorDescription string) {
	errorResponse := OAuth2Error{
		Error:            errorCode,
		ErrorDescription: errorDescription,
	}

	// Set response headers
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	ctx.Response.Header.Set("Cache-Control", "no-store")

	// Marshal and send error response
	jsonData, err := json.Marshal(errorResponse)
	if err != nil {
		// If we can't marshal the error, send a basic error
		ctx.SetBodyString(`{"error":"server_error"}`)
		return
	}

	ctx.SetBody(jsonData)
}
