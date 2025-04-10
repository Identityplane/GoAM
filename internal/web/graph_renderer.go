package web

import (
	"encoding/json"
	"fmt"
	"goiam/internal/auth/graph"
	"strings"

	"github.com/valyala/fasthttp"
)

func RenderPrompts(ctx *fasthttp.RequestCtx, flow *graph.FlowDefinition, state *graph.FlowState, prompts map[string]string) {
	var b strings.Builder
	b.WriteString("<html><body>")
	b.WriteString(fmt.Sprintf("<h2>%s</h2>", strings.Title(state.Current)))

	b.WriteString(`<form method="POST">`)
	b.WriteString(fmt.Sprintf(`<input type="hidden" name="step" value="%s">`, state.Current))

	for key, typ := range prompts {
		b.WriteString(fmt.Sprintf(`<label>%s:</label><br>`, key))
		inputType := "text"
		if typ == "password" {
			inputType = "password"
		}
		b.WriteString(fmt.Sprintf(`<input name="%s" type="%s"><br><br>`, key, inputType))
	}

	b.WriteString(`<input type="submit" value="Submit">`)
	b.WriteString(`</form>`)

	if isDebugMode(ctx) {
		b.WriteString("<hr><h3>Debug State</h3><pre>")
		if js, err := json.MarshalIndent(state, "", "  "); err == nil {
			b.WriteString(string(js))
		}
		b.WriteString("</pre>")

		b.WriteString(`<svg width="1000" height="666">`)
		b.WriteString(`<image xlink:href="/debug/flow/graph.svg?flow=` + flow.Name + `" src="/debug/flow/graph.svg?flow=` + flow.Name + `" width="100%" height="100%"/>`)
		b.WriteString(`</svg>`)

	}

	b.WriteString("</body></html>")
	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(b.String())
}

func RenderResult(ctx *fasthttp.RequestCtx, flow *graph.FlowDefinition, node *graph.GraphNode, state *graph.FlowState) {
	var b strings.Builder
	b.WriteString("<html><body>")
	b.WriteString("<h2>Flow Completed</h2>")

	message := node.CustomConfig["message"]
	if message != "" {
		b.WriteString(fmt.Sprintf("<p>%s</p>", message))
	}

	if isDebugMode(ctx) {
		b.WriteString("<hr><h3>Debug State</h3><pre>")
		if js, err := json.MarshalIndent(state, "", "  "); err == nil {
			b.WriteString(string(js))
		}
		b.WriteString("</pre>")

		b.WriteString(`<svg width="1000" height="666">`)
		b.WriteString(`<image xlink:href="/debug/flow/graph.svg?flow=` + flow.Name + `" src="/debug/flow/graph.svg?flow=` + flow.Name + `" width="100%" height="100%"/>`)
		b.WriteString(`</svg>`)
	}

	b.WriteString("</body></html>")
	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(b.String())
}

func RenderError(ctx *fasthttp.RequestCtx, msg string) {
	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
	ctx.SetBodyString(fmt.Sprintf("<html><body><h2>Error</h2><p>%s</p></body></html>", msg))
}

func isDebugMode(ctx *fasthttp.RequestCtx) bool {
	return string(ctx.URI().QueryArgs().Peek("debug")) == "true"
}
