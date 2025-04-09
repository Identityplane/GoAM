package unit

import (
	"goiam/internal/web"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestPingRoute(t *testing.T) {
	r := web.New()

	ctx := fasthttp.RequestCtx{}
	ctx.Init(&fasthttp.Request{}, nil, nil)
	ctx.Request.SetRequestURI("/ping")
	ctx.Request.Header.SetMethod("GET")

	r.Handler(&ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, "pong", string(ctx.Response.Body()))
}
