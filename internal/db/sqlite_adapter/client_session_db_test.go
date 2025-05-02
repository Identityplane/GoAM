package sqlite_adapter

import (
	"context"
	"goiam/internal/db"
	"goiam/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClientSessionCRUD(t *testing.T) {
	sqldb := setupTestDB(t)
	clientSessionDB := NewClientSessionDB(sqldb)

	db.TemplateTestClientSessionCRUD(t, clientSessionDB)
}

func TestClientSessionUniqueConstraints(t *testing.T) {
	sqldb := setupTestDB(t)
	clientSessionDB := NewClientSessionDB(sqldb)

	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	testClientID := "test-client"
	testUserID := "test-user"
	now := time.Now()

	// Create first session
	session1 := &model.ClientSession{
		Tenant:           testTenant,
		Realm:            testRealm,
		ClientSessionID:  "session1",
		ClientID:         testClientID,
		GrantType:        "authorization_code",
		AccessTokenHash:  "access-token-hash-1",
		RefreshTokenHash: "refresh-token-hash-1",
		AuthCodeHash:     "auth-code-hash-1",
		UserID:           testUserID,
		Scope:            "openid profile",
		Created:          now,
		Expire:           now.Add(1 * time.Hour),
	}
	err := clientSessionDB.CreateClientSession(ctx, session1)
	require.NoError(t, err)

	// Try to create another session with same access token hash (should fail)
	session2 := &model.ClientSession{
		Tenant:           testTenant,
		Realm:            testRealm,
		ClientSessionID:  "session2",
		ClientID:         testClientID,
		GrantType:        "authorization_code",
		AccessTokenHash:  "access-token-hash-1", // same access token hash
		RefreshTokenHash: "refresh-token-hash-2",
		AuthCodeHash:     "auth-code-hash-2",
		UserID:           testUserID,
		Scope:            "openid profile",
		Created:          now,
		Expire:           now.Add(1 * time.Hour),
	}
	err = clientSessionDB.CreateClientSession(ctx, session2)
	require.Error(t, err)

	// Try to create another session with same refresh token hash (should fail)
	session3 := &model.ClientSession{
		Tenant:           testTenant,
		Realm:            testRealm,
		ClientSessionID:  "session3",
		ClientID:         testClientID,
		GrantType:        "authorization_code",
		AccessTokenHash:  "access-token-hash-3",
		RefreshTokenHash: "refresh-token-hash-1", // same refresh token hash
		AuthCodeHash:     "auth-code-hash-3",
		UserID:           testUserID,
		Scope:            "openid profile",
		Created:          now,
		Expire:           now.Add(1 * time.Hour),
	}
	err = clientSessionDB.CreateClientSession(ctx, session3)
	require.Error(t, err)

	// Try to create another session with same auth code hash (should fail)
	session4 := &model.ClientSession{
		Tenant:           testTenant,
		Realm:            testRealm,
		ClientSessionID:  "session4",
		ClientID:         testClientID,
		GrantType:        "authorization_code",
		AccessTokenHash:  "access-token-hash-4",
		RefreshTokenHash: "refresh-token-hash-4",
		AuthCodeHash:     "auth-code-hash-1", // same auth code hash
		UserID:           testUserID,
		Scope:            "openid profile",
		Created:          now,
		Expire:           now.Add(1 * time.Hour),
	}
	err = clientSessionDB.CreateClientSession(ctx, session4)
	require.Error(t, err)
}
