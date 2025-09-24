package webutils

import (
	"fmt"
	"strings"

	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/valyala/fasthttp"
)

func GetUrlForRealm(ctx *fasthttp.RequestCtx, realm *model.Realm) string {

	if realm.BaseUrl != "" {
		return realm.BaseUrl
	}

	return GetFallbackUrl(ctx, realm.Tenant, realm.Realm)
}

// We we dont know the base url we use this to get the full url in the cases we need to assemble urls
func GetFallbackUrl(ctx *fasthttp.RequestCtx, tenant, realm string) string {

	// If we have a server settings that is overwriting the base url we use that one
	overwrite := config.ServerSettings.BaseUrlOverwrite[fmt.Sprintf("%s_%s", tenant, realm)]
	if overwrite != "" {
		return overwrite
	}

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

	// If the host as a port we assume http unless the port is 433 or 8433
	if hasPort && hostParts[1] != "443" && hostParts[1] != "8443" {
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
