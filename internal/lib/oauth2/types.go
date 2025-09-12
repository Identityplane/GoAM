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

// Oauth2TokenRequest represents an OAuth2 token request
type Oauth2TokenRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	ClientID     string `json:"client_id"`
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	RedirectURI  string `json:"redirect_uri"`
	Scope        string `json:"scope"` // Only used for the client credentials grant
}

// Oauth2ClientAuthentication represents OAuth2 client authentication
type Oauth2ClientAuthentication struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// Oauth2TokenResponse represents an OAuth2 token response
type Oauth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
}

// TokenIntrospectionRequest represents the request to the introspection endpoint
type TokenIntrospectionRequest struct {
	Token         string `json:"token"`
	TokenTypeHint string `json:"token_type_hint,omitempty"`
}

// TokenIntrospectionResponse represents the response from the introspection endpoint
type TokenIntrospectionResponse struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
	Nbf       int64  `json:"nbf,omitempty"`
	Sub       string `json:"sub,omitempty"`
	Aud       string `json:"aud,omitempty"`
	Iss       string `json:"iss,omitempty"`
	Jti       string `json:"jti,omitempty"`
}
