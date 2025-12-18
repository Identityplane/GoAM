package service

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/internal/lib/oauth2"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

// OAuth2Service handles OAuth2 related operations
type OAuth2Service struct {
}

// NewOAuth2Service creates a new OAuth2Service instance
func NewOAuth2Service() services_interface.OAuth2Service {
	return &OAuth2Service{}
}

func NewOAuth2Error(errorCode string, errorDescription string) *oauth2.OAuth2Error {
	errorResponse := oauth2.OAuth2Error{
		Error:            errorCode,
		ErrorDescription: errorDescription,
	}
	return &errorResponse
}

// ValidateOAuth2AuthorizationRequest validates the OAuth2 authorization request
func (s *OAuth2Service) ValidateOAuth2AuthorizationRequest(oauth2request *model.AuthorizeRequest, tenant, realm string, application *model.Application, flowId string) *oauth2.OAuth2Error {

	// Validate the redirect uri
	oauth2error := s.ValidateRedirectUri(oauth2request, application)
	if oauth2error != nil {
		return oauth2error
	}

	// Check which flow is requested, we differenciate between authorization_code and authorization_code_pkce, and client_credentials
	// If we have a code challenge and grant type code it is a pkce flow
	var oauth2_flow oauth2.OAuth2GrantType = oauth2.Oauth2_InvalidFlow
	if oauth2request.CodeChallenge != "" && oauth2request.ResponseType == "code" {
		oauth2_flow = oauth2.Oauth2_AuthorizationCodePKCE
	} else if oauth2request.ResponseType == "code" {
		oauth2_flow = oauth2.Oauth2_AuthorizationCode
	} else {
		// Return invalid flow
		return oauth2.NewOAuth2Error(oauth2.ErrorUnsupportedResponseType, "Unsupported response type, authroization endpoint only supports code grant")
	}

	// Check if the grant type is allowed
	if !slices.Contains(application.AllowedGrants, string(oauth2_flow)) {
		return oauth2.NewOAuth2Error(oauth2.ErrorUnauthorizedClient, "Grant type not allowed")
	}

	// if the application is public is must have a code_challenge and code_challenge_method for the pkce flow
	if !application.Confidential && (oauth2request.CodeChallenge == "" || oauth2request.CodeChallengeMethod == "") {
		return oauth2.NewOAuth2Error(oauth2.ErrorInvalidRequest, "Code challenge and code challenge method are required for public applications")
	}

	// if the application is public the code challenge must be S256
	if !application.Confidential && oauth2request.CodeChallengeMethod != "S256" {
		return oauth2.NewOAuth2Error(oauth2.ErrorInvalidRequest, "Only CodeChallengeMethod S256 is supported")
	}

	// Check if all requested scopes are allowed for each request scope
	for _, scope := range oauth2request.Scope {
		if !slices.Contains(application.AllowedScopes, scope) {
			return oauth2.NewOAuth2Error(oauth2.ErrorInvalidScope, "Invalid scope "+scope)
		}
	}

	// If CodeChallengeMethod is provided it must be S256
	if oauth2request.CodeChallengeMethod != "" && oauth2request.CodeChallengeMethod != "S256" {
		return oauth2.NewOAuth2Error(oauth2.ErrorInvalidRequest, "Code challenge method must be S256")
	}

	// If there is no allowed authentication flow we fail with a server error
	if len(application.AllowedAuthenticationFlows) == 0 {
		return oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. No allowed authentication flows")
	}

	// Check if the flow is allowed for the application
	if !slices.Contains(application.AllowedAuthenticationFlows, flowId) {
		return oauth2.NewOAuth2Error(oauth2.ErrorUnauthorizedClient, "Flow not allowed")
	}

	// If all is good we return nil
	return nil
}

func (s *OAuth2Service) FinishOauth2AuthorizationEndpoint(session *model.AuthenticationSession, tenant, realm string) (*oauth2.AuthorizationResponse, *oauth2.OAuth2Error) {
	if session.Oauth2SessionInformation == nil {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. No oauth2 session information")
	}

	if session.Oauth2SessionInformation.AuthorizeRequest == nil {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. No authorize request")
	}

	// If there is no result yet the ProcessAuthRequest rendered a response and we can return here
	if session.Result == nil {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. No result")
	}

	// if there is an error we return an oauth2 error

	if session.DidResultError() {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Unexpected result node")
	}

	// If the result node is a failure result we return an oauth2 error
	if !session.DidResultAuthenticated() {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorAccessDenied, "Authentication Failed")
	}

	// If the result user is is empty we return an error
	if session.Result.UserID == "" {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. No user found in result")
	}

	// If all ok we create a client session and issue an auth code
	scope := session.Oauth2SessionInformation.AuthorizeRequest.Scope

	// Get the user claims
	userClaims, err := s.GetUserClaims(*session.User, strings.Join(scope, " "), session.Oauth2SessionInformation)
	if err != nil {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not get user claims")
	}

	// Get the realm
	loadedRealm, ok := GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not get realm")
	}

	// Create the client session
	authCode, _, err := GetServices().SessionsService.CreateAuthCodeSession(
		context.Background(),
		tenant,
		realm,
		session.Oauth2SessionInformation.AuthorizeRequest.ClientID,
		session.Result.UserID,
		scope,
		"authorization_code",
		session.Oauth2SessionInformation.AuthorizeRequest.CodeChallenge,
		session.Oauth2SessionInformation.AuthorizeRequest.CodeChallengeMethod,
		session,
		userClaims)

	if err != nil {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not create session")
	}

	// Create the authorization response
	response := oauth2.AuthorizationResponse{
		Code:  authCode,
		State: session.Oauth2SessionInformation.AuthorizeRequest.State,
		Iss:   loadedRealm.Config.BaseUrl,
	}

	return &response, nil
}

func (s *OAuth2Service) ProcessTokenRequest(tenant, realm string, tokenRequest *oauth2.Oauth2TokenRequest, clientAuthentication *oauth2.Oauth2ClientAuthentication) (*oauth2.Oauth2TokenResponse, *oauth2.OAuth2Error) {

	// First we need to validate the client authentication if the client is confidential
	application, ok := GetServices().ApplicationService.GetApplication(tenant, realm, tokenRequest.ClientID)
	if !ok {
		return nil, NewOAuth2Error(oauth2.ErrorUnauthorizedClient, "Invalid client ID")
	}

	if application.Confidential {

		valid, err := GetServices().ApplicationService.VerifyClientSecret(tenant, realm, clientAuthentication.ClientID, clientAuthentication.ClientSecret)
		if err != nil {
			return nil, NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not verify client secret")
		}

		if !valid {
			return nil, NewOAuth2Error(oauth2.ErrorUnauthorizedClient, "Invalid client authentication")
		}
	}

	// Ensure that client_id in token request and clientAuthentication are the same
	if tokenRequest.ClientID != clientAuthentication.ClientID {
		return nil, NewOAuth2Error(oauth2.ErrorInvalidRequest, "Client ID mismatch")
	}

	// Ensure that the grant_type is allowed for the application, we check this already in the validateOAuth2AuthorizationRequest
	// but for the token request or refresh token we need to check if again

	if slices.Contains(application.AllowedGrants, string(oauth2.Oauth2_AuthorizationCodePKCE)) && tokenRequest.GrantType == "authorization_code" {
		// special case for the pkce flow
	} else if !slices.Contains(application.AllowedGrants, tokenRequest.GrantType) {
		return nil, NewOAuth2Error(oauth2.ErrorUnauthorizedClient, "Grant type not allowed")
	}

	// if the grant type is authorization_code we need to create an access token by looking up the auth code in the client sessions
	switch tokenRequest.GrantType {
	case "authorization_code":

		return s.processTokenRequestForAuthorizationCodeGrant(tenant, realm, tokenRequest, clientAuthentication, application)
	case "refresh_token":

		return s.processTokenRequestForRefreshTokenGrant(tenant, realm, tokenRequest, clientAuthentication, application)

	case "client_credentials":

		return s.processTokenRequestForClientCredentialsGrant(tenant, realm, tokenRequest, clientAuthentication, application)
	}

	return nil, NewOAuth2Error(oauth2.ErrorInvalidRequest, "Invalid grant type")
}

func (s *OAuth2Service) processTokenRequestForClientCredentialsGrant(tenant string, realm string, tokenRequest *oauth2.Oauth2TokenRequest, clientAuthentication *oauth2.Oauth2ClientAuthentication, application *model.Application) (*oauth2.Oauth2TokenResponse, *oauth2.OAuth2Error) {

	// Ensure that this is only allowed for confidential applications
	if !application.Confidential {
		return nil, NewOAuth2Error(oauth2.ErrorUnauthorizedClient, "Client credentials grant only allowed for confidential applications")
	}

	// Ensure that the scope is allowed for the application
	scopes := strings.Split(tokenRequest.Scope, " ")
	for _, scope := range scopes {
		if !slices.Contains(application.AllowedScopes, scope) {
			return nil, NewOAuth2Error(oauth2.ErrorInvalidScope, "Invalid scope "+scope)
		}
	}

	// Client authentication has already been validated by the ProcessTokenRequest
	// so we can directly generate the token response
	tokenResponse, err := s.generateTokenResponseForClientCredentialsGrant(application, scopes)
	if err != nil {
		return nil, NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not generate token response")
	}

	return tokenResponse, nil

}

func (s *OAuth2Service) generateTokenResponseForClientCredentialsGrant(application *model.Application, scopes []string) (*oauth2.Oauth2TokenResponse, *oauth2.OAuth2Error) {

	accessToken, expiresIn, scope, tokenType, err := s.generateAccessTokenForClientCredentialsGrant(application, scopes)
	if err != nil {
		return nil, NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not generate token response")
	}

	return &oauth2.Oauth2TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
		Scope:       scope,
		TokenType:   tokenType,
	}, nil
}

func (s *OAuth2Service) processTokenRequestForRefreshTokenGrant(tenant string, realm string, tokenRequest *oauth2.Oauth2TokenRequest, clientAuthentication *oauth2.Oauth2ClientAuthentication, application *model.Application) (*oauth2.Oauth2TokenResponse, *oauth2.OAuth2Error) {

	// Load the refresh token session
	session, err := GetServices().SessionsService.LoadAndDeleteRefreshTokenSession(context.Background(), tenant, realm, tokenRequest.RefreshToken)
	if err != nil {
		return nil, NewOAuth2Error(oauth2.ErrorAccessDenied, "Invalid refresh token")
	}

	// Check if the session is valid
	if session == nil {
		return nil, NewOAuth2Error(oauth2.ErrorInvalidRequest, "Invalid refresh token")
	}

	// Check if the token request is from the same client as the client session
	if session.ClientID != tokenRequest.ClientID {
		return nil, NewOAuth2Error(oauth2.ErrorInvalidGrant, "Invalid refresh token")
	}

	// Issue new access token and new refresh token
	tokenResponse, err := s.generateTokenResponse(session, nil, application, oauth2.Oauth2_RefreshToken)
	if err != nil {
		return nil, NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not generate token response")
	}

	return tokenResponse, nil
}

func (s *OAuth2Service) processTokenRequestForAuthorizationCodeGrant(tenant string, realm string, tokenRequest *oauth2.Oauth2TokenRequest, clientAuthentication *oauth2.Oauth2ClientAuthentication, application *model.Application) (*oauth2.Oauth2TokenResponse, *oauth2.OAuth2Error) {
	session, loginSession, err := GetServices().SessionsService.LoadAndDeleteAuthCodeSession(context.Background(), tenant, realm, tokenRequest.Code)
	if err != nil {
		return nil, NewOAuth2Error(oauth2.ErrorInvalidGrant, "Invalid authorization code")
	}

	// Check if the session is valid
	if session == nil {
		return nil, NewOAuth2Error(oauth2.ErrorInvalidGrant, "Invalid authorization code")
	}

	if loginSession == nil {
		return nil, NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not load login session")
	}

	if tenant != session.Tenant || realm != session.Realm || clientAuthentication.ClientID != session.ClientID {
		return nil, NewOAuth2Error(oauth2.ErrorInvalidGrant, "Invalid authorization code")
	}

	if !application.Confidential {

		// Else if the client is public we need to check if the code verifier is correct
		if tokenRequest.CodeVerifier == "" {
			return nil, NewOAuth2Error(oauth2.ErrorInvalidRequest, "Code verifier is required for public clients")
		}

		// Check if the code verifier is correct
		valid, err := oauth2.VerifyCodeChallenge(tokenRequest.CodeVerifier, session.CodeChallenge)
		if err != nil {
			return nil, NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not verify code verifier")
		}

		if !valid {
			return nil, NewOAuth2Error(oauth2.ErrorInvalidRequest, "Invalid code verifier")
		}
	}

	tokenResponse, err := s.generateTokenResponse(session, loginSession, application, oauth2.Oauth2_AuthorizationCode)
	if err != nil {
		return nil, NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not generate token response")
	}

	return tokenResponse, nil
}

func (s *OAuth2Service) generateTokenResponse(session *model.ClientSession, loginSession *model.AuthenticationSession, application *model.Application, grantType oauth2.OAuth2GrantType) (*oauth2.Oauth2TokenResponse, error) {

	// first we generate the access token
	accessToken, expiresIn, scopes, tokenType, err := s.generateAccessToken(session, loginSession, application, session.Claims)

	if err != nil {
		return nil, fmt.Errorf("internal server error. Could not generate access token: %w", err)
	}

	// if the appliaction as refresh_token grant enabled we need to generate a refresh token
	var refreshToken string
	if slices.Contains(application.AllowedGrants, string(oauth2.Oauth2_RefreshToken)) {
		refreshToken, err = s.generateRefreshToken(session, loginSession, application, session.Claims)
		if err != nil {
			return nil, fmt.Errorf("internal server error. Could not generate refresh token: %w", err)
		}
	}

	// If this is a oidc flow we need to generate an id token by checking the scopes from the session
	// Only during the authorization code flow we need to generate an id token
	var idToken string
	if slices.Contains(application.AllowedScopes, "openid") && grantType == oauth2.Oauth2_AuthorizationCode {

		// Check that the login session is not nil
		if loginSession == nil {
			return nil, fmt.Errorf("internal server error. Login session is nil but an id token is generated")
		}

		var err error
		idToken, err = s.generateIdToken(session, loginSession, application, session.Claims)
		if err != nil {
			// Log the error but continue with the response
			// The ID token will be empty in this case
			fmt.Printf("Failed to generate ID token: %v\n", err)
		}
	}

	tokenResponse := oauth2.Oauth2TokenResponse{
		AccessToken:  accessToken,
		IDToken:      idToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Scope:        scopes,
		TokenType:    tokenType,
	}

	return &tokenResponse, nil
}

func (s *OAuth2Service) generateIdToken(session *model.ClientSession, loginSession *model.AuthenticationSession, application *model.Application, userClaims map[string]interface{}) (string, error) {

	// Load the user from the login session
	if loginSession.User == nil {
		return "", fmt.Errorf("internal server error. User is nil")
	}

	otherClaims, err := s.GetOtherJwtClaims(session.Tenant, session.Realm, session.ClientID)
	if err != nil {
		return "", fmt.Errorf("internal server error. Could not get other claims")
	}

	// if we have a nonce in the login session we add it to the claims
	if loginSession.Oauth2SessionInformation.AuthorizeRequest.Nonce != "" {
		otherClaims["nonce"] = loginSession.Oauth2SessionInformation.AuthorizeRequest.Nonce
	}

	// if we have a auth_time in the login session we add it to the claims
	if loginSession.Result.AuthLevel != "" {
		otherClaims["acr"] = loginSession.Result.AuthLevel
	}

	// Merge the claims into the final set
	claims := maps.Clone(userClaims)
	for k, v := range otherClaims {
		claims[k] = v
	}

	// Sign the token using the JWT service
	token, err := GetServices().JWTService.SignJWT(session.Tenant, session.Realm, claims)
	if err != nil {
		return "", fmt.Errorf("internal server error. Could not sign token: %w", err)
	}

	return token, nil
}

func (s *OAuth2Service) generateAccessToken(session *model.ClientSession, loginSession *model.AuthenticationSession, application *model.Application, userClaims map[string]interface{}) (string, int, string, string, error) {

	// First we generate the access token
	expiresIn := application.AccessTokenLifetime
	scopes := session.Scope
	tokenType := "Bearer"
	tenant := session.Tenant
	realm := session.Realm

	scopesArray := strings.Split(scopes, " ")

	// Then we store it into the client sessions database using the service
	accessToken, _, err := GetServices().SessionsService.CreateAccessTokenSession(context.Background(), tenant, realm, session.ClientID, session.UserID, scopesArray, "authorization_code", expiresIn, userClaims)

	if err != nil {
		return "", 0, "", "", fmt.Errorf("internal server error. Could not create access token session: %w", err)
	}

	return accessToken, expiresIn, scopes, tokenType, nil
}

// Compared to the generateAccessToken this has no associated user, just an appliaction
func (s *OAuth2Service) generateAccessTokenForClientCredentialsGrant(application *model.Application, scopes []string) (string, int, string, string, error) {

	// First we generate the access token
	expiresIn := application.AccessTokenLifetime
	tokenType := "Bearer"
	tenant := application.Tenant
	realm := application.Realm
	clientId := application.ClientId
	userId := ""
	scope := strings.Join(scopes, " ")

	// Then we store it into the client sessions database using the service
	accessToken, _, err := GetServices().SessionsService.CreateAccessTokenSession(context.Background(), tenant, realm, clientId, userId, scopes, string(oauth2.Oauth2_ClientCredentials), expiresIn, nil)

	if err != nil {
		return "", 0, "", "", fmt.Errorf("internal server error. Could not create access token session: %w", err)
	}

	return accessToken, expiresIn, scope, tokenType, nil
}

func (s *OAuth2Service) generateRefreshToken(session *model.ClientSession, loginSession *model.AuthenticationSession, application *model.Application, userClaims map[string]interface{}) (string, error) {

	expiresIn := application.RefreshTokenLifetime

	if expiresIn == 0 {
		expiresIn = 60 * 60 * 24 * 365 * 100 // 100 years
	}

	scopes := session.Scope
	tenant := session.Tenant
	realm := session.Realm

	scopesArray := strings.Split(scopes, " ")

	// Create the refresh token
	refreshToken, _, err := GetServices().SessionsService.CreateRefreshTokenSession(context.Background(), tenant, realm, session.ClientID, session.UserID, scopesArray, "authorization_code", expiresIn, userClaims)

	if err != nil {
		return "", fmt.Errorf("internal server error. Could not create refresh token session: %w", err)
	}

	return refreshToken, nil
}

func (s *OAuth2Service) GetUserClaims(user model.User, scope string, oauth2Session *model.Oauth2Session) (map[string]interface{}, error) {

	// If userid is empty we return an error. This might be the case if a client uses the client_credentials grant and then accesses the userinfo endpoint
	if user.ID == "" {
		return nil, fmt.Errorf("internal server error. User ID is empty")
	}

	// now we map the user attributes into claims
	// we need to check the sesssion scopes and map the attributes accordingly
	claims := make(map[string]interface{})
	//scopes := strings.Split(scope, " ")

	// We always return the sub claim
	claims["sub"] = user.ID

	if oauth2Session != nil && !oauth2Session.AuthTime.IsZero() {
		claims["auth_time"] = oauth2Session.AuthTime.Unix()
	}

	//TODO we need to add the claims for the user attributes
	/*

		if slices.Contains(scopes, "email") {
			claims["email"] = user.Email
			claims["email_verified"] = user.EmailVerified
		}

		if slices.Contains(scopes, "profile") {
			claims["username"] = user.Username
			claims["name"] = user.DisplayName
			claims["given_name"] = user.GivenName
			claims["family_name"] = user.FamilyName
		}

		if slices.Contains(scopes, "phone") {
			claims["phone"] = user.Phone
			claims["phone_verified"] = user.PhoneVerified
		}

		if slices.Contains(scopes, "groups") {
			claims["groups"] = user.Groups
		}

		if slices.Contains(scopes, "roles") {
			claims["roles"] = user.Roles
		}*/

	return claims, nil
}

func (s *OAuth2Service) GetOtherJwtClaims(tenant, realm, client_id string) (map[string]interface{}, error) {

	// Issuer
	claims := make(map[string]interface{})

	// We need to load the realm to load the signing key
	loadedRealm, ok := GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		return nil, fmt.Errorf("internal server error. Could not get realm")
	}

	// Load the application
	application, ok := GetServices().ApplicationService.GetApplication(tenant, realm, client_id)
	if !ok {
		return nil, fmt.Errorf("internal server error. Could not get application")
	}

	now := time.Now()

	claims["iss"] = loadedRealm.Config.BaseUrl
	claims["aud"] = application.ClientId
	claims["exp"] = now.Add(time.Duration(application.AccessTokenLifetime) * time.Second).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	claims["jti"] = lib.GenerateSecureSessionID()

	return claims, nil
}

// ToQueryString converts the AuthorizationResponse to a URL query string
func (s *OAuth2Service) ToQueryString(response *oauth2.AuthorizationResponse) string {
	params := url.Values{}
	params.Add("code", response.Code)
	if response.State != "" {
		params.Add("state", response.State)
	}
	if response.Iss != "" {
		params.Add("iss", response.Iss)
	}
	return params.Encode()
}

// IntrospectAccessToken introspects an OAuth2 access token and returns information about it
func (s *OAuth2Service) IntrospectAccessToken(tenant, realm string, tokenIntrospectionRequest *oauth2.TokenIntrospectionRequest) (*oauth2.TokenIntrospectionResponse, *oauth2.OAuth2Error) {

	// Load the session from the token
	session, err := GetServices().SessionsService.GetClientSessionByAccessToken(context.Background(), tenant, realm, tokenIntrospectionRequest.Token)
	if err != nil {
		return nil, NewOAuth2Error(oauth2.ErrorServerError, "internal server error. Could not get client session")
	}

	// If no session found, token is not active
	if session == nil {
		return &oauth2.TokenIntrospectionResponse{Active: false}, nil
	}

	loadedRealm, ok := GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		return nil, NewOAuth2Error(oauth2.ErrorServerError, "internal server error. Could not get realm")
	}

	issuer := loadedRealm.Config.BaseUrl

	// Create the introspection response
	response := &oauth2.TokenIntrospectionResponse{
		Active:    true,
		Scope:     session.Scope,
		ClientID:  session.ClientID,
		TokenType: "Bearer",
		Exp:       session.Expire.Unix(),
		Iat:       session.Created.Unix(),
		Nbf:       session.Created.Unix(),
		Aud:       session.ClientID,
		Iss:       issuer,
		Jti:       session.ClientSessionID,
	}

	// If we have a user ID, add user-related fields
	if session.UserID != "" {
		user, err := GetServices().UserService.GetUserByID(context.Background(), tenant, realm, session.UserID)

		if err != nil {
			return nil, NewOAuth2Error(oauth2.ErrorServerError, "internal server error. Could not get user")
		}

		if user == nil {
			return nil, NewOAuth2Error(oauth2.ErrorServerError, "internal server error. Invalid token")
		}

		response.Sub = user.ID

	}

	return response, nil
}

func (s *OAuth2Service) ValidateRedirectUri(oauth2request *model.AuthorizeRequest, application *model.Application) *oauth2.OAuth2Error {

	// If the compatibility redirect uri prefix check is enabled we need to check if the redirect uri is a prefix of the allowed redirect uris
	if application != nil && application.Settings != nil && application.Settings.OAuth2Settings != nil && application.Settings.OAuth2Settings.CompatibilityRedirectUriPrefixCheck {
		for _, redirectUri := range application.RedirectUris {
			if strings.HasPrefix(oauth2request.RedirectURI, redirectUri) {
				return nil
			}
		}
	}

	// validate if the redirect_uir is in the list of allowed redirect uris
	if oauth2request.RedirectURI != "" && !slices.Contains(application.RedirectUris, oauth2request.RedirectURI) {
		return oauth2.NewOAuth2Error(oauth2.ErrorInvalidRequest, "Invalid redirect URI")
	}

	return nil
}
