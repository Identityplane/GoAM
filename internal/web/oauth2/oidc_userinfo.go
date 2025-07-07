package oauth2

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/gianlucafrei/GoAM/internal/lib/oauth2"
	"github.com/gianlucafrei/GoAM/internal/service"

	"github.com/valyala/fasthttp"
)

func HandleUserinfoEndpoint(ctx *fasthttp.RequestCtx) {

	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	accessToken, ok := readAccessTokenFromRequest(ctx)
	if !ok {
		returnBearerTokenError(ctx, oauth2.ErrorInvalidRequest, "No access token provided")
		return
	}

	// Validate the access token by reading the session
	session, err := service.GetServices().SessionsService.GetClientSessionByAccessToken(ctx, tenant, realm, accessToken)
	if err != nil {
		returnBearerTokenError(ctx, oauth2.ErrorInvalidRequest, "The Access Token is invalid")
		return
	}

	// Get the user claims
	claims, err := service.GetServices().OAuth2Service.GetUserClaims(session)
	if err != nil {
		returnBearerTokenError(ctx, oauth2.ErrorInvalidRequest, "The Access Token is invalid")
		return
	}

	// Marshal the claims into json
	jsonClaims, err := json.Marshal(claims)
	if err != nil {
		returnBearerTokenError(ctx, oauth2.ErrorServerError, "Error marshalling claims")
		return
	}

	// Return the claims
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetBody(jsonClaims)
}

func readAccessTokenFromRequest(ctx *fasthttp.RequestCtx) (string, bool) {
	authorizationHeader := ctx.Request.Header.Peek("Authorization")
	if len(authorizationHeader) == 0 {
		return "", false
	}

	// Client should use the bearer format however we are quite lax and just split per space and then take the last part
	parts := bytes.Split(authorizationHeader, []byte(" "))
	if len(parts) == 0 {
		return "", false
	}

	token := parts[len(parts)-1]

	return string(token), true
}

func returnBearerTokenError(ctx *fasthttp.RequestCtx, errorCode string, errorDescription string) {

	// return 401 error according to the oidc specificaion
	/* HTTP/1.1 401 Unauthorized
	WWW-Authenticate: Bearer error="invalid_token",
	  error_description="The Access Token expired"*/

	ctx.SetStatusCode(fasthttp.StatusUnauthorized)
	ctx.Response.Header.Set("WWW-Authenticate", fmt.Sprintf("Bearer error=\"%s\", error_description=\"%s\"", errorCode, errorDescription))
}
