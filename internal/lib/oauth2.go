package lib

// Enum for the different OAuth2 grant types, we differenciate between authorization_code and authorization_code_pkce
type OAuth2GrantType string

const (
	Oauth2_AuthorizationCode     OAuth2GrantType = "authorization_code"
	Oauth2_AuthorizationCodePKCE OAuth2GrantType = "authorization_code_pkce"
	Oauth2_ClientCredentials     OAuth2GrantType = "client_credentials"
	Oauth2_RefreshToken          OAuth2GrantType = "refresh_token"
	Oauth2_InvalidFlow           OAuth2GrantType = "invalid"
)
