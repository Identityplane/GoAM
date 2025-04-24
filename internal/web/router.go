package web

import (
	"encoding/json"
	"goiam/internal/web/admin_api"
	"goiam/internal/web/debug"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func New() *router.Router {
	r := router.New()

	// Set the NotFound handler
	r.NotFound = WrapMiddleware(handleNotFound)

	// Main authentication routes
	r.GET("/{tenant}/{realm}/auth/{path}", WrapMiddleware(HandleAuthRequest))
	r.POST("/{tenant}/{realm}/auth/{path}", WrapMiddleware(HandleAuthRequest))

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
	admin.GET("/{tenant}/{realm}/", admin_api.HandleGetRealm)

	// Debug routes
	r.GET("/debug/flows/all", WrapMiddleware(debug.HandleListAllFlows))
	r.GET("/{tenant}/{realm}/debug/flows", WrapMiddleware(debug.HandleListFlows))
	r.GET("/{tenant}/{realm}/debug/{flow}/graph.png", WrapMiddleware(debug.HandleFlowGraphPNG))
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
