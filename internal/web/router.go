package web

import (
	"goiam/internal"
	"goiam/internal/auth/graph/yaml"
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

	flows, err := yaml.LoadFlowsFromDir(internal.FlowsDir)
	if err != nil {
		log.Fatalf("failed to load flows: %v", err)
	}

	for _, flow := range flows {
		r.ANY(flow.Route, NewGraphHandler(flow.Flow).Handle)
		log.Printf("Registered flow %s on %s", flow.Flow.Name, flow.Route)
	}

	return r
}
