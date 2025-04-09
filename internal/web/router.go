package web

import (
	"goiam/internal/auth/flows"
	"goiam/internal/db/sqlite"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func New() *router.Router {
	r := router.New()

	r.GET("/ping", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("pong")
	})

	userRepo := sqlite.NewUserRepository()

	loginFlow := flows.NewUsernamePasswordFlow(userRepo)
	r.ANY("/login", NewHttpFlowRunner(loginFlow).Handle)

	registerFlow := flows.NewUserRegistrationFlow(userRepo)
	r.ANY("/register", NewHttpFlowRunner(registerFlow).Handle)

	return r
}
