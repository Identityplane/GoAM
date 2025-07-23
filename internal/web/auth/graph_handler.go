package auth

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Identityplane/GoAM/internal/auth/graph"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/internal/web/webutils"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/valyala/fasthttp"
)

const sessionCookieName = "session_id"

type GraphHandler struct {
	Flow     *model.FlowDefinition
	Tenant   string
	Realm    string
	Services *model.Repositories
}

// HandleAuthRequest processes authentication requests and manages the authentication flow
// @Summary Process authentication request
// @Description Handles authentication requests by executing the specified flow. Returns either a prompt for user input or a final result. Supports debug mode for additional information.
// @Tags Authentication
// @Accept application/x-www-form-urlencoded
// @Produce text/html
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param path path string true "Flow path/name"
// @Param debug query boolean false "Enable debug mode"
// @Param step formData string false "Current step in the flow"
// @Param {prompt_key} formData string false "User input for the current step's prompts"
// @Success 200 {string} string "HTML response containing either a prompt form or the final result"
// @Failure 404 {string} string "Realm or flow not found"
// @Failure 500 {string} string "Internal server error"
// @Router /{tenant}/{realm}/auth/{path} [get]
// @Router /{tenant}/{realm}/auth/{path} [post]
func HandleAuthRequest(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flowPath := ctx.UserValue("path").(string)

	// Load the flow
	flow, ok := service.GetServices().FlowService.GetFlowByPath(tenant, realm, flowPath)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString("flow not found")
		return
	}

	loadedRealm, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString("realm not found")
		return
	}

	// Check if debug is in the query parameters
	debug := ctx.QueryArgs().Has("debug")

	// TODO check if debug is allowed

	// Create a new or load the authentication session
	session, err := GetOrCreateAuthenticationSesssion(ctx, tenant, realm, loadedRealm.Config.BaseUrl, flow, debug)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Could not create session")
		return
	}

	// Process the auth request
	newSession, err := ProcessAuthRequest(ctx, flow, *session)

	// If there is an error we render the error, otherwiese the ProcessAuthRequest will render the result
	if err != nil {
		RenderError(ctx, err.Error(), newSession, loadedRealm.Config.BaseUrl)
		return
	}

	// Save the updated state in the session
	service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(ctx, tenant, realm, *newSession)

	// If the result is set and finish uri is set we redirect to the finish uri
	// without deleting the session so the endpoint can finish the flow
	if newSession.Result != nil && newSession.FinishUri != "" {

		// We forward to the finish authorization endpoint
		webutils.RedirectTo(ctx, newSession.FinishUri)
		return
	}

	// If the result is set we clear the session
	if newSession.Result != nil {
		service.GetServices().SessionsService.DeleteAuthenticationSession(ctx, tenant, realm, session.SessionIdHash)
	}

	// Render the result
	currentNode := flow.Definition.Nodes[newSession.Current]

	// If the base url is empty we use the fallback url
	baseUrl := loadedRealm.Config.BaseUrl
	if baseUrl == "" {
		baseUrl = webutils.GetFallbackUrl(ctx, tenant, realm)
	}

	Render(ctx, flow.Definition, newSession, currentNode, newSession.Prompts, baseUrl)
}

func ProcessAuthRequest(ctx *fasthttp.RequestCtx, flow *model.Flow, session model.AuthenticationSession) (*model.AuthenticationSession, error) {

	tenant := flow.Tenant
	realm := flow.Realm

	// Load the realm
	loadedRealm, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		return nil, fmt.Errorf("realm not found")
	}

	// Check if service registry is initialized
	registry := loadedRealm.Repositories
	if registry == nil {
		return nil, fmt.Errorf("service registry not initialized")
	}

	// Check if flow definiton is available
	if flow.Definition == nil {
		return nil, fmt.Errorf("flow definiton not found")
	}

	// load templates for rendering
	// TODO currently we reload this with every request, this should be improved
	if err := InitTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates")
	}

	// Load the inputs from the request
	var input map[string]string
	input = extractPromptsFromRequest(ctx, flow.Definition, session.Current)

	// Run the flow engine with the current state and input
	newSession, err := graph.Run(flow.Definition, &session, input, registry)
	if err != nil {
		log := logger.GetLogger()
		log.Debug().Err(err).Msg("flow resulted in error")
		return newSession, err
	}

	// This should be cleaned up in the future, its not beautiful to manually lookup the result node like this
	currentNode := flow.Definition.Nodes[session.Current]
	if currentNode == nil {
		return newSession, fmt.Errorf("result node not found")
	}

	return newSession, nil
}

func GetAuthenticationSession(ctx *fasthttp.RequestCtx, tenant, realm string) (*model.AuthenticationSession, bool) {
	log := logger.GetLogger()

	// Try load the session cookie from the request
	cookie := string(ctx.Request.Header.Cookie(sessionCookieName))

	// if present we load the session from the session service
	session, ok := service.GetServices().SessionsService.GetAuthenticationSessionByID(ctx, tenant, realm, cookie)
	if !ok {
		log.Debug().Msg("session not found")
		return nil, false
	}
	return session, true
}

func GetOrCreateAuthenticationSesssion(ctx *fasthttp.RequestCtx, tenant, realm, baseUrl string, flow *model.Flow, debug bool) (*model.AuthenticationSession, error) {

	// Try to get existing session first
	session, ok := GetAuthenticationSession(ctx, tenant, realm)

	if !ok {
		return CreateNewAuthenticationSession(ctx, tenant, realm, baseUrl, flow, debug)
	}

	// If the session if from a different flow we delete it and create a new one by overwriting it
	if session != nil && session.FlowId != flow.Id {
		service.GetServices().SessionsService.DeleteAuthenticationSession(ctx, tenant, realm, session.SessionIdHash)
		return CreateNewAuthenticationSession(ctx, tenant, realm, baseUrl, flow, debug)
	}

	// if the session was not debug, but now we have debug, we need to set the debug flag
	if session != nil && !session.Debug && debug {
		session.Debug = true
	}

	// If no session exists, create new one
	return session, nil
}

func CreateNewAuthenticationSession(ctx *fasthttp.RequestCtx, tenant, realm, baseUrl string, flow *model.Flow, debug bool) (*model.AuthenticationSession, error) {
	log := logger.GetLogger()

	// If the base url is empty we use the fallback url
	if baseUrl == "" {
		baseUrl = webutils.GetFallbackUrl(ctx, tenant, realm)
	}

	// if not we create a new session
	loginUri := baseUrl + "/auth/" + flow.Route
	session, sessionID := service.GetServices().SessionsService.CreateAuthSessionObject(tenant, realm, flow.Id, loginUri)

	// Set the debug flag
	session.Debug = debug

	isHttps := strings.HasPrefix(baseUrl, "https://")

	// Parse base url and get path
	parsedUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %v", err)
	}
	basePath := parsedUrl.Path

	c := &fasthttp.Cookie{}
	c.SetPath(basePath)
	c.SetKey(sessionCookieName)
	c.SetValue(sessionID)
	c.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	c.SetHTTPOnly(true)
	if isHttps {
		c.SetSecure(true)
	}
	ctx.Response.Header.SetCookie(c)

	log.Debug().Msg("created new authentication session")

	return session, nil
}

func extractPromptsFromRequest(ctx *fasthttp.RequestCtx, flow *model.FlowDefinition, step string) map[string]string {
	input := make(map[string]string)

	// Lookup the node definiton
	node := flow.Nodes[step]
	if node == nil {
		return input
	}

	body := string(ctx.PostBody())
	log := logger.GetLogger()
	log.Debug().Str("body", string(body)).Msg("response body")

	// Check the definiton to see which inputs are allowed
	def := graph.NodeDefinitions[node.Use]
	for key := range def.PossiblePrompts {

		// read from query parameters (this is needed for example for oauth2 flows)
		val := string(ctx.QueryArgs().Peek(key))
		if val != "" {
			input[key] = val
		}

		// read from post body, overwrites any query parameter
		val = string(ctx.PostArgs().Peek(key))
		if val != "" {
			input[key] = val
		}
	}

	if len(input) == 0 {
		return nil
	}

	return input
}
