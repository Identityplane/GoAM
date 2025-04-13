package web

import (
	"os"
	"path/filepath"

	"github.com/valyala/fasthttp"
)

// Base directory for static files
const staticBaseDir = "../config/renderer/static"

// Default realm
const defaultRealm = "default"

// StaticHandler serves static content for a given realm
func StaticHandler(ctx *fasthttp.RequestCtx) {

	// Extract realm and filename from the URL
	realm := ctx.UserValue("realm").(string)
	filename := ctx.UserValue("filename").(string)

	// Default to the "default" realm if not implemented
	if realm != defaultRealm {
		realm = defaultRealm
	}

	// Construct the file path
	filePath := filepath.Join(staticBaseDir, realm, filepath.Clean(filename))

	// Check if the file exists
	if !fileExists(filePath) {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString("File not found")
		return
	}

	// Serve the file
	ctx.SendFile(filePath)
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {

	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
