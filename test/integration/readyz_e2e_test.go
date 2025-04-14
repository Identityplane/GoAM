package integration

import (
	"goiam/internal/web"
	"net/http"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestPingE2E(t *testing.T) {

	e := *SetupIntegrationTest(t, "")

	e.GET("/readyz").
		Expect().
		Status(http.StatusOK).
		Body().Contains("\"Realms\": \"ready\"").Contains("\"UserRepo\": \"ready\"")
}

func TestCrashRecovery(t *testing.T) {
	e := *SetupIntegrationTest(t, "")

	// Simulate a crash by adding a new route that panics
	Router.GET("/crash", web.WrapMiddleware(func(ctx *fasthttp.RequestCtx) {
		var ptr *int
		_ = *ptr // This will cause a crash
	}))

	// This should be catched by the middleware
	e.GET("/crash").
		Expect().
		Status(http.StatusInternalServerError)

	// This should still work afterwards
	e.GET("/readyz").
		Expect().
		Status(http.StatusOK)
}
