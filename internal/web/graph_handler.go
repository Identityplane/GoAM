package web

import (
	"goiam/internal/auth/graph"
	"goiam/internal/web/session"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const sessionCookieName = "session_id"

type GraphHandler struct {
	Flow *graph.FlowDefinition
}

func NewGraphHandler(flow *graph.FlowDefinition) *GraphHandler {
	return &GraphHandler{Flow: flow}
}

func (h *GraphHandler) Handle(ctx *fasthttp.RequestCtx) {
	sessionID := h.getOrCreateSessionID(ctx)

	var state *graph.FlowState

	if ctx.IsGet() {
		// Reset flow on GET
		state = graph.InitFlow(h.Flow)
	} else {
		state = session.Load(sessionID)
		if state == nil {
			state = graph.InitFlow(h.Flow)
		}
	}

	// Handle POSTed input if present
	var input map[string]string
	if ctx.IsPost() {
		step := string(ctx.PostArgs().Peek("step"))
		if step != "" && state.Current == step {
			// Build input map
			input = make(map[string]string)
			for key, node := range h.Flow.Nodes {
				if key == step {
					def := graph.NodeDefinitions[node.Use]
					for promptKey := range def.Prompts {
						inputVal := string(ctx.PostArgs().Peek(promptKey))
						if inputVal != "" {
							input[promptKey] = inputVal
						}
					}
					break
				}
			}
		}
	}

	// Run the flow
	prompts, resultNode, err := graph.Run(h.Flow, state, input)

	session.Save(sessionID, state)

	// Render appropriate response
	if err != nil {
		RenderError(ctx, *state.Error)
		return
	}

	if resultNode != nil {
		RenderResult(ctx, h.Flow, resultNode, state)
		return
	}

	if prompts != nil {
		RenderPrompts(ctx, h.Flow, state, prompts)
		return
	}

	RenderError(ctx, "Unknown flow state")
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
