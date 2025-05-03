package web

import (
	"encoding/json"
	"goiam/internal/web/admin_api"
	"goiam/internal/web/auth"
	"goiam/internal/web/debug"
	"goiam/internal/web/oauth2"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func New() *router.Router {
	r := router.New()

	// Set the NotFound handler
	r.NotFound = WrapMiddleware(handleNotFound)

	// Main authentication routes
	r.GET("/{tenant}/{realm}/auth/{path}", WrapMiddleware(auth.HandleAuthRequest))
	r.POST("/{tenant}/{realm}/auth/{path}", WrapMiddleware(auth.HandleAuthRequest))

	// Admin routes
	admin := r.Group("/admin")
	admin.GET("/{tenant}/{realm}/users", WrapMiddleware(admin_api.HandleListUsers))
	admin.GET("/{tenant}/{realm}/users/stats", WrapMiddleware(admin_api.HandleGetUserStats))
	admin.GET("/{tenant}/{realm}/users/{username}", WrapMiddleware(admin_api.HandleGetUser))
	admin.POST("/{tenant}/{realm}/users/{username}", WrapMiddleware(admin_api.HandleCreateUser))
	admin.PUT("/{tenant}/{realm}/users/{username}", WrapMiddleware(admin_api.HandleUpdateUser))
	admin.DELETE("/{tenant}/{realm}/users/{username}", WrapMiddleware(admin_api.HandleDeleteUser))

	admin.GET("/{tenant}/{realm}/dashboard", WrapMiddleware(admin_api.HandleDashboard))

	admin.GET("/realms", WrapMiddleware(admin_api.HandleListRealms))
	admin.GET("/{tenant}/{realm}/", WrapMiddleware(admin_api.HandleGetRealm))
	admin.POST("/{tenant}/{realm}/", WrapMiddleware(admin_api.HandleCreateRealm))
	admin.PATCH("/{tenant}/{realm}/", WrapMiddleware(admin_api.HandleUpdateRealm))
	admin.DELETE("/{tenant}/{realm}/", WrapMiddleware(admin_api.HandleDeleteRealm))

	// Application management routes
	admin.GET("/{tenant}/{realm}/applications", WrapMiddleware(admin_api.HandleListApplications))
	admin.GET("/{tenant}/{realm}/applications/{client_id}", WrapMiddleware(admin_api.HandleGetApplication))
	admin.POST("/{tenant}/{realm}/applications/{client_id}", WrapMiddleware(admin_api.HandleCreateApplication))
	admin.PUT("/{tenant}/{realm}/applications/{client_id}", WrapMiddleware(admin_api.HandleUpdateApplication))
	admin.DELETE("/{tenant}/{realm}/applications/{client_id}", WrapMiddleware(admin_api.HandleDeleteApplication))
	admin.POST("/{tenant}/{realm}/applications/{client_id}/regenerate-secret", WrapMiddleware(admin_api.HandleRegenerateClientSecret))

	// Flow management routes
	admin.GET("/{tenant}/{realm}/flows", WrapMiddleware(admin_api.HandleListFlows))
	admin.GET("/{tenant}/{realm}/flows/{flow}", WrapMiddleware(admin_api.HandleGetFlow))
	admin.POST("/{tenant}/{realm}/flows/{flow}", WrapMiddleware(admin_api.HandleCreateFlow))
	admin.PATCH("/{tenant}/{realm}/flows/{flow}", WrapMiddleware(admin_api.HandleUpdateFlow))
	admin.DELETE("/{tenant}/{realm}/flows/{flow}", WrapMiddleware(admin_api.HandleDeleteFlow))

	// Flow defintion routes
	admin.POST("/{tenant}/{realm}/flows/validate", WrapMiddleware(admin_api.HandleValidateFlowDefinition))
	admin.GET("/{tenant}/{realm}/flows/{flow}/definition", WrapMiddleware(admin_api.HandleGetFlowDefintion))
	admin.PUT("/{tenant}/{realm}/flows/{flow}/definition", WrapMiddleware(admin_api.HandlePutFlowDefintion))

	// Node management routes
	admin.GET("/nodes", WrapMiddleware(admin_api.HandleListNodes))

	// Debug routes
	r.GET("/debug/flows/all", WrapMiddleware(debug.HandleListAllFlows))
	r.GET("/{tenant}/{realm}/debug/flows", WrapMiddleware(debug.HandleListFlows))
	r.GET("/{tenant}/{realm}/debug/{flow}/graph.svg", WrapMiddleware(debug.HandleFlowGraphSVG))

	// Static files
	r.GET("/{tenant}/{realm}/static/{filename}", WrapMiddleware(StaticHandler))

	// Health endpoints
	r.GET("/healthz", WrapMiddleware(handleLiveness))
	r.GET("/readyz", WrapMiddleware(handleReadiness))
	r.GET("/info", WrapMiddleware(handleInfo))

	// Swagger UI
	r.GET("/swagger/", WrapMiddleware(HandleSwaggerUI))
	r.GET("/swagger/{*path}", WrapMiddleware(HandleSwaggerUI))

	// Oauth + OIDC
	r.GET("/{tenant}/{realm}/oauth2/.well-known/openid-configuration", WrapMiddleware(oauth2.HandleOpenIDConfiguration))
	r.GET("/{tenant}/{realm}/oauth2/authorize", WrapMiddleware(oauth2.HandleAuthorize))
	return r
}

// handleNotFound is the fallback handler for unmatched routes
func handleNotFound(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetContentType("application/json")
	_ = json.NewEncoder(ctx).Encode(map[string]string{
		"error": "not found",
		"path":  string(ctx.Path()),
	})
}
