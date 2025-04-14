package web

import (
	"goiam/internal/web/debug"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func New() *router.Router {
	r := router.New()

	// Ping routes
	r.GET("/ping", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("pong")
	})

	// Main authentication routes
	r.ANY("/{tenant}/{realm}/auth/{path}", HandleAuthRequest)

	// Debug routes
	r.GET("/debug/flows/all", debug.HandleListAllFlows)
	r.GET("/{tenant}/{realm}/debug/flows", debug.HandleListFlows)
	r.GET("/{tenant}/{realm}/debug/{flow}/graph.png", debug.HandleFlowGraphPNG)
	r.GET("/{tenant}/{realm}/debug/{flow}/graph.svg", debug.HandleFlowGraphSVG)

	// Static files
	r.GET("/{tenant}/{realm}/static/{filename}", StaticHandler)

	return r
}
