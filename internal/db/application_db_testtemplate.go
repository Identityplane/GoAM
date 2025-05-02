package db

import (
	"context"
	"testing"
	"time"

	"goiam/internal/model"

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
		Tenant:          testTenant,
		Realm:           testRealm,
		ClientId:        "test-app",
		ClientSecret:    "test-secret",
		Confidential:    true,
		ConsentRequired: false,
		Description:     "A test application",
		AllowedScopes:   []string{"openid", "profile", "email"},
		AllowedFlows:    []string{"code", "code + pkce"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
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
		assert.Equal(t, testApp.AllowedFlows, app.AllowedFlows)
	})

	t.Run("UpdateApplication", func(t *testing.T) {
		app, err := db.GetApplication(ctx, testTenant, testRealm, testApp.ClientId)
		require.NoError(t, err)
		require.NotNil(t, app)

		app.Description = "Updated description"
		app.AllowedScopes = []string{"openid", "profile"}
		err = db.UpdateApplication(ctx, app)
		assert.NoError(t, err)

		updatedApp, err := db.GetApplication(ctx, testTenant, testRealm, testApp.ClientId)
		assert.NoError(t, err)
		assert.Equal(t, "Updated description", updatedApp.Description)
		assert.Equal(t, []string{"openid", "profile"}, updatedApp.AllowedScopes)
	})

	t.Run("ListApplications", func(t *testing.T) {
		apps, err := db.ListApplications(ctx, testTenant, testRealm)
		assert.NoError(t, err)
		assert.Len(t, apps, 1)
		assert.Equal(t, testApp.ClientId, apps[0].ClientId)
	})

	t.Run("DeleteApplication", func(t *testing.T) {
		err := db.DeleteApplication(ctx, testTenant, testRealm, testApp.ClientId)
		assert.NoError(t, err)

		app, err := db.GetApplication(ctx, testTenant, testRealm, testApp.ClientId)
		assert.NoError(t, err)
		assert.Nil(t, app)
	})
}
