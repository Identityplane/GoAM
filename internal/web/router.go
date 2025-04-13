package web

import (
	"goiam/internal"
	"goiam/internal/web/debug"
	"log"

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

	// Graph routes
	for name, flow := range internal.FlowRegistry {
		r.ANY(flow.Route, NewGraphHandler(flow.Flow).Handle)
		log.Printf("Registered flow %q at route %q", name, flow.Route)
	}

	// Debug routes
	r.GET("/debug/flows", debug.HandleListFlows)
	r.GET("/debug/flow/graph.png", debug.HandleFlowGraphPNG)
	r.GET("/debug/flow/graph.svg", debug.HandleFlowGraphSVG)

	// Static files
	r.GET("/theme/{realm}/{filename}", StaticHandler)

	return r
}
