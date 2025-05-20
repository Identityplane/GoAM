package auth

import (
	"fmt"
	"goiam/internal/auth/graph"
	"goiam/internal/auth/repository"
	"goiam/internal/logger"
	"goiam/internal/model"
	"goiam/internal/service"

	"github.com/valyala/fasthttp"
)

const sessionCookieName = "session_id"

type GraphHandler struct {
	Flow     *model.FlowDefinition
	Tenant   string
	Realm    string
	Services *repository.Repositories
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

	// Create a new or load the authentication session
	session, err := GetOrCreateAuthenticationSesssion(ctx, tenant, realm, loadedRealm.Config.BaseUrl, flow)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Could not create session")
		return
	}

	// Process the auth request
	newSession, err := ProcessAuthRequest(ctx, flow, *session)

	// If there is an error we render the error, otherwiese the ProcessAuthRequest will render the result
	if err != nil {
		RenderError(ctx, err.Error())
		return
	}

	// Save the updated state in the session
	service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(ctx, tenant, realm, *newSession)

	// If we are in an oauth2 flow we need to use the finish function to finish the oaith2 flow
	if newSession.Result != nil && newSession.Oauth2SessionInformation != nil {

		// We forward to the finish authorization endpoint
		url := fmt.Sprintf("%s/oauth2/finishauthorize", loadedRealm.Config.BaseUrl)
		ctx.Redirect(url, fasthttp.StatusSeeOther)
		return
	}

	if newSession.Result != nil {
		// If the result is set we clear the session
		service.GetServices().SessionsService.DeleteAuthenticationSession(ctx, tenant, realm, session.SessionIdHash)
	}

	// Render the result
	currentNode := flow.Definition.Nodes[newSession.Current]
	Render(ctx, flow.Definition, newSession, currentNode, newSession.Prompts, loadedRealm.Config.BaseUrl)
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
	if ctx.IsPost() {
		input = extractPromptsFromRequest(ctx, flow.Definition, session.Current)
	}

	// Run the flow engine with the current state and input
	newSession, err := graph.Run(flow.Definition, &session, input, registry)
	if err != nil {
		logger.DebugNoContext("flow resulted in error: %v", err)
		return nil, err
	}

	// This should be cleaned up in the future, its not beautiful to manually lookup the result node like this
	currentNode := flow.Definition.Nodes[session.Current]
	if currentNode == nil {
		return nil, fmt.Errorf("result node not found")
	}

	return newSession, nil
}

func GetAuthenticationSession(ctx *fasthttp.RequestCtx, tenant, realm string) (*model.AuthenticationSession, bool) {
	// Try load the session cookie from the request
	cookie := string(ctx.Request.Header.Cookie(sessionCookieName))

	// if present we load the session from the session service
	session, ok := service.GetServices().SessionsService.GetAuthenticationSessionByID(ctx, tenant, realm, cookie)
	return session, ok
}

func GetOrCreateAuthenticationSesssion(ctx *fasthttp.RequestCtx, tenant, realm, baseUrl string, flow *model.Flow) (*model.AuthenticationSession, error) {

	// Try to get existing session first
	session, ok := GetAuthenticationSession(ctx, tenant, realm)
	if !ok {
		return CreateNewAuthenticationSession(ctx, tenant, realm, baseUrl, flow)
	}

	// If the session if from a different flow we delete it and create a new one by overwriting it
	if session != nil && session.FlowId != flow.Id {
		service.GetServices().SessionsService.DeleteAuthenticationSession(ctx, tenant, realm, session.SessionIdHash)
		return CreateNewAuthenticationSession(ctx, tenant, realm, baseUrl, flow)
	}

	// If no session exists, create new one
	return session, nil
}

func CreateNewAuthenticationSession(ctx *fasthttp.RequestCtx, tenant, realm, baseUrl string, flow *model.Flow) (*model.AuthenticationSession, error) {

	// if not we create a new session
	loginUri := baseUrl + "/auth/" + flow.Route
	session, sessionID := service.GetServices().SessionsService.CreateAuthSessionObject(tenant, realm, flow.Id, loginUri)

	c := &fasthttp.Cookie{}
	c.SetPath("/")
	c.SetKey(sessionCookieName)
	c.SetValue(sessionID)
	c.SetHTTPOnly(true)
	ctx.Response.Header.SetCookie(c)

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
	logger.DebugNoContext("body: %s", body)

	// Check the definiton to see which inputs are allowed
	def := graph.NodeDefinitions[node.Use]
	for key := range def.PossiblePrompts {
		val := string(ctx.PostArgs().Peek(key))
		if val != "" {
			input[key] = val
		}
	}

	return input
}
