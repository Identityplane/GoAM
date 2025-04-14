package web

import (
	"goiam/internal/web/debug"

	"github.com/fasthttp/router"
)

func New() *router.Router {
	r := router.New()

	// Main authentication routes
	r.ANY("/{tenant}/{realm}/auth/{path}", WrapMiddleware(HandleAuthRequest))

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

	return r
}
