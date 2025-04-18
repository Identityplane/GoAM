package web

import (
	"goiam/internal/auth/graph"
	"goiam/internal/realms"
	"goiam/internal/web/session"
	"log"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const sessionCookieName = "session_id"

type GraphHandler struct {
	Flow     *graph.FlowDefinition
	Tenant   string
	Realm    string
	Services *graph.ServiceRegistry
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
	path := ctx.UserValue("path").(string)

	loadedRealm, ok := realms.GetRealm(tenant + "/" + realm)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString("realm not found")
		return
	}

	flow, err := realms.LookupFlow(tenant, realm, path)
	if err != nil {
		// return 404
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString("flow not found")
		return
	}

	// Check if service registry is initialized
	services := loadedRealm.Services
	if services == nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("service registry not initialized")
		return
	}

	handler := NewGraphHandler(tenant, realm, flow.Flow, services)

	// Execute the actual handler
	handler.Handle(ctx)
}

func NewGraphHandler(tenant string, realm string, flow *graph.FlowDefinition, services *graph.ServiceRegistry) *GraphHandler {

	// check if tenant, realm and flow are valid
	if tenant == "" || realm == "" || flow == nil {
		log.Fatalf("Invalid parameters: tenant=%s, realm=%s, flow=%v", tenant, realm, flow)
	}

	// load templates for rendering
	// currently we reload this with every request, this should be improved
	if err := InitTemplates(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	return &GraphHandler{Flow: flow, Tenant: tenant, Realm: realm, Services: services}
}

func (h *GraphHandler) Handle(ctx *fasthttp.RequestCtx) {
	sessionID := h.getOrCreateSessionID(ctx)

	var state *graph.FlowState

	if ctx.IsGet() {
		state = graph.InitFlow(h.Flow)
	} else {
		state = session.Load(sessionID)
		if state == nil {
			state = graph.InitFlow(h.Flow)
		}
	}

	var input map[string]string
	if ctx.IsPost() {
		step := string(ctx.PostArgs().Peek("step"))
		if step != "" && state.Current == step {
			input = extractPromptsFromRequest(ctx, h.Flow, step)
		}
	}

	//TODO Adapt to new interface
	// Run the flow engine with the current state and input
	state, err := graph.Run(h.Flow, state, input, h.Services)
	if err != nil {
		log.Printf("flow resulted in error: %v", err)
		RenderError(ctx, err.Error())
		return
	}

	prompts := state.Prompts
	session.Save(sessionID, state)

	// This should be cleaned up in the future, its not beautiful to manually lookup the result node like this
	var resultNode *graph.GraphNode
	if state.Result != nil {
		resultNode = h.Flow.Nodes[state.Current]
		if resultNode == nil {
			log.Printf("result node not found: %s", state.Current)
			RenderError(ctx, "Result node not found")
			return
		}
	}

	if err != nil {
		log.Printf("flow error: %v", err)
		RenderError(ctx, err.Error())
		return
	}

	Render(ctx, h.Flow, state, resultNode, prompts)
}

func (h *GraphHandler) getOrCreateSessionID(ctx *fasthttp.RequestCtx) string {
	cookie := ctx.Request.Header.Cookie(sessionCookieName)
	if len(cookie) > 0 {
		return string(cookie)
	}

	sessionID := uuid.New().String()
	c := &fasthttp.Cookie{}
	c.SetPath("/")
	c.SetKey(sessionCookieName)
	c.SetValue(sessionID)
	c.SetHTTPOnly(true)
	ctx.Response.Header.SetCookie(c)
	return sessionID
}

func extractPromptsFromRequest(ctx *fasthttp.RequestCtx, flow *graph.FlowDefinition, step string) map[string]string {
	input := make(map[string]string)

	node := flow.Nodes[step]
	if node == nil {
		return input
	}

	def := graph.NodeDefinitions[node.Use]
	for key := range def.Prompts {
		val := string(ctx.PostArgs().Peek(key))
		if val != "" {
			input[key] = val
		}
	}

	return input
}
