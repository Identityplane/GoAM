package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const applicationYaml = `management-ui:
  confidential: false
  consent_required: false
  description: Management UI Application
  client_secret: secret
  allowed_scopes:
    - openid
    - profile
    - write:user
    - write:flows
    - write:realms
    - write:applications
  allowed_grants:
    - authorization_code_pkce
  redirect_uris:
    - http://localhost:3000
  allowed_authentication_flows:
    - login_or_register
  access_token_lifetime: 600
  refresh_token_lifetime: 3600
  id_token_lifetime: 600
  access_token_type: session
`

func TestApplicationDeserialization(t *testing.T) {
	// Create a map to hold the YAML data
	var data map[string]Application

	// Unmarshal the YAML
	err := yaml.Unmarshal([]byte(applicationYaml), &data)
	assert.NoError(t, err, "Failed to unmarshal application YAML")

	// Get the application from the map
	app, exists := data["management-ui"]
	assert.True(t, exists, "Application 'management-ui' not found in YAML")

	// Verify all fields
	assert.False(t, app.Confidential, "Confidential should be false")
	assert.False(t, app.ConsentRequired, "ConsentRequired should be false")
	assert.Equal(t, "Management UI Application", app.Description, "Description mismatch")

	// Verify allowed scopes
	expectedScopes := []string{
		"openid",
		"profile",
		"write:user",
		"write:flows",
		"write:realms",
		"write:applications",
	}
	assert.Equal(t, expectedScopes, app.AllowedScopes, "AllowedScopes mismatch")

	// Verify allowed grants
	expectedGrants := []string{"authorization_code_pkce"}
	assert.Equal(t, expectedGrants, app.AllowedGrants, "AllowedGrants mismatch")

	// Verify redirect URIs
	expectedRedirectURIs := []string{"http://localhost:3000"}
	assert.Equal(t, expectedRedirectURIs, app.RedirectUris, "RedirectUris mismatch")

	// Verify allowed authentication flows
	expectedFlows := []string{"login_or_register"}
	assert.Equal(t, expectedFlows, app.AllowedAuthenticationFlows, "AllowedAuthenticationFlows mismatch")

	// Verify token lifetimes
	assert.Equal(t, 600, app.AccessTokenLifetime, "AccessTokenLifetime mismatch")
	assert.Equal(t, 3600, app.RefreshTokenLifetime, "RefreshTokenLifetime mismatch")
	assert.Equal(t, 600, app.IdTokenLifetime, "IdTokenLifetime mismatch")
	assert.Equal(t, "secret", app.ClientSecret, "ClientSecret mismatch")

	// Verify access token type
	assert.Equal(t, AccessTokenTypeSessionKey, app.AccessTokenType, "AccessTokenType mismatch")
}
