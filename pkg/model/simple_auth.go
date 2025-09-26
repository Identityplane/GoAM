package model

const (
	GRANT_SIMPLE_AUTH_BODY   = "simple-body"
	GRANT_SIMPLE_AUTH_COOKIE = "simple-cookie"
)

type SimpleAuthRequest struct {
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
	Grant        string `json:"grant"`
	ResponseType string `json:"response_type"`
}

type SimpleAuthContext struct {
	Request     *SimpleAuthRequest `json:"request"`
	RedirectURI string             `json:"redirect_uri"`
}

type SimpleAuthResponse struct {
	Success               bool                   `json:"success"`
	AccessToken           string                 `json:"access_token,omitempty"`
	TokenType             string                 `json:"token_type,omitempty"`
	RefreshToken          string                 `json:"refresh_token,omitempty"`
	UserClaims            map[string]interface{} `json:"user,omitempty"`
	ExpiresIn             int                    `json:"expires_in,omitempty"`
	RefreshTokenExpiresIn int                    `json:"refresh_token_expires_in,omitempty"`
	Scope                 string                 `json:"scope,omitempty"`
	Error                 string                 `json:"error,omitempty"`
}

type SimpleAuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
