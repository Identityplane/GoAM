package db

import (
	"context"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TemplateTestSigningKeyCRUD is a parameterized test for basic CRUD operations on signing keys
func TemplateTestSigningKeyCRUD(t *testing.T, db SigningKeyDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"

	// Create test signing key
	testKey := model.SigningKey{
		Tenant:             testTenant,
		Realm:              testRealm,
		Kid:                "test-kid",
		Active:             true,
		Algorithm:          "RS256",
		Implementation:     "RSA",
		SigningKeyMaterial: "test-key-material",
		PublicKeyJWK:       "test-public-key-jwk",
		Created:            time.Now(),
	}

	t.Run("CreateSigningKey", func(t *testing.T) {
		err := db.CreateSigningKey(ctx, testKey)
		assert.NoError(t, err)
	})

	t.Run("GetSigningKey", func(t *testing.T) {
		key, err := db.GetSigningKey(ctx, testTenant, testRealm, testKey.Kid)
		assert.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, testKey.Kid, key.Kid)
		assert.Equal(t, testKey.Algorithm, key.Algorithm)
		assert.Equal(t, testKey.Implementation, key.Implementation)
		assert.Equal(t, testKey.SigningKeyMaterial, key.SigningKeyMaterial)
		assert.Equal(t, testKey.PublicKeyJWK, key.PublicKeyJWK)
	})

	t.Run("UpdateSigningKey", func(t *testing.T) {
		key, err := db.GetSigningKey(ctx, testTenant, testRealm, testKey.Kid)
		require.NoError(t, err)
		require.NotNil(t, key)

		key.Algorithm = "ES256"
		key.Implementation = "ECDSA"
		err = db.UpdateSigningKey(ctx, key)
		assert.NoError(t, err)

		updatedKey, err := db.GetSigningKey(ctx, testTenant, testRealm, testKey.Kid)
		assert.NoError(t, err)
		assert.Equal(t, "ES256", updatedKey.Algorithm)
		assert.Equal(t, "ECDSA", updatedKey.Implementation)
	})

	t.Run("ListSigningKeys", func(t *testing.T) {
		keys, err := db.ListSigningKeys(ctx, testTenant, testRealm)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, testKey.Kid, keys[0].Kid)
	})

	t.Run("ListActiveSigningKeys", func(t *testing.T) {
		keys, err := db.ListActiveSigningKeys(ctx, testTenant, testRealm)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		if len(keys) > 0 {
			assert.Equal(t, testKey.Kid, keys[0].Kid)
			assert.True(t, keys[0].Active)
		}
	})

	t.Run("DisableSigningKey", func(t *testing.T) {
		err := db.DisableSigningKey(ctx, testTenant, testRealm, testKey.Kid)
		assert.NoError(t, err)

		key, err := db.GetSigningKey(ctx, testTenant, testRealm, testKey.Kid)
		assert.NoError(t, err)
		assert.NotNil(t, key)
		assert.False(t, key.Active)
		assert.NotNil(t, key.Disabled)
	})

	t.Run("DeleteSigningKey", func(t *testing.T) {
		err := db.DeleteSigningKey(ctx, testTenant, testRealm, testKey.Kid)
		assert.NoError(t, err)

		key, err := db.GetSigningKey(ctx, testTenant, testRealm, testKey.Kid)
		assert.NoError(t, err)
		assert.Nil(t, key)
	})
}
