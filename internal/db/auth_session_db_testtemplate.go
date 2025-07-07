package db

import (
	"context"
	"testing"
	"time"

	"github.com/gianlucafrei/GoAM/internal/model"

	"github.com/stretchr/testify/assert"
)

// TemplateTestAuthSessionCRUD is a parameterized test for basic CRUD operations on auth sessions
func TemplateTestAuthSessionCRUD(t *testing.T, db AuthSessionDB) {
	ctx := context.Background()
	testTenant := "test-tenant"
	testRealm := "test-realm"
	now := time.Now().Truncate(time.Second) // Truncate to seconds

	// Create test session
	testSession := &model.PersistentAuthSession{
		Tenant:             testTenant,
		Realm:              testRealm,
		RunID:              "test-run-id",
		SessionIDHash:      "test-session-hash",
		CreatedAt:          now,
		ExpiresAt:          now.Add(1 * time.Hour),
		SessionInformation: []byte(`{"flow_id":"test-flow","current":"test-node"}`),
	}

	t.Run("CreateOrUpdateAuthSession", func(t *testing.T) {
		// Test initial creation
		err := db.CreateOrUpdateAuthSession(ctx, testSession)
		assert.NoError(t, err)

		// Test update with modified data
		updatedSession := *testSession
		updatedSession.RunID = "updated-run-id"
		updatedSession.SessionInformation = []byte(`{"flow_id":"updated-flow","current":"updated-node"}`)

		err = db.CreateOrUpdateAuthSession(ctx, &updatedSession)
		assert.NoError(t, err)

		// Verify the update
		retrieved, err := db.GetAuthSessionByHash(ctx, testTenant, testRealm, testSession.SessionIDHash)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "updated-run-id", retrieved.RunID)
		assert.Equal(t, []byte(`{"flow_id":"updated-flow","current":"updated-node"}`), retrieved.SessionInformation)
	})

	t.Run("GetAuthSessionByID", func(t *testing.T) {
		session, err := db.GetAuthSessionByID(ctx, testTenant, testRealm, "updated-run-id")
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "updated-run-id", session.RunID)
		assert.Equal(t, testSession.SessionIDHash, session.SessionIDHash)
		assert.Equal(t, testSession.CreatedAt.Truncate(time.Second), session.CreatedAt.Truncate(time.Second))
		assert.Equal(t, testSession.ExpiresAt.Truncate(time.Second), session.ExpiresAt.Truncate(time.Second))
		assert.Equal(t, []byte(`{"flow_id":"updated-flow","current":"updated-node"}`), session.SessionInformation)
	})

	t.Run("GetAuthSessionByHash", func(t *testing.T) {
		session, err := db.GetAuthSessionByHash(ctx, testTenant, testRealm, testSession.SessionIDHash)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "updated-run-id", session.RunID)
		assert.Equal(t, testSession.SessionIDHash, session.SessionIDHash)
	})

	t.Run("ListAuthSessions", func(t *testing.T) {
		// Create another session in the same tenant/realm
		anotherSession := &model.PersistentAuthSession{
			Tenant:             testTenant,
			Realm:              testRealm,
			RunID:              "another-run-id",
			SessionIDHash:      "another-session-hash",
			CreatedAt:          now,
			ExpiresAt:          now.Add(1 * time.Hour),
			SessionInformation: []byte(`{"flow_id":"another-flow","current":"another-node"}`),
		}
		err := db.CreateOrUpdateAuthSession(ctx, anotherSession)
		assert.NoError(t, err)

		sessions, err := db.ListAuthSessions(ctx, testTenant, testRealm)
		assert.NoError(t, err)
		assert.Len(t, sessions, 2)

		// Verify both sessions are in the list
		foundTest := false
		foundAnother := false
		for _, session := range sessions {
			if session.RunID == "updated-run-id" {
				foundTest = true
			}
			if session.RunID == anotherSession.RunID {
				foundAnother = true
			}
		}
		assert.True(t, foundTest)
		assert.True(t, foundAnother)
	})

	t.Run("ListAllAuthSessions", func(t *testing.T) {
		// Create a session in a different realm
		differentRealmSession := &model.PersistentAuthSession{
			Tenant:             testTenant,
			Realm:              "different-realm",
			RunID:              "different-run-id",
			SessionIDHash:      "different-session-hash",
			CreatedAt:          now,
			ExpiresAt:          now.Add(1 * time.Hour),
			SessionInformation: []byte(`{"flow_id":"different-flow","current":"different-node"}`),
		}
		err := db.CreateOrUpdateAuthSession(ctx, differentRealmSession)
		assert.NoError(t, err)

		sessions, err := db.ListAllAuthSessions(ctx, testTenant)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), 3) // Should include all sessions across realms

		// Verify all sessions are in the list
		foundTest := false
		foundAnother := false
		foundDifferent := false
		for _, session := range sessions {
			if session.RunID == "updated-run-id" {
				foundTest = true
			}
			if session.RunID == "another-run-id" {
				foundAnother = true
			}
			if session.RunID == differentRealmSession.RunID {
				foundDifferent = true
			}
		}
		assert.True(t, foundTest)
		assert.True(t, foundAnother)
		assert.True(t, foundDifferent)
	})

	t.Run("DeleteAuthSession", func(t *testing.T) {
		err := db.DeleteAuthSession(ctx, testTenant, testRealm, testSession.SessionIDHash)
		assert.NoError(t, err)

		session, err := db.GetAuthSessionByHash(ctx, testTenant, testRealm, testSession.SessionIDHash)
		assert.NoError(t, err)
		assert.Nil(t, session)
	})

	t.Run("DeleteExpiredAuthSessions", func(t *testing.T) {
		// Create an expired session
		expiredSession := &model.PersistentAuthSession{
			Tenant:             testTenant,
			Realm:              testRealm,
			RunID:              "expired-run-id",
			SessionIDHash:      "expired-session-hash",
			CreatedAt:          now.Add(-2 * time.Hour),
			ExpiresAt:          now.Add(-1 * time.Hour),
			SessionInformation: []byte(`{"flow_id":"expired-flow","current":"expired-node"}`),
		}

		err := db.CreateOrUpdateAuthSession(ctx, expiredSession)
		assert.NoError(t, err)

		// Delete expired sessions
		err = db.DeleteExpiredAuthSessions(ctx, testTenant, testRealm)
		assert.NoError(t, err)

		// Verify expired session is deleted
		session, err := db.GetAuthSessionByHash(ctx, testTenant, testRealm, expiredSession.SessionIDHash)
		assert.NoError(t, err)
		assert.Nil(t, session)
	})
}
