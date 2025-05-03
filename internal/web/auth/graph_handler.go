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

	// Process the auth request
	_, err := ProcessAuthRequest(ctx, flow)

	// If there is an error we render the error, otherwiese the ProcessAuthRequest will render the result
	if err != nil {
		RenderError(ctx, err.Error())
		return
	}
}

func ProcessAuthRequest(ctx *fasthttp.RequestCtx, flow *model.Flow) (*model.AuthenticationSession, error) {

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

	// Create a new or load the authentication session
	state, err := GetOrCreateAuthenticationSesssion(ctx, tenant, realm)
	if err != nil {

		return nil, fmt.Errorf("invalid session")
	}

	// Load the inputs from the request
	var input map[string]string
	if ctx.IsPost() {
		step := string(ctx.PostArgs().Peek("step"))
		if step != "" && state.Current == step {
			input = extractPromptsFromRequest(ctx, flow.Definition, step)
		}
	}

	// Run the flow engine with the current state and input
	state, err = graph.Run(flow.Definition, state, input, registry)
	if err != nil {
		logger.DebugNoContext("flow resulted in error: %v", err)
		return nil, err
	}

	// Save the updated state in the session
	service.GetServices().SessionsService.CreateOrUpdateAuthenticationSession(tenant, realm, *state)

	// This should be cleaned up in the future, its not beautiful to manually lookup the result node like this
	currentNode := flow.Definition.Nodes[state.Current]
	if currentNode == nil {
		return nil, fmt.Errorf("result node not found")
	}

	// If the result is set we clear the session
	if state.Result != nil {
		service.GetServices().SessionsService.DeleteAuthenticationSession(state.SessionIdHash)
	}

	// Render the result
	Render(ctx, flow.Definition, state, currentNode, state.Prompts)

	return state, nil
}

func GetOrCreateAuthenticationSesssion(ctx *fasthttp.RequestCtx, tenant, realm string) (*model.AuthenticationSession, error) {

	// Try load the session cookie from the request
	cookie := string(ctx.Request.Header.Cookie(sessionCookieName))

	// if present we load the session from the session service
	session, ok := service.GetServices().SessionsService.GetAuthenticationSessionByID(cookie)
	if ok {
		return session, nil
	}

	return CreateNewAuthenticationSession(ctx, tenant, realm)
}

func CreateNewAuthenticationSession(ctx *fasthttp.RequestCtx, tenant, realm string) (*model.AuthenticationSession, error) {

	// if not we create a new session
	session, sessionID := service.GetServices().SessionsService.CreateSessionObject(tenant, realm)

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
