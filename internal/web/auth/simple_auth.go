package auth

import (
	"fmt"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/valyala/fasthttp"
)

func createSimpleAuthSession(ctx *fasthttp.RequestCtx, flow *model.Flow, session *model.AuthenticationSession) error {

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
			return fmt.Errorf("application not found")
		}

		// Create the simple auth request
		req := &model.SimpleAuthRequest{
			ClientID:     clientID,
			RedirectURI:  redirectUrl,
			Scope:        scope,
			State:        state,
			Grant:        model.GRANT_SIMPLE_AUTH_COOKIE,
			ResponseType: responseType,
		}

		// Set the default redirect uri if not set in the request
		if req.Grant == model.GRANT_SIMPLE_AUTH_COOKIE && req.RedirectURI == "" && len(application.RedirectUris) > 0 {
			req.RedirectURI = application.RedirectUris[0]
		}

		// Verify the simple auth flow request
		err := service.GetServices().SimpleAuthService.VerifySimpleAuthFlowRequest(ctx, req, application, flow)
		if err != nil {
			return fmt.Errorf("failed to verify simple auth flow request: %v", err)
		}

		// Add the simple auth session to the execution context
		session.SimpleAuthSessionInformation = &model.SimpleAuthContext{Request: req}
	}

	return nil
}

func finishSimpleAuthFlow(ctx *fasthttp.RequestCtx, session *model.AuthenticationSession, realm *model.Realm) error {

	simpleAuthResponse, simpleAuthError := service.GetServices().SimpleAuthService.FinishSimpleAuthFlow(ctx, session, realm.Tenant, realm.Realm)
	if simpleAuthError != nil {
		return fmt.Errorf("failed to finish auth flow: %v", simpleAuthError)
	}

	if session.SimpleAuthSessionInformation.Request.Grant == model.GRANT_SIMPLE_AUTH_COOKIE {
		err := finishSimpleAuthCookieGrant(ctx, simpleAuthResponse, session)
		if err != nil {
			return fmt.Errorf("failed to finish simple auth cookie grant: %v", err)
		}
	}

	if session.SimpleAuthSessionInformation.Request.Grant == model.GRANT_SIMPLE_AUTH_BODY {
		err := finishSimpleAuthBodyGrant(ctx, simpleAuthResponse, session)
		if err != nil {
			return fmt.Errorf("failed to finish simple auth body grant: %v", err)
		}
	}

	// Detele the session
	service.GetServices().SessionsService.DeleteAuthenticationSession(ctx, realm.Tenant, realm.Realm, session.SessionIdHash)

	return nil
}

func finishSimpleAuthCookieGrant(ctx *fasthttp.RequestCtx, simpleAuthResponse *model.SimpleAuthResponse, session *model.AuthenticationSession) error {

	if session.DidResultAuthenticated() {

		// We need to get the application to get the cookie specifications
		application, ok := service.GetServices().ApplicationService.GetApplication(session.Tenant, session.Realm, session.SimpleAuthSessionInformation.Request.ClientID)
		if !ok {
			return fmt.Errorf("application not found")
		}

		spec := application.Settings.Cookie
		if spec == nil {

			// Use default cookie specification if none set in the application
			spec = &model.CookieSpecification{
				Name:          "session_id",
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

func finishSimpleAuthBodyGrant(ctx *fasthttp.RequestCtx, simpleAuthResponse *model.SimpleAuthResponse, session *model.AuthenticationSession) error {

	// Nothing to do, renderer should add the values to the body

	return nil
}
