package auth

import (
	"crypto/md5"
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
)

// Loaded assets, we load this from the manifest.json file as the hashes change for every build
var (
	AssetsManifestPath = "templates/static/manifest.json"
	AssetsJSName       string
	AssetsCSSName      string
	AssetsJSContent    []byte
	AssetsCSSContent   []byte
)

//go:embed templates/*
var templatesFS embed.FS

func InitAssets() error {

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

	// Only load CSS if it exists in the manifest
	if AssetsCSSName != "" {
		AssetsCSSContent, err = templatesFS.ReadFile("templates/static/" + AssetsCSSName)
		if err != nil {
			return fmt.Errorf("failed to read CSS asset: %w", err)
		}
	}

	return nil
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
		if AssetsCSSContent == nil {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}
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
