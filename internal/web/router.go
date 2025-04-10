package web

import (
	"goiam/internal/auth/graph"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func New() *router.Router {
	r := router.New()

	r.GET("/ping", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("pong")
	})

	//userRepo := sqlite.NewUserRepository()

	r.ANY("/login", NewGraphHandler(graph.UsernamePasswordAuthFlow).Handle)
	r.ANY("/register", NewGraphHandler(graph.UserRegisterFlow).Handle)

	return r
}
