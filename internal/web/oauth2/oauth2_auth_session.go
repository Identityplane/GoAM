package oauth2

import (
	"fmt"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/internal/web/auth"
	"github.com/Identityplane/GoAM/internal/web/webutils"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/valyala/fasthttp"
)

func CreateSessionForOauth2Flow(ctx *fasthttp.RequestCtx, realm *model.Realm, flow *model.Flow, oauth2request *model.AuthorizeRequest, acrValue string) (*model.AuthenticationSession, *model.AuthError) {

	baseUrl := webutils.GetUrlForRealm(ctx, realm)
	if baseUrl == "" {
		baseUrl = webutils.GetFallbackUrl(ctx, realm.Tenant, realm.Realm)
	}

	// Use the new auth-ui for login
	loginUri := fmt.Sprintf("%s/authui/%s", baseUrl, flow.Route)
	session, sessionID := service.GetServices().SessionsService.CreateAuthSessionObject(realm.Tenant, realm.Realm, flow.Id, loginUri)

	// set the session id in the url fragment
	uri := fmt.Sprintf("%s#session=%s", loginUri, sessionID)
	session.LoginUriNext = uri

	// Set the http auth context from the request
	auth.SetHttpAuthContextFromRequest(session, ctx)

	// We set the finish url of the auth session to the oauth2/finishauthorize endpoint
	session.FinishUri = fmt.Sprintf("%s/oauth2/finishauthorize?session=%s", baseUrl, sessionID)

	// Set the oauth2 context to the session
	session.Oauth2SessionInformation = &model.Oauth2Session{}
	session.Oauth2SessionInformation.AuthorizeRequest = oauth2request
	session.Oauth2SessionInformation.Acr = acrValue

	return session, nil
}
