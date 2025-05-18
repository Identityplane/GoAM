package web

import (
	"bytes"
	"embed"
	"net/http"
	"path"
	"strings"

	"github.com/valyala/fasthttp"
)

//go:embed swagger-ui/*
var swaggerUIFiles embed.FS

// HandleSwaggerUI serves the Swagger UI
func HandleSwaggerUI(ctx *fasthttp.RequestCtx) {
	// Get the requested path
	requestPath := string(ctx.Path())
	if requestPath == "/swagger" || requestPath == "/swagger/" {
		requestPath = "/swagger/index.html"
	}

	// Remove the /swagger prefix
	filePath := strings.TrimPrefix(requestPath, "/swagger/")
	if filePath == "" {
		filePath = "index.html"
	}

	// If this is the OpenAPI specification
	if filePath == "doc.json" {
		HandleOpenAPISpec(ctx)
		return
	}

	// Read the file from the embedded filesystem
	data, err := swaggerUIFiles.ReadFile("swagger-ui/" + filePath)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Not Found")
		return
	}

	// Set the appropriate content type
	ext := path.Ext(filePath)
	switch ext {
	case ".html":
		ctx.SetContentType("text/html")
	case ".css":
		ctx.SetContentType("text/css")
	case ".js":
		ctx.SetContentType("application/javascript")
	case ".png":
		ctx.SetContentType("image/png")
	case ".json":
		ctx.SetContentType("application/json")
	default:
		ctx.SetContentType("text/plain")
	}

	ctx.SetUserValue("csp", "")

	// If this is the index.html, we need to replace the Swagger UI configuration
	if filePath == "index.html" {
		// Replace the Swagger UI configuration
		config := []byte(`
window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/swagger/doc.json",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });
};
`)
		data = bytes.Replace(data, []byte("window.onload = function() {"), config, 1)
	}

	ctx.SetBody(data)
}

// HandleOpenAPISpec serves the OpenAPI specification
func HandleOpenAPISpec(ctx *fasthttp.RequestCtx) {
	// Read the swagger.json file from the embedded filesystem
	data, err := swaggerUIFiles.ReadFile("swagger-ui/swagger.json")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to read Swagger specification")
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(data)
}
