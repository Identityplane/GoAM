package auth

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/valyala/fasthttp"
)

// Render is the single public entry point
func Render(ctx *fasthttp.RequestCtx, flow *model.FlowDefinition, state *model.AuthenticationSession, resultNode *model.GraphNode, prompts map[string]string, baseUrl string) {
	var customMessage string

	// Debug information
	debug := state.Debug
	var stateJSON string
	if debug {
		if js, err := json.MarshalIndent(state, "", "  "); err == nil {
			stateJSON = string(js)
		}
	}

	// Special prompt __redirect is used to redirect the user to a different page
	if redirect, ok := prompts["__redirect"]; ok {
		ctx.Redirect(redirect, fasthttp.StatusSeeOther)
		return
	}

	// get the right tempalte
	currentNode := flow.Nodes[state.Current]

	// Get the template service and load the template
	templatesService := service.GetServices().TemplatesService
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	tmpl, err := templatesService.GetTemplates(tenant, realm, state.FlowId, currentNode.Use)

	if err != nil {
		RenderError(ctx, "Error loading template: "+err.Error(), state, baseUrl)
		return
	}

	// Lookup current node
	currentGraphNode, ok := flow.Nodes[state.Current]
	if !ok {
		RenderError(ctx, "Did not find current graph node: "+state.Current, state, baseUrl)
		return
	}

	// Lookup custom config of node to make it available to the template
	CustomConfig := currentGraphNode.CustomConfig
	if CustomConfig == nil {
		CustomConfig = make(map[string]string)
	}

	flowPath := ""
	if ctx.UserValue("path") != nil {
		flowPath = ctx.UserValue("path").(string)
	}

	loginUri := state.LoginUri
	//Keep debug query parameter if flow is debug mode
	if debug {
		loginUri += "?debug"
	}

	stylePath := baseUrl + "/static/style.css"
	scriptPath := baseUrl + "/static/style.js"

	cspNonce := lib.GenerateSecureSessionID()
	ctx.SetUserValue("cspNonce", cspNonce)

	// Create the view data
	view := &service.ViewData{
		Title:         state.Current,
		NodeName:      state.Current,
		Prompts:       prompts,
		Debug:         debug,
		Error:         resolveErrorMessage(state),
		State:         state,
		StateJSON:     stateJSON,
		FlowName:      currentGraphNode.Name,
		Node:          resultNode,
		Message:       customMessage,
		StylePath:     stylePath,
		ScriptPath:    scriptPath,
		CustomConfig:  CustomConfig,
		Tenant:        ctx.UserValue("tenant").(string),
		Realm:         ctx.UserValue("realm").(string),
		FlowPath:      flowPath,
		LoginUri:      loginUri,
		AssetsJSPath:  baseUrl + "/" + AssetsJSName,
		AssetsCSSPath: baseUrl + "/" + AssetsCSSName,
		CspNonce:      cspNonce,
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", view); err != nil {
		RenderError(ctx, "Render error: "+err.Error(), state, baseUrl)
		return
	}

	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(buf.Bytes())
}

func RenderError(ctx *fasthttp.RequestCtx, msg string, state *model.AuthenticationSession, baseUrl string) {

	if state == nil {

		msg := "Error without state"
		SimpleErrorHtml(ctx, msg)
		return
	}

	// Debug information
	debug := state.Debug
	var stateJSON string
	if debug {
		if js, err := json.MarshalIndent(state, "", "  "); err == nil {
			stateJSON = string(js)
		}
	}

	// CSP
	cspNonce := lib.GenerateSecureSessionID()
	ctx.SetUserValue("cspNonce", cspNonce)

	var stylePath, scriptPath string
	if baseUrl != "" {
		stylePath = baseUrl + "/static/style.css"
		scriptPath = baseUrl + "/static/style.js"
	}

	// Create the view data
	view := &service.ViewData{
		Title:      state.Current,
		NodeName:   state.Current,
		Debug:      debug,
		Error:      msg,
		State:      state,
		StateJSON:  stateJSON,
		StylePath:  stylePath,
		ScriptPath: scriptPath,
		Tenant:     ctx.UserValue("tenant").(string),
		Realm:      ctx.UserValue("realm").(string),
		CspNonce:   cspNonce,
	}

	templatesService := service.GetServices().TemplatesService
	tmpl, err := templatesService.GetErrorTemplate(ctx.UserValue("tenant").(string), ctx.UserValue("realm").(string), state.FlowId)

	if err != nil {

		msg := "Cannot load error template"
		SimpleErrorHtml(ctx, msg)
		return
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "error", view); err != nil {
		msg := "Cannot execute error template: " + err.Error()
		SimpleErrorHtml(ctx, msg)
		return
	}

	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(buf.Bytes())

}

func SimpleErrorHtml(ctx *fasthttp.RequestCtx, msg string) {
	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.SetBodyString(fmt.Sprintf("<html><body><h2>Error</h2><p>%s</p></body></html>", msg))
}

func resolveErrorMessage(state *model.AuthenticationSession) string {
	if state.Error != nil {
		return *state.Error
	}
	return ""
}
