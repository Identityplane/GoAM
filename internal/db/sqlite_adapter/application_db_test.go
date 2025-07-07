package sqlite_adapter

import (
	"context"
	"testing"

	"github.com/gianlucafrei/GoAM/internal/db"
	"github.com/gianlucafrei/GoAM/internal/model"

	"github.com/stretchr/testify/require"
)

func TestApplicationCRUD(t *testing.T) {
	sqldb := setupTestDB(t)
	appDB, err := NewApplicationDB(sqldb)
	require.NoError(t, err)

	db.TemplateTestApplicationCRUD(t, appDB)
}

func TestApplicationUniqueConstraints(t *testing.T) {
	sqldb := setupTestDB(t)
	appDB, err := NewApplicationDB(sqldb)
	require.NoError(t, err)

	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create first application
	app1 := model.Application{
		Tenant:                     testTenant,
		Realm:                      testRealm,
		ClientId:                   "app1",
		ClientSecret:               "secret1",
		Confidential:               true,
		ConsentRequired:            false,
		Description:                "Test app 1",
		AllowedScopes:              []string{"openid", "profile"},
		AllowedGrants:              []string{"authorization_code"},
		AllowedAuthenticationFlows: []string{"login"},
	}
	err = appDB.CreateApplication(ctx, app1)
	require.NoError(t, err)

	// Try to create another application with same client_id (should fail)
	app2 := model.Application{
		Tenant:                     testTenant,
		Realm:                      testRealm,
		ClientId:                   "app1", // same client_id
		ClientSecret:               "secret2",
		Confidential:               true,
		ConsentRequired:            false,
		Description:                "Test app 2",
		AllowedScopes:              []string{"openid", "email"},
		AllowedGrants:              []string{"authorization_code"},
		AllowedAuthenticationFlows: []string{"login"},
	}
	err = appDB.CreateApplication(ctx, app2)
	require.Error(t, err)
}
