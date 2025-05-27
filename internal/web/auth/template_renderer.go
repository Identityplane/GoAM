package auth

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"goiam/internal/lib"
	"goiam/internal/model"
	"html/template"
	"path/filepath"
	"strings"

	"crypto/md5"

	"github.com/valyala/fasthttp"
)

// Configuration
var (
	baseTemplates      *template.Template
	LayoutTemplatePath = "templates/layout.html"
	ErrorTemplatePath  = "templates/error.html"
	NodeTemplatesPath  = "templates/nodes"
	ComponentsPath     = "templates/components"
	AssetsManifestPath = "templates/static/manifest.json"
)

// Loaded asset
var (
	AssetsJSName  string
	AssetsCSSName string

	AssetsJSContent  []byte
	AssetsCSSContent []byte
)

// ViewData is passed to all templates for dynamic rendering
type ViewData struct {
	Title         string
	NodeName      string
	Prompts       map[string]string
	Debug         bool
	Error         string
	State         *model.AuthenticationSession
	StateJSON     string
	FlowName      string
	Node          *model.GraphNode
	StylePath     string
	ScriptPath    string
	Message       string
	CustomConfig  map[string]string
	Tenant        string
	Realm         string
	FlowPath      string
	LoginUri      string
	AssetsJSPath  string
	AssetsCSSPath string
	CspNonce      string
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

	// Initialize the assets
	// Read the manifest file from the templates FS
	manifest, err := templatesFS.ReadFile(AssetsManifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse the manifest
	var manifestData struct {
		JS  string `json:"js"`
		CSS string `json:"css"`
	}

	if err := json.Unmarshal(manifest, &manifestData); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	AssetsJSName = manifestData.JS
	AssetsCSSName = manifestData.CSS

	AssetsJSContent, err = templatesFS.ReadFile("templates/static/" + AssetsJSName)
	if err != nil {
		return fmt.Errorf("failed to read JS asset: %w", err)
	}

	AssetsCSSContent, err = templatesFS.ReadFile("templates/static/" + AssetsCSSName)
	if err != nil {
		return fmt.Errorf("failed to read CSS asset: %w", err)
	}

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

	// Special prompt __redirect is used to redirect the user to a different page
	if redirect, ok := prompts["__redirect"]; ok {
		ctx.Redirect(redirect, fasthttp.StatusSeeOther)
		return
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
	view := &ViewData{
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

	// Clone the base template
	tmpl, err := baseTemplates.Clone()
	if err != nil {
		RenderError(ctx, "Template clone error: "+err.Error(), state, baseUrl)
		return
	}

	// Parse the node template
	filepath := filepath.Join(NodeTemplatesPath, templateFile)
	_, err = tmpl.ParseFS(templatesFS, filepath)
	if err != nil {
		RenderError(ctx, "Parse error: "+err.Error(), state, baseUrl)
		return
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
	debug := isDebugMode(ctx)
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
	view := &ViewData{
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

	tmpl, err := getErrorTemplate()

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

func getErrorTemplate() (*template.Template, error) {
	// Parse the base layout template
	tmpl, err := template.New("error").Funcs(template.FuncMap{
		"title": title,
	}).ParseFS(templatesFS, ErrorTemplatePath)

	if err != nil {
		return nil, fmt.Errorf("failed to parse base template: %w", err)
	}

	return tmpl, nil
}

// HandleStaticAssets serves the static assets from memory
// This is used to serve the assets for the global static files
// like the main.js and style.css that are idependend from the realm
// served at /{tenant}/{realm}/assets/{filename}
func HandleStaticAssets(ctx *fasthttp.RequestCtx) {
	filename := ctx.UserValue("filename").(string)

	// Get the appropriate content and content type
	var content []byte
	var contentType string
	if strings.HasSuffix(filename, ".js") {
		content = AssetsJSContent
		contentType = "text/javascript"
	} else if strings.HasSuffix(filename, ".css") {
		content = AssetsCSSContent
		contentType = "text/css"
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	// Generate ETag from content hash
	etag := fmt.Sprintf(`"%x"`, md5.Sum(content))

	// Check if client has a matching ETag
	if match := ctx.Request.Header.Peek("If-None-Match"); match != nil {
		if string(match) == etag {
			ctx.SetStatusCode(fasthttp.StatusNotModified)
			return
		}
	}

	// Set caching headers
	ctx.Response.Header.Set("ETag", etag)
	ctx.Response.Header.Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	ctx.Response.Header.Set("Vary", "Accept-Encoding")
	ctx.SetContentType(contentType)
	ctx.SetBody(content)
	ctx.SetStatusCode(fasthttp.StatusOK)
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
