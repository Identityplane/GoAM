package webutils

import (
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
)

// We we dont know the base url we use this to get the full url in the cases we need to assemble urls
func GetFallbackUrl(ctx *fasthttp.RequestCtx, tenant, realm string) string {

	requestUrl := GetUrlOfRequest(ctx)
	return fmt.Sprintf("%s/%s/%s", requestUrl, tenant, realm)
}

func RedirectTo(ctx *fasthttp.RequestCtx, location string) {
	ctx.SetStatusCode(fasthttp.StatusSeeOther)
	ctx.Response.Header.Set("Location", location)
	ctx.Response.Header.Set("Cache-Control", "no-store")
	ctx.Response.Header.Set("Pragma", "no-cache")
}

func GetAdminBaseUrl(ctx *fasthttp.RequestCtx) string {

	requestUrl := GetUrlOfRequest(ctx)
	return fmt.Sprintf("%s/admin", requestUrl)
}

func GetUrlOfRequest(ctx *fasthttp.RequestCtx) string {

	// get the host request header
	host := string(ctx.Request.Header.Peek("Host"))

	if host == "" {
		return ""
	}

	// Check if the host has a port
	hostParts := strings.Split(host, ":")
	hasPort := (len(hostParts) > 1)

	protocol := "https"

	// If the host as a port we assume http
	if hasPort {
		protocol = "http"
	}

	// if the host is 127.0.0.1 we assume we are in development and use http
	if host == "127.0.0.1" {
		protocol = "http"
	}

	// If the host is localhost we assume we are in development and use http
	if host == "localhost" {
		protocol = "http"
	}

	return fmt.Sprintf("%s://%s", protocol, host)
}
