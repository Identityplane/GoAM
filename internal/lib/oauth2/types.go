package oauth2

// OAuth2GrantType is an enum for the different OAuth2 grant types
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
	ErrorLoginRequired           = "login_required"
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

// NewOAuth2Error creates a new OAuth2 error response
func NewOAuth2Error(errorCode string, errorDescription string) *OAuth2Error {
	errorResponse := OAuth2Error{
		Error:            errorCode,
		ErrorDescription: errorDescription,
	}
	return &errorResponse
}
