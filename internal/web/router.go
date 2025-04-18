package web

import (
	"goiam/internal/db/service"
	"goiam/internal/web/admin_api"
	"goiam/internal/web/debug"

	"github.com/fasthttp/router"
)

func New(userAdminService service.UserAdminService) *router.Router {
	r := router.New()

	// Main authentication routes
	r.ANY("/{tenant}/{realm}/auth/{path}", WrapMiddleware(HandleAuthRequest))

	// Admin routes
	adminHandler := admin_api.New(userAdminService)
	r.GET("/{tenant}/{realm}/admin/users", WrapMiddleware(adminHandler.HandleListUsers))
	r.GET("/{tenant}/{realm}/admin/users/stats", WrapMiddleware(adminHandler.HandleGetUserStats))
	r.GET("/{tenant}/{realm}/admin/users/{username}", WrapMiddleware(adminHandler.HandleGetUser))
	r.POST("/{tenant}/{realm}/admin/users/{username}", WrapMiddleware(adminHandler.HandleCreateUser))
	r.PUT("/{tenant}/{realm}/admin/users/{username}", WrapMiddleware(adminHandler.HandleUpdateUser))
	r.DELETE("/{tenant}/{realm}/admin/users/{username}", WrapMiddleware(adminHandler.HandleDeleteUser))

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
