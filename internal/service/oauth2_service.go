package service

import (
	"context"
	"goiam/internal/auth/graph"
	"goiam/internal/lib/oauth2"
	"goiam/internal/model"
	"net/url"
	"slices"
)

// OAuth2Service handles OAuth2 related operations
type OAuth2Service struct {
}

// NewOAuth2Service creates a new OAuth2Service instance
func NewOAuth2Service() *OAuth2Service {
	return &OAuth2Service{}
}

// Enum for the different OAuth2 grant types, we differenciate between authorization_code and authorization_code_pkce
type OAuth2GrantType string

// Valid OAuth2 grant types
const (
	Oauth2_AuthorizationCode     OAuth2GrantType = "authorization_code"
	Oauth2_AuthorizationCodePKCE OAuth2GrantType = "authorization_code_pkce"
	Oauth2_ClientCredentials     OAuth2GrantType = "client_credentials"
	Oauth2_RefreshToken          OAuth2GrantType = "refresh_token"
	Oauth2_InvalidFlow           OAuth2GrantType = "invalid"
)

// Valid OAuth2 error codes as defined in RFC 6749
const (
	ErrorInvalidRequest          = "invalid_request"
	ErrorUnauthorizedClient      = "unauthorized_client"
	ErrorAccessDenied            = "access_denied"
	ErrorUnsupportedResponseType = "unsupported_response_type"
	ErrorInvalidScope            = "invalid_scope"
	ErrorServerError             = "server_error"
	ErrorTemporarilyUnavailable  = "temporarily_unavailable"
)

// AuthorizationResponse represents the OAuth2 authorization response
type AuthorizationResponse struct {
	Code  string `json:"code"`  // REQUIRED. The authorization code
	State string `json:"state"` // REQUIRED if state was present in the request
	Iss   string `json:"iss"`   // OPTIONAL. The identifier of the authorization server
}

// OAuth2Error represents an OAuth2 error response
type OAuth2Error struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

func NewOAuth2Error(errorCode string, errorDescription string) *OAuth2Error {
	errorResponse := OAuth2Error{
		Error:            errorCode,
		ErrorDescription: errorDescription,
	}
	return &errorResponse
}

// ValidateOAuth2AuthorizationRequest validates the OAuth2 authorization request
func (s *OAuth2Service) ValidateOAuth2AuthorizationRequest(oauth2request *model.AuthorizeRequest, tenant, realm string, application *model.Application, flowId string) *oauth2.OAuth2Error {

	// validate if the redirect_uir is in the list of allowed redirect uris
	if oauth2request.RedirectURI != "" && !slices.Contains(application.RedirectUris, oauth2request.RedirectURI) {
		return oauth2.NewOAuth2Error(oauth2.ErrorInvalidRequest, "Invalid redirect URI")
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

	// Too lookup the result node type we need to get the floew
	flow, ok := GetServices().FlowService.GetFlowById(tenant, realm, session.FlowId)
	if !ok {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Flow not found")
	}

	// Get the type of the current node
	currentNode, ok := flow.Definition.Nodes[session.Current]
	if !ok {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Current node not found")
	}

	// If the result node is a failure result we return an oauth2 error
	if currentNode.Use == graph.FailureResultNode.Name {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorAccessDenied, "Authentication Failed")
	}

	if currentNode.Use != graph.SuccessResultNode.Name {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Unexpected result node")
	}

	// In order to set a access token the result auth level must be at least 1
	/*
		if session.Result.AuthLevel == "" || session.Result.AuthLevel == model.AuthLevelUnauthenticated {
			return nil, oauth2.NewOAuth2Error(oauth2.ErrorAccessDenied, "Authentication level unauthenticated")
		}*/

	// If all ok we create a client session and issue an auth code
	scope := session.Oauth2SessionInformation.AuthorizeRequest.Scope
	authCode, err := GetServices().SessionsService.CreateAuthCodeSession(
		context.Background(),
		tenant,
		realm,
		session.Oauth2SessionInformation.AuthorizeRequest.ClientID,
		session.Result.UserID,
		scope,
		"authorization_code")

	if err != nil {
		return nil, oauth2.NewOAuth2Error(oauth2.ErrorServerError, "Internal server error. Could not create session")
	}

	// Create the authorization response
	response := oauth2.AuthorizationResponse{
		Code:  authCode,
		State: session.Oauth2SessionInformation.AuthorizeRequest.State,
		Iss:   session.LoginUri,
	}

	return &response, nil
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
