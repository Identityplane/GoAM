package web

import (
	"encoding/json"
	"goiam/internal/service"
	"goiam/internal/web/admin_api"
	"goiam/internal/web/debug"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func New() *router.Router {
	r := router.New()

	// Get services
	services := service.GetServices()

	// Set the NotFound handler
	r.NotFound = WrapMiddleware(handleNotFound)

	// Main authentication routes
	r.GET("/{tenant}/{realm}/auth/{path}", WrapMiddleware(HandleAuthRequest))
	r.POST("/{tenant}/{realm}/auth/{path}", WrapMiddleware(HandleAuthRequest))

	// Admin routes
	adminHandler := admin_api.New(services.UserService)
	r.GET("/{tenant}/{realm}/admin/users", WrapMiddleware(adminHandler.HandleListUsers))
	r.GET("/{tenant}/{realm}/admin/users/stats", WrapMiddleware(adminHandler.HandleGetUserStats))
	r.GET("/{tenant}/{realm}/admin/users/{username}", WrapMiddleware(adminHandler.HandleGetUser))
	r.POST("/{tenant}/{realm}/admin/users/{username}", WrapMiddleware(adminHandler.HandleCreateUser))
	r.PUT("/{tenant}/{realm}/admin/users/{username}", WrapMiddleware(adminHandler.HandleUpdateUser))
	r.DELETE("/{tenant}/{realm}/admin/users/{username}", WrapMiddleware(adminHandler.HandleDeleteUser))
	r.GET("/{tenant}/{realm}/admin/dashboard", WrapMiddleware(adminHandler.HandleDashboard))
	r.GET("/admin/realms", WrapMiddleware(adminHandler.HandleListRealms))

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
