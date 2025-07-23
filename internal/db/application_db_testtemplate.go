package db

import (
	"context"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TemplateTestApplicationCRUD is a parameterized test for basic CRUD operations on applications
func TemplateTestApplicationCRUD(t *testing.T, db ApplicationDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create test application
	testApp := model.Application{
		Tenant:                     testTenant,
		Realm:                      testRealm,
		ClientId:                   "test-app",
		ClientSecret:               "test-secret",
		Confidential:               true,
		ConsentRequired:            false,
		Description:                "A test application",
		AllowedScopes:              []string{"openid", "profile", "email"},
		AllowedGrants:              []string{"authorization_code", "refresh_token"},
		AllowedAuthenticationFlows: []string{"password", "client_credentials"},
		AccessTokenLifetime:        3600,
		RefreshTokenLifetime:       86400,
		IdTokenLifetime:            3600,
		AccessTokenType:            "jwt",
		AccessTokenAlgorithm:       "RS256",
		AccessTokenMapping:         "default",
		IdTokenAlgorithm:           "RS256",
		IdTokenMapping:             "default",
		RedirectUris:               []string{"https://example.com/callback", "https://example.com/oauth2/callback"},
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	t.Run("CreateApplication", func(t *testing.T) {
		err := db.CreateApplication(ctx, testApp)
		assert.NoError(t, err)
	})

	t.Run("GetApplication", func(t *testing.T) {
		app, err := db.GetApplication(ctx, testTenant, testRealm, testApp.ClientId)
		assert.NoError(t, err)
		assert.NotNil(t, app)
		assert.Equal(t, testApp.ClientId, app.ClientId)
		assert.Equal(t, testApp.ClientSecret, app.ClientSecret)
		assert.Equal(t, testApp.Description, app.Description)
		assert.Equal(t, testApp.AllowedScopes, app.AllowedScopes)
		assert.Equal(t, testApp.AllowedGrants, app.AllowedGrants)
		assert.Equal(t, testApp.AllowedAuthenticationFlows, app.AllowedAuthenticationFlows)
		assert.Equal(t, testApp.AccessTokenLifetime, app.AccessTokenLifetime)
		assert.Equal(t, testApp.RefreshTokenLifetime, app.RefreshTokenLifetime)
		assert.Equal(t, testApp.IdTokenLifetime, app.IdTokenLifetime)
		assert.Equal(t, testApp.AccessTokenType, app.AccessTokenType)
		assert.Equal(t, testApp.AccessTokenAlgorithm, app.AccessTokenAlgorithm)
		assert.Equal(t, testApp.AccessTokenMapping, app.AccessTokenMapping)
		assert.Equal(t, testApp.IdTokenAlgorithm, app.IdTokenAlgorithm)
		assert.Equal(t, testApp.IdTokenMapping, app.IdTokenMapping)
		assert.Equal(t, testApp.RedirectUris, app.RedirectUris)
	})

	t.Run("UpdateApplication", func(t *testing.T) {
		app, err := db.GetApplication(ctx, testTenant, testRealm, testApp.ClientId)
		require.NoError(t, err)
		require.NotNil(t, app)

		app.Description = "Updated description"
		app.AllowedScopes = []string{"openid", "profile"}
		app.AccessTokenLifetime = 7200
		app.RefreshTokenLifetime = 172800
		app.IdTokenLifetime = 7200
		app.AccessTokenType = "session_key"
		app.AccessTokenAlgorithm = "HS256"
		app.AccessTokenMapping = "custom"
		app.IdTokenAlgorithm = "HS256"
		app.IdTokenMapping = "custom"
		app.RedirectUris = []string{"https://example.com/new-callback"}
		err = db.UpdateApplication(ctx, app)
		assert.NoError(t, err)

		updatedApp, err := db.GetApplication(ctx, testTenant, testRealm, testApp.ClientId)
		assert.NoError(t, err)
		assert.Equal(t, "Updated description", updatedApp.Description)
		assert.Equal(t, []string{"openid", "profile"}, updatedApp.AllowedScopes)
		assert.Equal(t, 7200, updatedApp.AccessTokenLifetime)
		assert.Equal(t, 172800, updatedApp.RefreshTokenLifetime)
		assert.Equal(t, 7200, updatedApp.IdTokenLifetime)
		assert.Equal(t, "session_key", string(updatedApp.AccessTokenType))
		assert.Equal(t, "HS256", updatedApp.AccessTokenAlgorithm)
		assert.Equal(t, "custom", updatedApp.AccessTokenMapping)
		assert.Equal(t, "HS256", updatedApp.IdTokenAlgorithm)
		assert.Equal(t, "custom", updatedApp.IdTokenMapping)
		assert.Equal(t, []string{"https://example.com/new-callback"}, updatedApp.RedirectUris)
	})

	t.Run("ListApplications", func(t *testing.T) {
		apps, err := db.ListApplications(ctx, testTenant, testRealm)
		assert.NoError(t, err)
		assert.Len(t, apps, 1)
		assert.Equal(t, testApp.ClientId, apps[0].ClientId)
		assert.Equal(t, []string{"https://example.com/new-callback"}, apps[0].RedirectUris)
	})

	t.Run("DeleteApplication", func(t *testing.T) {
		err := db.DeleteApplication(ctx, testTenant, testRealm, testApp.ClientId)
		assert.NoError(t, err)

		app, err := db.GetApplication(ctx, testTenant, testRealm, testApp.ClientId)
		assert.NoError(t, err)
		assert.Nil(t, app)
	})
}
