package oauth2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/Identityplane/GoAM/internal/auth/graph"
	"github.com/Identityplane/GoAM/internal/lib/oauth2"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/internal/web/auth"
	"github.com/Identityplane/GoAM/internal/web/webutils"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
	"github.com/valyala/fasthttp"
)

// HandleAuthorizeEndpoint handles the OAuth2 authorization endpoint
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
func HandleAuthorizeEndpoint(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flowId := string(ctx.QueryArgs().Peek("flow"))

	// Parse URL parameters
	oauth2request := &model.AuthorizeRequest{
		ClientID:            string(ctx.QueryArgs().Peek("client_id")),
		RedirectURI:         string(ctx.QueryArgs().Peek("redirect_uri")),
		ResponseType:        string(ctx.QueryArgs().Peek("response_type")),
		Scope:               strings.Split(string(ctx.QueryArgs().Peek("scope")), " "),
		State:               string(ctx.QueryArgs().Peek("state")),
		CodeChallenge:       string(ctx.QueryArgs().Peek("code_challenge")),
		CodeChallengeMethod: string(ctx.QueryArgs().Peek("code_challenge_method")),
		Nonce:               string(ctx.QueryArgs().Peek("nonce")),
		Prompt:              string(ctx.QueryArgs().Peek("prompt")),
	}

	var redirectUri string = ""

	// Load the client from the database
	application, ok := service.GetServices().ApplicationService.GetApplication(tenant, realm, oauth2request.ClientID)
	if !ok {
		RenderOauth2ErrorWithoutRedirect(ctx, oauth2.ErrorUnauthorizedClient, "Invalid client ID")
		return
	}

	if len(application.RedirectUris) == 0 {
		RenderOauth2ErrorWithoutRedirect(ctx, oauth2.ErrorInvalidRequest, "No redirect URI found for client")
		return
	}

	// Take the first redirect URI as the trusted redirect URI
	redirectUri = application.RedirectUris[0]

	// Check if the redirect URI is trusted
	if !slices.Contains(application.RedirectUris, oauth2request.RedirectURI) {
		RenderOauth2Error(ctx, oauth2.ErrorInvalidRequest, "Invalid redirect URI", oauth2request, redirectUri)
		return
	}

	// After we validated the redirect URI we can use it as the trusted redirect URI
	redirectUri = oauth2request.RedirectURI

	// If there are no allowed flows we return an error
	if len(application.AllowedAuthenticationFlows) == 0 {
		RenderOauth2Error(ctx, oauth2.ErrorInvalidRequest, "No allowed authentication flows", oauth2request, redirectUri)
		return
	}

	// If the flow id is not set as an additional paramater we default to the first flow of the allowed flows
	if flowId == "" {
		flowId = application.AllowedAuthenticationFlows[0]
	}

	// Load the realm
	loadedRealm, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		RenderOauth2Error(ctx, oauth2.ErrorInvalidRequest, "Realm not found: "+flowId, oauth2request, redirectUri)
		return
	}

	// Load the flow
	flow, ok := service.GetServices().FlowService.GetFlowById(tenant, realm, flowId)
	if !ok {
		RenderOauth2Error(ctx, oauth2.ErrorInvalidRequest, "Flow not found: "+flowId, oauth2request, redirectUri)
		return
	}

	// Load the flow for execution
	flow, ok = service.GetServices().FlowService.GetFlowForExecution(flow.Route, loadedRealm)
	if !ok {
		RenderOauth2Error(ctx, oauth2.ErrorInvalidRequest, "Cannot load flow for execution: "+flowId, oauth2request, redirectUri)
		return
	}

	// Validate and process the authorization request
	oauth2error := service.GetServices().OAuth2Service.ValidateOAuth2AuthorizationRequest(oauth2request, tenant, realm, application, flowId)
	if oauth2error != nil {
		RenderOauth2Error(ctx, oauth2error.Error, oauth2error.ErrorDescription, oauth2request, redirectUri)
		return
	}

	// We create a new session for this auth request
	var session *model.AuthenticationSession
	var err error
	session, err = auth.CreateNewAuthenticationSession(ctx, tenant, realm, loadedRealm.Config.BaseUrl, flow, false)
	if err != nil {
		RenderOauth2Error(ctx, oauth2.ErrorServerError, "Internal server error. Cannot create session", oauth2request, redirectUri)
		return
	}

	// We set the finish url of the auth session to the oauth2/finishauthorize endpoint
	baseUrl := loadedRealm.Config.BaseUrl
	if baseUrl == "" {
		baseUrl = webutils.GetFallbackUrl(ctx, tenant, realm)
	}
	session.FinishUri = fmt.Sprintf("%s/oauth2/finishauthorize", baseUrl)

	// Set the oauth2 context to the session
	session.Oauth2SessionInformation = &model.Oauth2Session{}
	session.Oauth2SessionInformation.AuthorizeRequest = oauth2request

	session, oauth2error = peekGraphExecutionForPromptParameter(session, flow, loadedRealm)

	if oauth2error != nil {
		RenderOauth2Error(ctx, oauth2error.Error, oauth2error.ErrorDescription, oauth2request, redirectUri)
		return
	}

	// If the resulting state is a result node we directly process the FinsishOauth2AuthorizationEndpoint
	if session.Result != nil {
		ctx.SetUserValue("session", session)
		FinsishOauth2AuthorizationEndpoint(ctx)
		return
	}

	// Otherwise we redirect to the login page where the user will be prompted

	// Save the session
	service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(ctx, tenant, realm, *session)

	// Redirect to the login page
	webutils.RedirectTo(ctx, session.LoginUri)
}

// This functions starts the graph execution and peeks if there is a prompt
// This is needed for the OIDC prompt parameter to check if the user is prompted or not
func peekGraphExecutionForPromptParameter(session *model.AuthenticationSession, flow *model.Flow, loadedRealm *services_interface.LoadedRealm) (*model.AuthenticationSession, *oauth2.OAuth2Error) {

	// Check if service registry is initialized
	registry := loadedRealm.Repositories
	if registry == nil {
		return nil, &oauth2.OAuth2Error{Error: oauth2.ErrorServerError, ErrorDescription: "Internal server error. Service registry not initialized"}
	}

	// if the flow has no definition we return an error
	if flow.Definition == nil {
		return nil, &oauth2.OAuth2Error{Error: oauth2.ErrorServerError, ErrorDescription: "Internal server error. Flow has no definition"}
	}

	// Run the flow engine with the current state and input
	newSession, err := graph.Run(flow.Definition, session, nil, registry)
	if err != nil {
		log := logger.GetGoamLogger()
		log.Debug().Err(err).Msg("flow resulted in error")
		return nil, &oauth2.OAuth2Error{Error: oauth2.ErrorServerError, ErrorDescription: "Internal server error. Flow resulted in error"}
	}

	// Check if there is prompt of a result
	asksForPrompts := (newSession.Prompts != nil)

	// If prompt parameter is set to none but there is prompt to the user we return an error
	if session.Oauth2SessionInformation.AuthorizeRequest.Prompt == "none" && asksForPrompts {
		return nil, &oauth2.OAuth2Error{Error: oauth2.ErrorLoginRequired, ErrorDescription: "Login required"}
	}

	if session.Oauth2SessionInformation.AuthorizeRequest.Prompt == "login" && !asksForPrompts {
		return nil, &oauth2.OAuth2Error{Error: oauth2.ErrorServerError, ErrorDescription: "No login required but prompt parameter is set to login"}
	}

	return newSession, nil
}

// FinishOauth2AuthorizationEndpoint finishes the OAuth2 authorization endpoint
// This endpoint is called by the login page after the flow has been completed

func FinsishOauth2AuthorizationEndpoint(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Load session from contex if available
	var session *model.AuthenticationSession
	sessionAny := ctx.UserValue("session")
	if sessionAny != nil {
		session = ctx.UserValue("session").(*model.AuthenticationSession)
	}

	// Otherwise load session from cookie
	if session == nil {
		// if the session is not already set we try to get it from the session service
		var ok bool
		session, ok = auth.GetAuthenticationSession(ctx, tenant, realm)

		// If no session is found we return an error as finish authrotize needs a session
		// as this should never happen we return it as internal error so that we can log it
		if !ok {
			RenderOauth2ErrorWithoutRedirect(ctx, oauth2.ErrorServerError, "Internal server error. No session")
			return
		}
	}

	// We can use it here as it was validated before in the authorize endpoint, otherwise no session would be created
	redirectUri := session.Oauth2SessionInformation.AuthorizeRequest.RedirectURI
	oauth2request := session.Oauth2SessionInformation.AuthorizeRequest

	// Get the authorization response
	response, oauth2error := service.GetServices().OAuth2Service.FinishOauth2AuthorizationEndpoint(session, tenant, realm)
	if oauth2error != nil {
		RenderOauth2Error(ctx, oauth2error.Error, oauth2error.ErrorDescription, oauth2request, redirectUri)
		return
	}

	// Build the redirect URL with the response parameters
	redirectURL := redirectUri + "?" + service.GetServices().OAuth2Service.ToQueryString(response)

	// Set the response headers
	ctx.SetStatusCode(fasthttp.StatusSeeOther)
	ctx.Response.Header.Set("Location", redirectURL)
	ctx.Response.Header.Set("Cache-Control", "no-store")
	ctx.Response.Header.Set("Pragma", "no-cache")
}

// HandleTokenEndpoint handles the OAuth2 token endpoint
// @Summary OAuth2 Token Endpoint
// @Description Handles the OAuth2 token request and returns a token response
// @Tags OAuth2
// @Accept x-www-form-urlencoded
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param code formData string false "Authorization code"
// @Param code_verifier formData string false "Code verifier"
// @Param client_id formData string true "Client ID"
// @Param client_secret formData string true "Client Secret"
// @Param grant_type formData string true "Grant Type"
// @Param refresh_token formData string false "Refresh Token"
// @Param scope formData string false "Scope"
// @Success 200 {object} service.Oauth2TokenResponse "Token response"
// @Failure 400 {string} string "Bad Request - Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/oauth2/token [post]
func HandleTokenEndpoint(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	tokenRequest := &oauth2.Oauth2TokenRequest{}

	// parse request parameter from application/x-www-form-urlencoded
	body := ctx.PostBody()
	bodyParams, err := url.ParseQuery(string(body))
	if err != nil {
		RenderOauth2Error(ctx, oauth2.ErrorInvalidRequest, "Invalid request body", nil, "")
		return
	}

	// parse the body parameters to the token request
	tokenRequest.Code = bodyParams.Get("code")
	tokenRequest.CodeVerifier = bodyParams.Get("code_verifier")
	tokenRequest.ClientID = bodyParams.Get("client_id")
	tokenRequest.GrantType = bodyParams.Get("grant_type")
	tokenRequest.RefreshToken = bodyParams.Get("refresh_token")
	tokenRequest.Scope = bodyParams.Get("scope")

	// Parse the client authentication
	clientAuthentication := getClientAuthenticationFromRequest(ctx)

	// If the client_id it not set in the body we take it from the client authentication
	if tokenRequest.ClientID == "" {
		tokenRequest.ClientID = clientAuthentication.ClientID
	}

	// Process the token request
	tokenResponse, oauthError := service.GetServices().OAuth2Service.ProcessTokenRequest(tenant, realm, tokenRequest, &clientAuthentication)
	if oauthError != nil {
		RenderOauth2ErrorWithoutRedirect(ctx, oauthError.Error, oauthError.ErrorDescription)
		return
	}

	// Set the response headers
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")

	// Set the body to the token response
	jsonData, err := json.MarshalIndent(tokenResponse, "", "  ")
	if err != nil {
		RenderOauth2ErrorWithoutRedirect(ctx, oauth2.ErrorServerError, "Internal server error. Cannot marshal token response")
		return
	}

	ctx.SetBody(jsonData)
}

func getClientAuthenticationFromRequest(ctx *fasthttp.RequestCtx) oauth2.Oauth2ClientAuthentication {

	clientAuthentication := oauth2.Oauth2ClientAuthentication{}

	// If we have an Authorization header with Basic we use it
	authorization := string(ctx.Request.Header.Peek("Authorization"))
	if strings.HasPrefix(authorization, "Basic ") {

		// Base 64 decode the client id and secret
		decoded, err := base64.StdEncoding.DecodeString(authorization[6:])

		if err == nil {
			basicAuth := strings.SplitN(string(decoded), ":", 2)

			if len(basicAuth) == 2 {
				clientAuthentication.ClientID = basicAuth[0]
				clientAuthentication.ClientSecret = basicAuth[1]
				return clientAuthentication
			}
		}
	}

	// parse request parameter from application/x-www-form-urlencoded
	body := ctx.PostBody()
	bodyParams, err := url.ParseQuery(string(body))
	if err != nil {
		return clientAuthentication
	}

	clientAuthentication.ClientID = bodyParams.Get("client_id")
	clientAuthentication.ClientSecret = bodyParams.Get("client_secret")

	return clientAuthentication
}

func RenderOauth2ErrorWithoutRedirect(ctx *fasthttp.RequestCtx, errorCode string, errorDescription string) {

	// Set the status code to 400 and the content type to json
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
	ctx.SetContentType("application/json")

	// Build the error response
	errorResponse := oauth2.OAuth2Error{
		Error:            errorCode,
		ErrorDescription: errorDescription,
	}

	// Marshal the error response
	jsonData, err := json.Marshal(errorResponse)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Internal server error. Cannot marshal error response")
		return
	}

	// Set the body to the error response
	ctx.SetBody(jsonData)
}

// RenderOauth2Error sends an OAuth2 error response as a redirect
func RenderOauth2Error(ctx *fasthttp.RequestCtx, errorCode string, errorDescription string, oauth2request *model.AuthorizeRequest, trustedRedirectURI string) {

	// Get the redirect URI from the parameter to extend with the error parameters
	redirectURI := trustedRedirectURI

	// Build query parameters using url.Values for proper encoding
	params := url.Values{}
	params.Add("error", errorCode)
	params.Add("error_description", errorDescription)
	if oauth2request.State != "" {
		params.Add("state", oauth2request.State)
	}

	// Build the redirect URL with encoded parameters
	redirectURL := redirectURI + "?" + params.Encode()

	// Set the response headers
	ctx.SetStatusCode(fasthttp.StatusSeeOther)
	ctx.Response.Header.Set("Location", redirectURL)
	ctx.Response.Header.Set("Cache-Control", "no-store")
	ctx.Response.Header.Set("Pragma", "no-cache")
}
