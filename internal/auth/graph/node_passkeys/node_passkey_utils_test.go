package node_passkeys

import (
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
)

func TestGetWebAuthnConfig(t *testing.T) {
	tests := []struct {
		name           string
		loginUri       string
		customConfig   map[string]string
		expectedConfig *webauthn.Config
		expectError    bool
	}{
		{
			name:         "localhost with default config",
			loginUri:     "http://localhost:8080/acme/default/auth/login?debug",
			customConfig: map[string]string{},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "localhost",
				RPOrigins:     []string{"http://localhost:8080"},
			},
			expectError: false,
		},
		{
			name:         "identityplane.cloud with default config",
			loginUri:     "https://identityplane.cloud/bitmex/default/auth/login",
			customConfig: map[string]string{},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "identityplane.cloud",
				RPOrigins:     []string{"https://identityplane.cloud"},
			},
			expectError: false,
		},
		{
			name:     "custom rpId override",
			loginUri: "https://identityplane.cloud/bitmex/default/auth/login",
			customConfig: map[string]string{
				"rpId": "custom.domain.com",
			},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "custom.domain.com",
				RPOrigins:     []string{"https://custom.domain.com"},
			},
			expectError: false,
		},
		{
			name:     "custom rpOrigin override",
			loginUri: "https://identityplane.cloud/bitmex/default/auth/login",
			customConfig: map[string]string{
				"rpOrigin": "https://custom.origin.com",
			},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "identityplane.cloud",
				RPOrigins:     []string{"https://custom.origin.com"},
			},
			expectError: false,
		},
		{
			name:     "custom rpDisplayName override",
			loginUri: "https://identityplane.cloud/bitmex/default/auth/login",
			customConfig: map[string]string{
				"rpDisplayName": "Custom App Name",
			},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Custom App Name",
				RPID:          "identityplane.cloud",
				RPOrigins:     []string{"https://identityplane.cloud"},
			},
			expectError: false,
		},
		{
			name:     "all custom configs",
			loginUri: "https://identityplane.cloud/bitmex/default/auth/login",
			customConfig: map[string]string{
				"rpId":          "custom.domain.com",
				"rpOrigin":      "https://custom.origin.com",
				"rpDisplayName": "Custom App Name",
			},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Custom App Name",
				RPID:          "custom.domain.com",
				RPOrigins:     []string{"https://custom.origin.com"},
			},
			expectError: false,
		},
		{
			name:         "subdomain with default config",
			loginUri:     "https://manage.identityplane.cloud/bitmex/default/auth/login",
			customConfig: map[string]string{},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "manage.identityplane.cloud",
				RPOrigins:     []string{"https://manage.identityplane.cloud"},
			},
			expectError: false,
		},
		{
			name:         "http subdomain with port",
			loginUri:     "http://dev.localhost:3000/auth/login",
			customConfig: map[string]string{},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "dev.localhost",
				RPOrigins:     []string{"http://dev.localhost:3000"},
			},
			expectError: false,
		},
		{
			name:         "invalid login uri",
			loginUri:     "not-a-valid-url",
			customConfig: map[string]string{},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "",
				RPOrigins:     []string{"://"},
			},
			expectError: false, // url.Parse doesn't return error for invalid URLs
		},
		{
			name:         "empty login uri",
			loginUri:     "",
			customConfig: map[string]string{},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "",
				RPOrigins:     []string{"://"},
			},
			expectError: false, // url.Parse doesn't return error for empty URLs
		},
		{
			name:         "complex path with query params and port",
			loginUri:     "https://app.example.com:8443/tenant/realm/auth/login?redirect=/dashboard&debug=true",
			customConfig: map[string]string{},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "app.example.com",
				RPOrigins:     []string{"https://app.example.com:8443"},
			},
			expectError: false,
		},
		{
			name:     "custom config with empty values",
			loginUri: "https://identityplane.cloud/bitmex/default/auth/login",
			customConfig: map[string]string{
				"rpId":          "",
				"rpOrigin":      "",
				"rpDisplayName": "",
			},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "identityplane.cloud",
				RPOrigins:     []string{"https://identityplane.cloud"},
			},
			expectError: false,
		},
		{
			name:         "standard port 80",
			loginUri:     "http://example.com:80/auth/login",
			customConfig: map[string]string{},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "example.com",
				RPOrigins:     []string{"http://example.com:80"},
			},
			expectError: false,
		},
		{
			name:         "standard port 443",
			loginUri:     "https://example.com:443/auth/login",
			customConfig: map[string]string{},
			expectedConfig: &webauthn.Config{
				RPDisplayName: "Go IAM",
				RPID:          "example.com",
				RPOrigins:     []string{"https://example.com:443"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test state
			state := &model.AuthenticationSession{
				LoginUri: tt.loginUri,
			}

			// Create test node
			node := &model.GraphNode{
				Name:         "testNode",
				Use:          "testNode",
				CustomConfig: tt.customConfig,
			}

			// Call the function
			config, err := getWebAuthnConfig(state, node)

			// Check error expectations
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				return
			}

			// Check success expectations
			assert.NoError(t, err)
			assert.NotNil(t, config)
			assert.Equal(t, tt.expectedConfig.RPDisplayName, config.RPDisplayName)
			assert.Equal(t, tt.expectedConfig.RPID, config.RPID)
			assert.Equal(t, tt.expectedConfig.RPOrigins, config.RPOrigins)
		})
	}
}
