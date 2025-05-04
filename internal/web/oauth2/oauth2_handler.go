package oauth2

import (
	"encoding/json"
	"fmt"
	"goiam/internal/lib/oauth2"
	"goiam/internal/model"
	"goiam/internal/service"
	"goiam/internal/web/auth"
	"net/url"
	"slices"
	"strings"

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
		RenderOauth2Error(ctx, oauth2.ErrorServerError, "Flow not found: "+flowId, oauth2request, redirectUri)
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
	session, err = auth.CreateNewAuthenticationSession(ctx, tenant, realm, loadedRealm.Config.BaseUrl, flow)
	if err != nil {
		RenderOauth2Error(ctx, oauth2.ErrorServerError, "Internal server error. Cannot create session", oauth2request, redirectUri)
		return
	}

	// Set the oauth2 context to the session
	session.Oauth2SessionInformation = &model.Oauth2Session{}
	session.Oauth2SessionInformation.AuthorizeRequest = oauth2request

	// Save the session
	service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(tenant, realm, *session)

	// For now we just redirect the user to the login page.
	// TODO later we should optimize this by calculting the first iteration of the graph directly
	// this allows us to peek if the user is prompted, complying with OIDC prompt paramter

	loginUrl := fmt.Sprintf("%s/auth/%s", loadedRealm.Config.BaseUrl, flow.Route)

	// If the debug paramter is set we add it to the login url
	if ctx.QueryArgs().Has("debug") {
		loginUrl += "?debug"
	}

	// Set the response headers
	ctx.SetStatusCode(fasthttp.StatusFound)
	ctx.Response.Header.Set("Location", loginUrl)
	ctx.Response.Header.Set("Cache-Control", "no-store")
	ctx.Response.Header.Set("Pragma", "no-cache")
}

func FinsishOauth2AuthorizationEndpoint(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	session, ok := auth.GetAuthenticationSession(ctx, tenant, realm)
	if !ok {
		RenderOauth2ErrorWithoutRedirect(ctx, oauth2.ErrorServerError, "Internal server error. No session")
		return
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
	ctx.SetStatusCode(fasthttp.StatusFound)
	ctx.Response.Header.Set("Location", redirectURL)
	ctx.Response.Header.Set("Cache-Control", "no-store")
	ctx.Response.Header.Set("Pragma", "no-cache")
}

func HandleTokenEndpoint(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// TODO
	service.GetServices().OAuth2Service.HandleTokenEndpoint(ctx, tenant, realm)
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
	ctx.SetStatusCode(fasthttp.StatusFound)
	ctx.Response.Header.Set("Location", redirectURL)
	ctx.Response.Header.Set("Cache-Control", "no-store")
	ctx.Response.Header.Set("Pragma", "no-cache")
}
