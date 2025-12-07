package model

import "time"

// AccessTokenType represents the type of access token
type AccessTokenType string

const (
	// AccessTokenTypeJWT represents a JWT-based access token
	AccessTokenTypeJWT AccessTokenType = "jwt"
	// AccessTokenTypeSessionKey represents a session key-based access token
	AccessTokenTypeSessionKey AccessTokenType = "session"
)

// Application represents an OAuth2 client application
type Application struct {
	Tenant                     string          `json:"tenant" yaml:"tenant" db:"tenant"`
	Realm                      string          `json:"realm" yaml:"realm" db:"realm"`
	ClientId                   string          `json:"client_id" yaml:"client_id" db:"client_id"`
	ClientSecret               string          `json:"-" yaml:"client_secret" db:"client_secret"` // Only in yaml as this is used to load from static configuration
	Confidential               bool            `json:"confidential" yaml:"confidential" db:"confidential"`
	ConsentRequired            bool            `json:"consent_required" yaml:"consent_required" db:"consent_required"`
	Description                string          `json:"description" yaml:"description" db:"description"`
	AllowedScopes              []string        `json:"allowed_scopes" yaml:"allowed_scopes" db:"allowed_scopes"`
	AllowedGrants              []string        `json:"allowed_grants" yaml:"allowed_grants" db:"allowed_grants"`
	AllowedAuthenticationFlows []string        `json:"allowed_authentication_flows" yaml:"allowed_authentication_flows" db:"allowed_authentication_flows"`
	AccessTokenLifetime        int             `json:"access_token_lifetime" yaml:"access_token_lifetime" db:"access_token_lifetime"`
	RefreshTokenLifetime       int             `json:"refresh_token_lifetime" yaml:"refresh_token_lifetime" db:"refresh_token_lifetime"`
	IdTokenLifetime            int             `json:"id_token_lifetime" yaml:"id_token_lifetime" db:"id_token_lifetime"`
	AccessTokenType            AccessTokenType `json:"access_token_type" yaml:"access_token_type" db:"access_token_type"`
	AccessTokenAlgorithm       string          `json:"access_token_algorithm" yaml:"access_token_algorithm" db:"access_token_algorithm"`
	AccessTokenMapping         string          `json:"access_token_mapping" yaml:"access_token_mapping" db:"access_token_mapping"`
	IdTokenAlgorithm           string          `json:"id_token_algorithm" yaml:"id_token_algorithm" db:"id_token_algorithm"`
	IdTokenMapping             string          `json:"id_token_mapping" yaml:"id_token_mapping" db:"id_token_mapping"`
	RedirectUris               []string        `json:"redirect_uris" yaml:"redirect_uris" db:"redirect_uris"`
	CreatedAt                  time.Time       `json:"created_at" yaml:"created_at" db:"created_at"`
	UpdatedAt                  time.Time       `json:"updated_at" yaml:"updated_at" db:"updated_at"`

	// @Description: For 1. party authentication the access token can be returned as cookie
	// The expiry of the cookie will be set via the access token lifetime
	Settings *ApplicationExtensionSettings `json:"settings,omitempty" yaml:"settings,omitempty" db:"settings"`
}

type ApplicationExtensionSettings struct {
	Cookie         *CookieSpecification `json:"cookie_specification,omitempty" yaml:"cookie_specification,omitempty" db:"cookie_specification"`
	OAuth2Settings *OAuth2Settings      `json:"oauth2_settings,omitempty" yaml:"oauth2_settings,omitempty" db:"oauth2_settings"`
}

type CookieSpecification struct {
	Name          string `json:"name" yaml:"name"`
	Domain        string `json:"domain" yaml:"domain"`
	Path          string `json:"path" yaml:"path"`
	Secure        bool   `json:"secure" yaml:"secure"`
	HttpOnly      bool   `json:"http_only" yaml:"http_only"`
	SameSite      string `json:"same_site" yaml:"same_site"`
	SessionExpiry bool   `json:"session_expiry" yaml:"session_expiry"` // If true the cookie will be set without max-age and expires when the browser window is closed
}

type OAuth2Settings struct {
	CompatibilityRedirectUriPrefixCheck bool `json:"compatibility_redirect_uri_prefix_check" yaml:"compatibility_redirect_uri_prefix_check"` // Enables prefix check for oauth2.0 compatibility. OAuth2.1 does not support this and requires exact match.
	LoadUserFromLoginSession            bool `json:"load_user_from_login_session" yaml:"load_user_from_login_session"`                       // Enables loading the user from the login session instead of the database
}
