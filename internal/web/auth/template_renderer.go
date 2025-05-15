package auth

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"goiam/internal/model"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/valyala/fasthttp"
)

// Configuration
var (
	baseTemplates      *template.Template
	LayoutTemplatePath = "templates/layout.html"
	NodeTemplatesPath  = "templates/nodes"
	ComponentsPath     = "templates/components"
)

// ViewData is passed to all templates for dynamic rendering
type ViewData struct {
	Title        string
	NodeName     string
	Prompts      map[string]string
	Debug        bool
	Error        string
	State        *model.AuthenticationSession
	StateJSON    string
	FlowName     string
	StylePath    string
	ScriptPath   string
	Message      string
	CustomConfig map[string]string
	Tenant       string
	Realm        string
	FlowPath     string
	LoginUri     string
}

//go:embed templates/*
var templatesFS embed.FS

// loadComponents loads all component templates from the components directory
func loadComponents(tmpl *template.Template) error {
	entries, err := templatesFS.ReadDir(ComponentsPath)
	if err != nil {
		return fmt.Errorf("failed to read components directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			componentPath := filepath.Join(ComponentsPath, entry.Name())
			_, err := tmpl.ParseFS(templatesFS, componentPath)
			if err != nil {
				return fmt.Errorf("failed to parse component %s: %w", entry.Name(), err)
			}
		}
	}
	return nil
}

// InitTemplates loads and parses the base layout template and all components
func InitTemplates() error {
	// Parse the base layout template
	tmpl, err := template.New("layout").Funcs(template.FuncMap{
		"title": title,
	}).ParseFS(templatesFS, LayoutTemplatePath)

	if err != nil {
		return fmt.Errorf("failed to parse base template: %w", err)
	}

	// Load all component templates
	if err := loadComponents(tmpl); err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	baseTemplates = tmpl
	return nil
}

// Render is the single public entry point
func Render(ctx *fasthttp.RequestCtx, flow *model.FlowDefinition, state *model.AuthenticationSession, resultNode *model.GraphNode, prompts map[string]string, baseUrl string) {
	var templateFile string
	var customMessage string

	// Debug information
	debug := isDebugMode(ctx)
	var stateJSON string
	if debug {
		if js, err := json.MarshalIndent(state, "", "  "); err == nil {
			stateJSON = string(js)
		}
	}

	// Choosing the right template file
	if state.Result != nil {
		templateFile = "result.html"
		customMessage = resultNode.CustomConfig["message"]
	} else {
		currentNode := flow.Nodes[state.Current]
		templateFile = fmt.Sprintf("%s.html", currentNode.Use)
	}

	// Lookup current node
	currentGraphNode, ok := flow.Nodes[state.Current]
	if !ok {
		RenderError(ctx, "Did not find current graph node: "+state.Current)
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
	stylePath := baseUrl + "/static/style.css"
	scriptPath := baseUrl + "/static/style.js"

	// Create the view data
	view := &ViewData{
		Title:        state.Current,
		NodeName:     state.Current,
		Prompts:      prompts,
		Debug:        debug,
		Error:        resolveErrorMessage(state),
		State:        state,
		StateJSON:    stateJSON,
		FlowName:     currentGraphNode.Name,
		Message:      customMessage,
		StylePath:    stylePath,
		ScriptPath:   scriptPath,
		CustomConfig: CustomConfig,
		Tenant:       ctx.UserValue("tenant").(string),
		Realm:        ctx.UserValue("realm").(string),
		FlowPath:     flowPath,
		LoginUri:     loginUri,
	}

	// Clone the base template
	tmpl, err := baseTemplates.Clone()
	if err != nil {
		RenderError(ctx, "Template clone error: "+err.Error())
		return
	}

	// Parse the node template
	filepath := filepath.Join(NodeTemplatesPath, templateFile)
	_, err = tmpl.ParseFS(templatesFS, filepath)
	if err != nil {
		RenderError(ctx, "Parse error: "+err.Error())
		return
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", view); err != nil {
		RenderError(ctx, "Render error: "+err.Error())
		return
	}

	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(buf.Bytes())
}

func RenderError(ctx *fasthttp.RequestCtx, msg string) {
	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.SetBodyString(fmt.Sprintf("<html><body><h2>Error</h2><p>%s</p></body></html>", msg))
}

func isDebugMode(ctx *fasthttp.RequestCtx) bool {

	debugParam := ctx.URI().QueryArgs().Peek("debug")
	return debugParam != nil
}

func resolveErrorMessage(state *model.AuthenticationSession) string {
	if state.Error != nil {
		return *state.Error
	}
	return ""
}

func title(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(s[0]-32) + s[1:] // capitalize first letter
}
