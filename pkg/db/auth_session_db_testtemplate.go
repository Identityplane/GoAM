package db

import (
	"context"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"

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
		Tenant:                 testTenant,
		Realm:                  testRealm,
		RunID:                  "test-run-id",
		PrimarySessionIDHash:   "test-session-hash",
		SecondarySessionIDHash: "test-secondary-session-hash",
		CreatedAt:              now,
		ExpiresAt:              now.Add(1 * time.Hour),
		SessionInformation:     []byte(`{"flow_id":"test-flow","current":"test-node"}`),
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
		retrieved, err := db.GetAuthSessionByHash(ctx, testTenant, testRealm, testSession.PrimarySessionIDHash)
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
		assert.Equal(t, testSession.PrimarySessionIDHash, session.PrimarySessionIDHash)
		assert.Equal(t, testSession.CreatedAt.Truncate(time.Second), session.CreatedAt.Truncate(time.Second))
		assert.Equal(t, testSession.ExpiresAt.Truncate(time.Second), session.ExpiresAt.Truncate(time.Second))
		assert.Equal(t, []byte(`{"flow_id":"updated-flow","current":"updated-node"}`), session.SessionInformation)
	})

	t.Run("GetAuthSessionByHash", func(t *testing.T) {
		// Test finding by primary session ID hash
		session, err := db.GetAuthSessionByHash(ctx, testTenant, testRealm, testSession.PrimarySessionIDHash)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "updated-run-id", session.RunID)
		assert.Equal(t, testSession.PrimarySessionIDHash, session.PrimarySessionIDHash)
		assert.Equal(t, testSession.SecondarySessionIDHash, session.SecondarySessionIDHash)

		// Test finding by secondary session ID hash (should also work due to OR condition)
		session, err = db.GetAuthSessionByHash(ctx, testTenant, testRealm, testSession.SecondarySessionIDHash)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "updated-run-id", session.RunID)
		assert.Equal(t, testSession.PrimarySessionIDHash, session.PrimarySessionIDHash)
		assert.Equal(t, testSession.SecondarySessionIDHash, session.SecondarySessionIDHash)

		// Test with non-existent hash
		nonExistent, err := db.GetAuthSessionByHash(ctx, testTenant, testRealm, "non-existent-hash")
		assert.NoError(t, err)
		assert.Nil(t, nonExistent)
	})


	t.Run("ListAuthSessions", func(t *testing.T) {
		// Create another session in the same tenant/realm
		anotherSession := &model.PersistentAuthSession{
			Tenant:               testTenant,
			Realm:                testRealm,
			RunID:                "another-run-id",
			PrimarySessionIDHash: "another-session-hash",
			CreatedAt:            now,
			ExpiresAt:            now.Add(1 * time.Hour),
			SessionInformation:   []byte(`{"flow_id":"another-flow","current":"another-node"}`),
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
			Tenant:               testTenant,
			Realm:                "different-realm",
			RunID:                "different-run-id",
			PrimarySessionIDHash: "different-session-hash",
			CreatedAt:            now,
			ExpiresAt:            now.Add(1 * time.Hour),
			SessionInformation:   []byte(`{"flow_id":"different-flow","current":"different-node"}`),
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
		err := db.DeleteAuthSession(ctx, testTenant, testRealm, testSession.PrimarySessionIDHash)
		assert.NoError(t, err)

		session, err := db.GetAuthSessionByHash(ctx, testTenant, testRealm, testSession.PrimarySessionIDHash)
		assert.NoError(t, err)
		assert.Nil(t, session)
	})

	t.Run("DeleteExpiredAuthSessions", func(t *testing.T) {
		// Create an expired session
		expiredSession := &model.PersistentAuthSession{
			Tenant:               testTenant,
			Realm:                testRealm,
			RunID:                "expired-run-id",
			PrimarySessionIDHash: "expired-session-hash",
			CreatedAt:            now.Add(-2 * time.Hour),
			ExpiresAt:            now.Add(-1 * time.Hour),
			SessionInformation:   []byte(`{"flow_id":"expired-flow","current":"expired-node"}`),
		}

		err := db.CreateOrUpdateAuthSession(ctx, expiredSession)
		assert.NoError(t, err)

		// Delete expired sessions
		err = db.DeleteExpiredAuthSessions(ctx, testTenant, testRealm)
		assert.NoError(t, err)

		// Verify expired session is deleted
		session, err := db.GetAuthSessionByHash(ctx, testTenant, testRealm, expiredSession.PrimarySessionIDHash)
		assert.NoError(t, err)
		assert.Nil(t, session)
	})
}
