package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Identityplane/GoAM/internal/lib/oauth2"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

type simpleAuthService struct {
}

// NewSimpleAuthService creates a new SimpleAuthService instance
func NewSimpleAuthService() services_interface.SimpleAuthService {
	return &simpleAuthService{}
}

// Simple auth is a simplified version of OAuth2 for first party flows
// It is used to authenticate users without in first party contexts like mobile apps, first party websites on the same domain

// A flow consists of the following information
// ClientId: the application, this sets fields like allowed flows, as well as duration of access and refresh tokens
// RedirectURI: the redirect URI, this is used to redirect the user back to the application. For mobile apps etc this can be omitted
// ResoponseType: The application can request directly the following tokens:
// - token: access token
// - refresh_token: a refresh token
// - id_token: an id token

// GRANT: The following additional grant types are available:
// simple-body: access and refresh token is directly returned
// simple-cookie: access token is directly returned in a cookie

// REDIRECT:
// Even in a first party scenario it makes sense to redirect the user back to the applicaiton. For this a redirect uri can be used in the same way
// as with Oauth2. The uri is also strictly validated agains the allowed redirect uris of the application.

func (s *simpleAuthService) VerifySimpleAuthFlowRequest(ctx context.Context, req *model.SimpleAuthRequest, application *model.Application, flow *model.Flow) error {

	// Check if the redirect URI is allowed for the application
	if req.RedirectURI != "" {
		if !slices.Contains(application.RedirectUris, req.RedirectURI) {
			return errors.New("invalid redirect URI")
		}
	}

	// Check if the grant type is either simple-body or simple-cookie
	if req.Grant != model.GRANT_SIMPLE_AUTH_BODY && req.Grant != model.GRANT_SIMPLE_AUTH_COOKIE {
		return errors.New("invalid grant type")
	}

	// Check if the grant type is allowed
	if !slices.Contains(application.AllowedGrants, req.Grant) {
		return errors.New("grant type not allowed")
	}

	// Check if the scopes are allowed
	scopes := strings.Split(req.Scope, " ")
	for _, scope := range scopes {
		if scope != "" && !slices.Contains(application.AllowedScopes, scope) {
			return errors.New("scope not allowed")
		}
	}

	// Check if the flow is allowed for the application
	flowAllowed := false
	for _, allowedFlow := range application.AllowedAuthenticationFlows {
		if allowedFlow == "*" || allowedFlow == flow.Id {
			flowAllowed = true
			break
		}
	}

	if !flowAllowed {
		return errors.New("flow not allowed")
	}

	return nil
}

func (s *simpleAuthService) FinishSimpleAuthFlow(ctx context.Context, session *model.AuthenticationSession, tenant, realm string) (*model.SimpleAuthResponse, *model.SimpleAuthError) {

	// If there is no oauth2 session information we return an error
	if session.SimpleAuthSessionInformation == nil || session.SimpleAuthSessionInformation.Request == nil {
		return nil, returnSimpleAuthError(oauth2.ErrorServerError, "Internal server error. No oauth2 session information")
	}

	// if there is an error we return an error
	if session.DidResultError() {
		return nil, returnSimpleAuthError(oauth2.ErrorServerError, "Internal server error. Unexpected result node")
	}

	// if the result node is a failure result we return an error
	if !session.DidResultAuthenticated() {
		return nil, returnSimpleAuthError(oauth2.ErrorAccessDenied, "Authentication Failed")
	}

	// Load the application
	application, ok := GetServices().ApplicationService.GetApplication(tenant, realm, session.SimpleAuthSessionInformation.Request.ClientID)
	if !ok {
		return nil, returnSimpleAuthError(oauth2.ErrorServerError, "Internal server error. Could not get application")
	}

	// In the simple auth flow we dont do any token exchange but directly return the tokens
	request := session.SimpleAuthSessionInformation.Request
	userID := session.Result.UserID
	scope := request.Scope

	var accessToken string
	var tokenType string
	var refreshToken string
	var expiresIn int
	var refreshTokenExpiresIn int
	var err error

	// We always create a access token
	accessToken, expiresIn, scope, tokenType, err = s.generateAccessToken(request, application, userID)
	if err != nil {
		return nil, returnSimpleAuthError(oauth2.ErrorServerError, "Internal server error. Could not generate access token")
	}

	// If the application has refresh token grant enabled we generate a refresh token
	if slices.Contains(application.AllowedGrants, string(oauth2.Oauth2_RefreshToken)) {
		refreshToken, refreshTokenExpiresIn, _, err = s.generateRefreshToken(request, application, userID)
		if err != nil {
			return nil, returnSimpleAuthError(oauth2.ErrorServerError, "Internal server error. Could not generate refresh token")
		}
	}

	// Get the user claims
	userClaims, err := GetServices().OAuth2Service.GetUserClaims(*session.User, scope)
	if err != nil {
		return nil, returnSimpleAuthError(oauth2.ErrorServerError, "Internal server error. Could not get user claims")
	}

	// if the result node is a success result we return the tokens
	if session.DidResultAuthenticated() {
		return &model.SimpleAuthResponse{
			AccessToken:           accessToken,
			TokenType:             tokenType,
			ExpiresIn:             expiresIn,
			RefreshToken:          refreshToken,
			RefreshTokenExpiresIn: refreshTokenExpiresIn,
			Scope:                 scope,
			UserClaims:            userClaims,
		}, nil
	}

	return nil, nil

}

func (s *simpleAuthService) generateAccessToken(request *model.SimpleAuthRequest, application *model.Application, userID string) (string, int, string, string, error) {

	// First we generate the access token
	expiresIn := application.AccessTokenLifetime
	scopes := request.Scope
	tokenType := "Bearer"
	tenant := application.Tenant
	realm := application.Realm

	scopesArray := strings.Split(scopes, " ")

	// Then we store it into the client sessions database using the service
	accessToken, _, err := GetServices().SessionsService.CreateAccessTokenSession(context.Background(), tenant, realm, request.ClientID, userID, scopesArray, "authorization_code", expiresIn)

	if err != nil {
		return "", 0, "", "", fmt.Errorf("internal server error. Could not create access token session: %w", err)
	}

	return accessToken, expiresIn, scopes, tokenType, nil
}

func (s *simpleAuthService) generateRefreshToken(request *model.SimpleAuthRequest, application *model.Application, userID string) (string, int, string, error) {

	expiresIn := application.RefreshTokenLifetime

	if expiresIn == 0 {
		expiresIn = 60 * 60 * 24 * 365 * 100 // 100 years
	}

	scopes := request.Scope
	tenant := application.Tenant
	realm := application.Realm

	scopesArray := strings.Split(scopes, " ")

	// Create the refresh token
	refreshToken, _, err := GetServices().SessionsService.CreateRefreshTokenSession(context.Background(), tenant, realm, request.ClientID, userID, scopesArray, "authorization_code", expiresIn)

	if err != nil {
		return "", 0, "", fmt.Errorf("internal server error. Could not create refresh token session: %w", err)
	}

	return refreshToken, expiresIn, scopes, nil
}

func returnSimpleAuthError(err, errorDescription string) *model.SimpleAuthError {
	return &model.SimpleAuthError{
		Error:            err,
		ErrorDescription: errorDescription,
	}
}
