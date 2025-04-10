package web

import (
	"goiam/internal"
	"log"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func New() *router.Router {
	r := router.New()

	r.GET("/ping", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("pong")
	})

	for name, flow := range internal.FlowRegistry {

		r.ANY(flow.Route, NewGraphHandler(flow.Flow).Handle)
		log.Printf("Registered flow %q at route %q", name, flow.Route)
	}

	r.GET("/debug/flows", HandleListFlows)
	r.GET("/debug/flow/graph.png", HandleFlowGraphPNG)
	r.GET("/debug/flow/graph.svg", HandleFlowGraphSVG)

	return r
}
