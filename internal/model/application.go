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
	Tenant                     string          `json:"tenant"`
	Realm                      string          `json:"realm"`
	ClientId                   string          `json:"client_id"`
	ClientSecret               string          `json:"-"`
	Confidential               bool            `json:"confidential"`
	ConsentRequired            bool            `json:"consent_required"`
	Description                string          `json:"description"`
	AllowedScopes              []string        `json:"allowed_scopes"`
	AllowedGrants              []string        `json:"allowed_grants"`
	AllowedAuthenticationFlows []string        `json:"allowed_authentication_flows"`
	AccessTokenLifetime        int             `json:"access_token_lifetime"`
	RefreshTokenLifetime       int             `json:"refresh_token_lifetime"`
	IdTokenLifetime            int             `json:"id_token_lifetime"`
	AccessTokenType            AccessTokenType `json:"access_token_type"`
	AccessTokenAlgorithm       string          `json:"access_token_algorithm"`
	AccessTokenMapping         string          `json:"access_token_mapping"`
	IdTokenAlgorithm           string          `json:"id_token_algorithm"`
	IdTokenMapping             string          `json:"id_token_mapping"`
	RedirectUris               []string        `json:"redirect_uris"`
	CreatedAt                  time.Time       `json:"created_at"`
	UpdatedAt                  time.Time       `json:"updated_at"`
}
