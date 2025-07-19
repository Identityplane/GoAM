package web

import (
	"os"
	"path/filepath"

	"github.com/Identityplane/GoAM/internal/config"

	"github.com/valyala/fasthttp"
)

// StaticHandler serves static content for a given realm
// @Summary Serve static files
// @Description Serves static files (images, CSS, JavaScript, etc.) for a specific tenant and realm
// @Tags Static
// @Produce application/octet-stream
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param filename path string true "Name of the static file to serve"
// @Success 200 {file} binary "The requested static file"
// @Failure 404 {string} string "File not found"
// @Router /{tenant}/{realm}/static/{filename} [get]
func StaticHandler(ctx *fasthttp.RequestCtx) {

	// Extract realm and filename from the URL
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	filename := ctx.UserValue("filename").(string)

	// Construct the file path
	filePath := filepath.Join(config.ConfigPath, "tenants", tenant, realm, "static", filepath.Clean(filename))

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
