package web

import (
	"goiam/internal/auth/graph"
	"goiam/internal/web/session"
	"log"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const sessionCookieName = "session_id"

type GraphHandler struct {
	Flow *graph.FlowDefinition
}

func NewGraphHandler(flow *graph.FlowDefinition) *GraphHandler {
	if err := InitTemplates(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
	return &GraphHandler{Flow: flow}
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

	prompts, resultNode, err := graph.Run(h.Flow, state, input)
	session.Save(sessionID, state)

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
