package db

import (
	"context"
	"testing"

	"github.com/gianlucafrei/GoAM/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TemplateTestRealmCRUD is a parameterized test for basic CRUD operations on realms
func TemplateTestRealmCRUD(t *testing.T, db RealmDB) {
	ctx := context.Background()
	testTenant := "test-tenant"

	// Create test realm
	testRealm := model.Realm{
		Realm:     "test-realm",
		RealmName: "Test Realm",
		Tenant:    testTenant,
		BaseUrl:   "https://test.example.com",
	}

	t.Run("CreateRealm", func(t *testing.T) {
		err := db.CreateRealm(ctx, testRealm)
		assert.NoError(t, err)
	})

	t.Run("GetRealm", func(t *testing.T) {
		realm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		assert.NoError(t, err)
		assert.NotNil(t, realm)
		assert.Equal(t, testRealm.Realm, realm.Realm)
		assert.Equal(t, testRealm.RealmName, realm.RealmName)
		assert.Equal(t, testRealm.BaseUrl, realm.BaseUrl)
	})

	t.Run("UpdateRealm", func(t *testing.T) {
		realm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		require.NoError(t, err)
		require.NotNil(t, realm)

		realm.RealmName = "Updated Test Realm"
		realm.BaseUrl = "https://updated.example.com"
		err = db.UpdateRealm(ctx, realm)
		assert.NoError(t, err)

		updatedRealm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Test Realm", updatedRealm.RealmName)
		assert.Equal(t, "https://updated.example.com", updatedRealm.BaseUrl)
	})

	t.Run("ListRealms", func(t *testing.T) {
		realms, err := db.ListRealms(ctx, testTenant)
		assert.NoError(t, err)
		assert.Len(t, realms, 1)
		assert.Equal(t, testRealm.Realm, realms[0].Realm)
		assert.Equal(t, "https://updated.example.com", realms[0].BaseUrl)
	})

	t.Run("DeleteRealm", func(t *testing.T) {
		// First check if realm is empty
		err := db.DeleteRealm(ctx, testTenant, testRealm.Realm)
		assert.NoError(t, err)

		realm, err := db.GetRealm(ctx, testTenant, testRealm.Realm)
		assert.NoError(t, err)
		assert.Nil(t, realm)
	})
}
