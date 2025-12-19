package oauth2

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Identityplane/GoAM/internal/lib/oauth2"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"

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

	// get the application
	application, ok := service.GetServices().ApplicationService.GetApplication(tenant, realm, session.ClientID)
	if !ok {
		returnBearerTokenError(ctx, oauth2.ErrorInvalidRequest, "Application not found")
		return
	}

	// Get the user claims
	claims, err := getUserClaims(ctx, tenant, realm, session, application)
	if err != nil {
		returnBearerTokenError(ctx, oauth2.ErrorInvalidRequest, "Could not get user claims")
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

	// Not needed by spec but we support access_token via post body as well
	accessToken := string(ctx.PostArgs().Peek("access_token"))
	if accessToken != "" {
		return accessToken, true
	}

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

func getUserClaims(ctx *fasthttp.RequestCtx, tenant, realm string, session *model.ClientSession, application *model.Application) (map[string]interface{}, error) {

	if application != nil && application.Settings != nil && application.Settings.OAuth2Settings != nil && application.Settings.OAuth2Settings.LoadUserFromLoginSession {
		return getUserClaimsFromLoginSession(ctx, tenant, realm, session, application)
	} else {
		return getUserClaimsFromDatabase(ctx, tenant, realm, session, application)
	}

}

func getUserClaimsFromLoginSession(ctx *fasthttp.RequestCtx, tenant, realm string, session *model.ClientSession, application *model.Application) (map[string]interface{}, error) {

	if session.Claims == nil {
		return nil, fmt.Errorf("no claims found in session")
	}

	return session.Claims, nil
}

func getUserClaimsFromDatabase(ctx *fasthttp.RequestCtx, tenant, realm string, session *model.ClientSession, application *model.Application) (map[string]interface{}, error) {

	// Load the user of the session
	user, err := service.GetServices().UserService.GetUserWithAttributesByID(ctx, tenant, realm, session.UserID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("could not load user")
	}

	// Get the user claims
	claims, err := service.GetServices().UserClaimsService.GetUserClaims(*user, session.Scope, nil)
	if err != nil {
		return nil, fmt.Errorf("could not get user claims")
	}

	return claims, nil
}
