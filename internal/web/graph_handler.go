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
	Flow   *graph.FlowDefinition
	Tenant string
	Realm  string
}

func HandleAuthRequest(ctx *fasthttp.RequestCtx) {

	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	path := ctx.UserValue("path").(string)

	flow, err := realms.LookupFlow(tenant, realm, path)
	if err != nil {
		// return 404
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString("flow not found")
		return
	}

	handler := NewGraphHandler(tenant, realm, flow.Flow)

	// Execute the actual handler
	handler.Handle(ctx)
}

func NewGraphHandler(tenant string, realm string, flow *graph.FlowDefinition) *GraphHandler {

	// check if tenant, realm and flow are valid
	if tenant == "" || realm == "" || flow == nil {
		log.Fatalf("Invalid parameters: tenant=%s, realm=%s, flow=%v", tenant, realm, flow)
	}

	// load templates for rendering
	// currently we reload this with every request, this should be improved
	if err := InitTemplates(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	return &GraphHandler{Flow: flow, Tenant: tenant, Realm: realm}
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
	state, err := graph.Run(h.Flow, state, input)
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
