package auth_ui

import (
	"log"
	"net/http"
	"net/url"

	"net/http/httputil"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var (
	devServerURL *url.URL
	proxy        *httputil.ReverseProxy
)

func init() {
	// In development, the Next.js server typically runs on port 3000
	var err error
	devServerURL, err = url.Parse("http://localhost:3000")
	if err != nil {
		log.Fatalf("Failed to parse dev server URL: %v", err)
	}
	proxy = httputil.NewSingleHostReverseProxy(devServerURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = devServerURL.Host
	}
}

// HandleAuthUIProxy proxies requests from /tenant/realm/authui/... to the local Next.js dev server.
func HandleAuthUIProxy(ctx *fasthttp.RequestCtx) {
	// Let the proxied server handle its own security headers (e.g. Next.js in dev mode)
	ctx.SetUserValue("cspDisable", true)

	// Convert fasthttp request to standard net/http request for the proxy handler
	httpHandler := fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}))
	httpHandler(ctx)
}
