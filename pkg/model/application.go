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
	Tenant                     string          `json:"tenant" yaml:"tenant"`
	Realm                      string          `json:"realm" yaml:"realm"`
	ClientId                   string          `json:"client_id" yaml:"client_id"`
	ClientSecret               string          `json:"-" yaml:"client_secret"` // Only in yaml as this is used to load from static configuration
	Confidential               bool            `json:"confidential" yaml:"confidential"`
	ConsentRequired            bool            `json:"consent_required" yaml:"consent_required"`
	Description                string          `json:"description" yaml:"description"`
	AllowedScopes              []string        `json:"allowed_scopes" yaml:"allowed_scopes"`
	AllowedGrants              []string        `json:"allowed_grants" yaml:"allowed_grants"`
	AllowedAuthenticationFlows []string        `json:"allowed_authentication_flows" yaml:"allowed_authentication_flows"`
	AccessTokenLifetime        int             `json:"access_token_lifetime" yaml:"access_token_lifetime"`
	RefreshTokenLifetime       int             `json:"refresh_token_lifetime" yaml:"refresh_token_lifetime"`
	IdTokenLifetime            int             `json:"id_token_lifetime" yaml:"id_token_lifetime"`
	AccessTokenType            AccessTokenType `json:"access_token_type" yaml:"access_token_type"`
	AccessTokenAlgorithm       string          `json:"access_token_algorithm" yaml:"access_token_algorithm"`
	AccessTokenMapping         string          `json:"access_token_mapping" yaml:"access_token_mapping"`
	IdTokenAlgorithm           string          `json:"id_token_algorithm" yaml:"id_token_algorithm"`
	IdTokenMapping             string          `json:"id_token_mapping" yaml:"id_token_mapping"`
	RedirectUris               []string        `json:"redirect_uris" yaml:"redirect_uris"`
	CreatedAt                  time.Time       `json:"created_at" yaml:"created_at"`
	UpdatedAt                  time.Time       `json:"updated_at" yaml:"updated_at"`
}
