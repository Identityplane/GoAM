package web

import (
	"encoding/json"
	"fmt"
	"goiam/internal/auth/flows"
	"goiam/internal/web/session"
	"html/template"
	"strings"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const sessionCookieName = "session_id"

type HttpFlowRunner struct {
	Flow flows.Flow
}

func NewHttpFlowRunner(flow flows.Flow) *HttpFlowRunner {
	return &HttpFlowRunner{Flow: flow}
}

func (r *HttpFlowRunner) Handle(ctx *fasthttp.RequestCtx) {
	sessionID := r.getOrCreateSessionID(ctx)

	var state *flows.FlowState

	if ctx.IsGet() {
		// Always reset flow on GET
		state = &flows.FlowState{
			RunID: uuid.NewString(),
			Steps: []flows.FlowStep{{Name: "init"}},
		}
	} else {
		// POST — restore flow state from session
		state = session.Load(sessionID)
		if state == nil {
			state = &flows.FlowState{
				RunID: uuid.NewString(),
				Steps: []flows.FlowStep{{Name: "init"}},
			}
		} else {
			r.appendStepFromRequest(state, ctx)
		}
	}

	r.Flow.Run(state)
	session.Save(sessionID, state)

	if state.Error != nil {
		r.renderError(ctx, *state.Error)
		return
	}

	if state.Result != nil {
		r.renderResult(ctx, state.Result, state)
		return
	}

	step := flows.LastStep(state)
	if step == nil || len(step.Parameters) == 0 {
		r.renderError(ctx, "Invalid or empty flow step")
		return
	}

	r.renderForm(ctx, step, state)
}

func (r *HttpFlowRunner) getOrCreateSessionID(ctx *fasthttp.RequestCtx) string {
	cookie := ctx.Request.Header.Cookie(sessionCookieName)
	if len(cookie) > 0 {
		return string(cookie)
	}

	// Generate a new session ID
	sessionID := uuid.New().String()
	sessionCookie := &fasthttp.Cookie{}
	sessionCookie.SetPath("/")
	sessionCookie.SetKey(sessionCookieName)
	sessionCookie.SetValue(sessionID)
	sessionCookie.SetHTTPOnly(true)
	sessionCookie.SetSecure(false)
	sessionCookie.SetSameSite(fasthttp.CookieSameSiteDisabled)
	ctx.Response.Header.SetCookie(sessionCookie)

	return sessionID
}

func (r *HttpFlowRunner) appendStepFromRequest(state *flows.FlowState, ctx *fasthttp.RequestCtx) {

	// Parse submitted step name
	stepName := string(ctx.PostArgs().Peek("step"))
	if stepName == "" {
		return // silently ignore invalid POST
	}

	// Get the last step in the state
	lastStep := flows.LastStep(state)
	if lastStep == nil || lastStep.Name != stepName {
		// Prevent replay/step injection
		msg := fmt.Sprintf("Invalid step submission: expected '%s', got '%s'", lastStep.Name, stepName)
		state.Error = &msg
		return
	}

	// Update each expected parameter in the last step using values from the POST request
	for key := range lastStep.Parameters {
		val := ctx.PostArgs().Peek(key)
		lastStep.Parameters[key] = string(val)
	}
}

func (r *HttpFlowRunner) renderForm(ctx *fasthttp.RequestCtx, step *flows.FlowStep, state *flows.FlowState) {
	var b strings.Builder
	b.WriteString("<html><body>")
	b.WriteString(fmt.Sprintf("<h2>%s</h2>", strings.Title(step.Name)))
	b.WriteString(`<form method="POST">`)
	b.WriteString(fmt.Sprintf(`<input type="hidden" name="step" value="%s">`, step.Name))

	for field := range step.Parameters {
		b.WriteString(fmt.Sprintf(`<label>%s:</label><br>`, field))
		b.WriteString(fmt.Sprintf(`<input name="%s" type="text"><br><br>`, field))
	}

	b.WriteString(`<input type="submit" value="Submit">`)
	b.WriteString(`</form>`)

	// Debug JSON output
	if isDebugMode(ctx) {
		b.WriteString("<hr><h3>Debug Flow State</h3><pre>")
		jsonState, err := marshalStatePretty(state) // We'll create this helper next
		if err != nil {
			b.WriteString(fmt.Sprintf("Failed to serialize flow state: %v", err))
		} else {
			b.WriteString(jsonState)
		}
		b.WriteString("</pre>")
	}

	b.WriteString("</body></html>")

	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(b.String())
}

func (r *HttpFlowRunner) renderError(ctx *fasthttp.RequestCtx, msg string) {
	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
	ctx.SetBodyString(fmt.Sprintf("<html><body><h2>Error</h2><p>%s</p></body></html>", msg))
}

func (r *HttpFlowRunner) renderResult(ctx *fasthttp.RequestCtx, res *flows.FlowResult, state *flows.FlowState) {
	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusOK)

	var b strings.Builder

	b.WriteString(`<html><body>
	<h2>Login Successful</h2>
	<p>User: {{.Username}}</p>
	<p>Auth Level: {{.AuthLevel}}</p>`)

	// Debug JSON output
	if isDebugMode(ctx) {
		b.WriteString("<hr><h3>Debug Flow State</h3><pre>")
		jsonState, err := marshalStatePretty(state) // We'll create this helper next
		if err != nil {
			b.WriteString(fmt.Sprintf("Failed to serialize flow state: %v", err))
		} else {
			b.WriteString(jsonState)
		}
		b.WriteString("</pre>")
	}

	b.WriteString(`</body></html>`)

	t := template.Must(template.New("result").Parse(b.String()))
	var sb strings.Builder
	_ = t.Execute(&sb, res)

	ctx.SetBodyString(sb.String())
}

func isDebugMode(ctx *fasthttp.RequestCtx) bool {
	return string(ctx.URI().QueryArgs().Peek("debug")) == "true"
}

func marshalStatePretty(v interface{}) (string, error) {
	out, err := json.MarshalIndent(v, "", "  ")
	return string(out), err
}
