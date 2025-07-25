package db

import (
	"context"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TemplateTestClientSessionCRUD is a parameterized test for basic CRUD operations on client sessions
func TemplateTestClientSessionCRUD(t *testing.T, db ClientSessionDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	testClientID := "test-client"
	testUserID := "test-user"
	now := time.Now().Truncate(time.Second) // Truncate to seconds

	// Create test session
	testSession := &model.ClientSession{
		Tenant:           testTenant,
		Realm:            testRealm,
		ClientSessionID:  "test-session",
		ClientID:         testClientID,
		GrantType:        "authorization_code",
		AccessTokenHash:  "access-token-hash",
		RefreshTokenHash: "refresh-token-hash",
		AuthCodeHash:     "auth-code-hash",
		UserID:           testUserID,
		Scope:            "openid profile",
		LoginSessionJson: `{"state":"test-state","nonce":"test-nonce"}`,
		Created:          now,
		Expire:           now.Add(1 * time.Hour),
	}

	t.Run("CreateClientSession", func(t *testing.T) {
		err := db.CreateClientSession(ctx, testTenant, testRealm, testSession)
		assert.NoError(t, err)
	})

	t.Run("GetClientSessionByID", func(t *testing.T) {
		session, err := db.GetClientSessionByID(ctx, testTenant, testRealm, testSession.ClientSessionID)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, testSession.ClientSessionID, session.ClientSessionID)
		assert.Equal(t, testSession.ClientID, session.ClientID)
		assert.Equal(t, testSession.GrantType, session.GrantType)
		assert.Equal(t, testSession.AccessTokenHash, session.AccessTokenHash)
		assert.Equal(t, testSession.RefreshTokenHash, session.RefreshTokenHash)
		assert.Equal(t, testSession.AuthCodeHash, session.AuthCodeHash)
		assert.Equal(t, testSession.UserID, session.UserID)
		assert.Equal(t, testSession.Scope, session.Scope)
		assert.Equal(t, testSession.LoginSessionJson, session.LoginSessionJson)
		assert.Equal(t, testSession.Created.Truncate(time.Second), session.Created.Truncate(time.Second))
		assert.Equal(t, testSession.Expire.Truncate(time.Second), session.Expire.Truncate(time.Second))
	})

	t.Run("GetClientSessionByAccessToken", func(t *testing.T) {
		session, err := db.GetClientSessionByAccessToken(ctx, testTenant, testRealm, testSession.AccessTokenHash)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, testSession.ClientSessionID, session.ClientSessionID)
	})

	t.Run("GetClientSessionByRefreshToken", func(t *testing.T) {
		session, err := db.GetClientSessionByRefreshToken(ctx, testTenant, testRealm, testSession.RefreshTokenHash)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, testSession.ClientSessionID, session.ClientSessionID)
	})

	t.Run("GetClientSessionByAuthCode", func(t *testing.T) {
		session, err := db.GetClientSessionByAuthCode(ctx, testTenant, testRealm, testSession.AuthCodeHash)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, testSession.ClientSessionID, session.ClientSessionID)
	})

	t.Run("ListClientSessions", func(t *testing.T) {
		sessions, err := db.ListClientSessions(ctx, testTenant, testRealm, testClientID)
		assert.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, testSession.ClientSessionID, sessions[0].ClientSessionID)
	})

	t.Run("ListUserClientSessions", func(t *testing.T) {
		sessions, err := db.ListUserClientSessions(ctx, testTenant, testRealm, testUserID)
		assert.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, testSession.ClientSessionID, sessions[0].ClientSessionID)
	})

	t.Run("UpdateClientSession", func(t *testing.T) {
		session, err := db.GetClientSessionByID(ctx, testTenant, testRealm, testSession.ClientSessionID)
		require.NoError(t, err)
		require.NotNil(t, session)

		session.Scope = "openid profile email"
		session.Expire = now.Add(2 * time.Hour)
		err = db.UpdateClientSession(ctx, testTenant, testRealm, session)
		assert.NoError(t, err)

		updatedSession, err := db.GetClientSessionByID(ctx, testTenant, testRealm, testSession.ClientSessionID)
		assert.NoError(t, err)
		assert.Equal(t, "openid profile email", updatedSession.Scope)
		assert.Equal(t, now.Add(2*time.Hour).Truncate(time.Second), updatedSession.Expire.Truncate(time.Second))
	})

	t.Run("DeleteClientSession", func(t *testing.T) {
		err := db.DeleteClientSession(ctx, testTenant, testRealm, testSession.ClientSessionID)
		assert.NoError(t, err)

		session, err := db.GetClientSessionByID(ctx, testTenant, testRealm, testSession.ClientSessionID)
		assert.NoError(t, err)
		assert.Nil(t, session)
	})

	t.Run("DeleteExpiredClientSessions", func(t *testing.T) {
		// Create an expired session
		expiredSession := &model.ClientSession{
			Tenant:           testTenant,
			Realm:            testRealm,
			ClientSessionID:  "expired-session",
			ClientID:         testClientID,
			GrantType:        "authorization_code",
			AccessTokenHash:  "expired-access-token-hash",
			RefreshTokenHash: "expired-refresh-token-hash",
			AuthCodeHash:     "expired-auth-code-hash",
			UserID:           testUserID,
			Scope:            "openid profile",
			LoginSessionJson: `{"state":"expired-state","nonce":"expired-nonce"}`,
			Created:          now.Add(-2 * time.Hour),
			Expire:           now.Add(-1 * time.Hour),
		}

		err := db.CreateClientSession(ctx, testTenant, testRealm, expiredSession)
		assert.NoError(t, err)

		// Delete expired sessions
		err = db.DeleteExpiredClientSessions(ctx, testTenant, testRealm)
		assert.NoError(t, err)

		// Verify expired session is deleted
		session, err := db.GetClientSessionByID(ctx, testTenant, testRealm, expiredSession.ClientSessionID)
		assert.NoError(t, err)
		assert.Nil(t, session)
	})

	t.Run("CreateAndQueryByAuthCodeHash", func(t *testing.T) {
		// Create a new session with a specific auth code hash
		authCodeSession := &model.ClientSession{
			Tenant:           testTenant,
			Realm:            testRealm,
			ClientSessionID:  "auth-code-session",
			ClientID:         testClientID,
			GrantType:        "authorization_code",
			AccessTokenHash:  "new-access-token-hash",
			RefreshTokenHash: "new-refresh-token-hash",
			AuthCodeHash:     "new-auth-code-hash",
			UserID:           testUserID,
			Scope:            "openid profile",
			LoginSessionJson: `{"state":"auth-code-state","nonce":"auth-code-nonce"}`,
			Created:          now,
			Expire:           now.Add(1 * time.Hour),
		}

		err := db.CreateClientSession(ctx, testTenant, testRealm, authCodeSession)
		assert.NoError(t, err)

		// Query by auth code hash
		session, err := db.GetClientSessionByAuthCode(ctx, testTenant, testRealm, authCodeSession.AuthCodeHash)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, authCodeSession.ClientSessionID, session.ClientSessionID)
		assert.Equal(t, authCodeSession.AuthCodeHash, session.AuthCodeHash)
	})

	t.Run("CreateAndQueryByAccessTokenHash", func(t *testing.T) {
		// Create a new session with a specific access token hash
		accessTokenSession := &model.ClientSession{
			Tenant:           testTenant,
			Realm:            testRealm,
			ClientSessionID:  "access-token-session",
			ClientID:         testClientID,
			GrantType:        "authorization_code",
			AccessTokenHash:  "unique-access-token-hash",
			RefreshTokenHash: "unique-refresh-token-hash",
			AuthCodeHash:     "unique-auth-code-hash",
			UserID:           testUserID,
			Scope:            "openid profile",
			LoginSessionJson: `{"state":"access-token-state","nonce":"access-token-nonce"}`,
			Created:          now,
			Expire:           now.Add(1 * time.Hour),
		}

		err := db.CreateClientSession(ctx, testTenant, testRealm, accessTokenSession)
		assert.NoError(t, err)

		// Query by access token hash
		session, err := db.GetClientSessionByAccessToken(ctx, testTenant, testRealm, accessTokenSession.AccessTokenHash)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, accessTokenSession.ClientSessionID, session.ClientSessionID)
		assert.Equal(t, accessTokenSession.AccessTokenHash, session.AccessTokenHash)
	})

	t.Run("CreateAndQueryByRefreshTokenHash", func(t *testing.T) {
		// Create a new session with a specific refresh token hash
		refreshTokenSession := &model.ClientSession{
			Tenant:           testTenant,
			Realm:            testRealm,
			ClientSessionID:  "refresh-token-session",
			ClientID:         testClientID,
			GrantType:        "authorization_code",
			AccessTokenHash:  "special-access-token-hash",
			RefreshTokenHash: "special-refresh-token-hash",
			AuthCodeHash:     "special-auth-code-hash",
			UserID:           testUserID,
			Scope:            "openid profile",
			LoginSessionJson: `{"state":"refresh-token-state","nonce":"refresh-token-nonce"}`,
			Created:          now,
			Expire:           now.Add(1 * time.Hour),
		}

		err := db.CreateClientSession(ctx, testTenant, testRealm, refreshTokenSession)
		assert.NoError(t, err)

		// Query by refresh token hash
		session, err := db.GetClientSessionByRefreshToken(ctx, testTenant, testRealm, refreshTokenSession.RefreshTokenHash)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, refreshTokenSession.ClientSessionID, session.ClientSessionID)
		assert.Equal(t, refreshTokenSession.RefreshTokenHash, session.RefreshTokenHash)
	})
}
