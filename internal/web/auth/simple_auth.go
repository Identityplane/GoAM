package auth

import (
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/valyala/fasthttp"
)

func CreateSimpleAuthSession(ctx *fasthttp.RequestCtx, flow *model.Flow, session *model.AuthenticationSession, grantType string) *model.AuthError {

	clientID := string(ctx.QueryArgs().Peek("client_id"))

	// If there is a client id in the query params we create a simple auth session
	if clientID != "" {

		// Get the simple auth parameters form the query params
		redirectUrl := string(ctx.QueryArgs().Peek("redirect_uri"))
		scope := string(ctx.QueryArgs().Peek("scope"))
		state := string(ctx.QueryArgs().Peek("state"))
		responseType := string(ctx.QueryArgs().Peek("response_type"))

		// Get the application from the database
		application, ok := service.GetServices().ApplicationService.GetApplication(flow.Tenant, flow.Realm, clientID)
		if !ok {
			authError := model.SimpleAuthErrorInvalidClientID()
			authError.ErrorDescription = "Application not found"
			return authError
		}

		// Create the simple auth request
		req := &model.SimpleAuthRequest{
			ClientID:     clientID,
			RedirectURI:  redirectUrl,
			Scope:        scope,
			State:        state,
			Grant:        grantType,
			ResponseType: responseType,
		}

		// Set the default redirect uri if not set in the request
		if req.Grant == model.GRANT_SIMPLE_AUTH_COOKIE && req.RedirectURI == "" && len(application.RedirectUris) > 0 {
			req.RedirectURI = application.RedirectUris[0]
		}

		// Verify the simple auth flow request
		err := service.GetServices().SimpleAuthService.VerifySimpleAuthFlowRequest(ctx, req, application, flow)
		if err != nil {
			authError := model.SimpleAuthErrorClientUnauthorized()
			authError.ErrorDescription = "Client unauthorized: " + err.Error()
			return authError
		}

		// Add the simple auth session to the execution context
		session.SimpleAuthSessionInformation = &model.SimpleAuthContext{Request: req}
	}

	return nil
}

func FinishSimpleAuthFlow(ctx *fasthttp.RequestCtx, session *model.AuthenticationSession, realm *model.Realm) (*model.SimpleAuthResponse, *model.AuthError) {

	if session.SimpleAuthSessionInformation == nil || session.SimpleAuthSessionInformation.Request == nil {
		// This means the flow was not initialized correctly
		return nil, model.SimpleAuthServerError()
	}

	if session.DidResultError() {
		return nil, model.SimpleAuthServerError()
	}

	if session.DidResultFailure() {
		return nil, model.SimpleAuthFailure()
	}

	simpleAuthResponse, simpleAuthError := service.GetServices().SimpleAuthService.FinishSimpleAuthFlow(ctx, session, realm.Tenant, realm.Realm)
	if simpleAuthError != nil {
		authError := model.SimpleAuthServerError()
		authError.ErrorDescription = "Failed to finish auth flow"
		return nil, authError
	}

	if session.SimpleAuthSessionInformation.Request.Grant == model.GRANT_SIMPLE_AUTH_COOKIE {
		err := finishSimpleAuthCookieGrant(ctx, simpleAuthResponse, session)
		if err != nil {
			return nil, err
		}
	}

	if session.SimpleAuthSessionInformation.Request.Grant == model.GRANT_SIMPLE_AUTH_BODY {
		err := finishSimpleAuthBodyGrant(ctx, simpleAuthResponse, session)
		if err != nil {
			return nil, err
		}
	}

	// Detele the session
	service.GetServices().SessionsService.DeleteAuthenticationSession(ctx, realm.Tenant, realm.Realm, session.SessionIdHash)

	return simpleAuthResponse, nil
}

func finishSimpleAuthCookieGrant(ctx *fasthttp.RequestCtx, simpleAuthResponse *model.SimpleAuthResponse, session *model.AuthenticationSession) *model.AuthError {

	if session.DidResultAuthenticated() {

		// We need to get the application to get the cookie specifications
		application, ok := service.GetServices().ApplicationService.GetApplication(session.Tenant, session.Realm, session.SimpleAuthSessionInformation.Request.ClientID)
		if !ok {
			authError := model.SimpleAuthErrorFound()
			authError.ErrorDescription = "Application not found"
			return authError
		}

		var spec *model.CookieSpecification
		if application.Settings != nil && application.Settings.Cookie != nil {
			spec = application.Settings.Cookie
		}

		if spec == nil {

			// Use default cookie specification if none set in the application
			spec = &model.CookieSpecification{
				Name:          "access_token",
				Secure:        true,
				HttpOnly:      true,
				SameSite:      "Lax",
				SessionExpiry: false,
			}
		}

		// Set the cookie
		cookie := &fasthttp.Cookie{}
		cookie.SetKey(spec.Name)
		cookie.SetValue(simpleAuthResponse.AccessToken)
		cookie.SetSecure(spec.Secure)
		cookie.SetHTTPOnly(spec.HttpOnly)

		if spec.Domain != "" {
			cookie.SetDomain(spec.Domain)
		}

		if spec.Path != "" {
			cookie.SetPath(spec.Path)
		}

		switch spec.SameSite {
		case "Lax":
			cookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
		case "Strict":
			cookie.SetSameSite(fasthttp.CookieSameSiteStrictMode)
		case "None":
			cookie.SetSameSite(fasthttp.CookieSameSiteDefaultMode)
		default:
			cookie.SetSameSite(fasthttp.CookieSameSiteDisabled)
		}

		if !spec.SessionExpiry {
			cookie.SetMaxAge(simpleAuthResponse.ExpiresIn)
		}

		ctx.Response.Header.SetCookie(cookie)
	}

	return nil
}

func finishSimpleAuthBodyGrant(ctx *fasthttp.RequestCtx, simpleAuthResponse *model.SimpleAuthResponse, session *model.AuthenticationSession) *model.AuthError {

	// Nothing to do, renderer should add the values to the body
	return nil
}
