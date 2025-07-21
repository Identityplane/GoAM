package webutils

import (
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
)

// We we dont know the base url we use this to get the full url in the cases we need to assemble urls
func GetFallbackUrl(ctx *fasthttp.RequestCtx, tenant, realm string) string {

	// get the host request header
	host := string(ctx.Request.Header.Peek("Host"))

	if host == "" {
		return ""
	}

	// Check if the host has a port
	hostParts := strings.Split(host, ":")
	hasPort := (len(hostParts) > 1)

	if !hasPort {
		// if there is no port we assume we have https
		return fmt.Sprintf("https://%s/%s/%s", host, tenant, realm)
	}

	// if there is a port we assume we have http as we use this usually for 8080
	return fmt.Sprintf("http://%s/%s/%s", host, tenant, realm)
}

func RedirectTo(ctx *fasthttp.RequestCtx, location string) {
	ctx.SetStatusCode(fasthttp.StatusSeeOther)
	ctx.Response.Header.Set("Location", location)
	ctx.Response.Header.Set("Cache-Control", "no-store")
	ctx.Response.Header.Set("Pragma", "no-cache")
}
