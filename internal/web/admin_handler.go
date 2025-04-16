package web

import (
	"encoding/json"
	"goiam/internal/config"
	"goiam/internal/db/model"
	"goiam/internal/realms"
	"net/http"

	"github.com/valyala/fasthttp"
)

// AdminHandler handles admin API endpoints
type AdminHandler struct {
	userDB model.UserDB
}

// HandleListUsers handles the GET /admin/users endpoint
func handleListUsers(ctx *fasthttp.RequestCtx) {
	// Get tenant and realm from path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Lookup the loaded realm
	_, ok := realms.GetRealm(tenant + "/" + realm)

	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	userDb := config.PostgresUserDB

	if userDb == nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("User database not found")
		return
	}

	// List users from the database
	users, err := userDb.ListUsers(ctx, tenant, realm)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to list users: " + err.Error())
		return
	}

	// Marshal users to JSON with pretty printing
	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal users: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}
