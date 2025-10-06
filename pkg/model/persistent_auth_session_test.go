package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPersistentAuthSession_MarshalUnmarshal(t *testing.T) {
	// Create a sample AuthenticationSession
	now := time.Now()
	session := &AuthenticationSession{
		RunID:         "test-run-id",
		SessionIdHash: "test-session-hash",
		FlowId:        "test-flow-id",
		Current:       "test-node",
		Context: map[string]string{
			"key1": "value1",
		},
		History:      []string{"node1", "node2"},
		CreatedAt:    now,
		ExpiresAt:    now.Add(time.Hour),
		LoginUriBase: "http://example.com/login",
		LoginUriNext: "http://example.com/login/next",
	}

	// Create PersistentAuthSession
	tenant := "test-tenant"
	realm := "test-realm"
	persistentSession, err := NewPersistentAuthSession(tenant, realm, session)
	assert.NoError(t, err)
	assert.NotNil(t, persistentSession)

	// Verify PersistentAuthSession fields
	assert.Equal(t, tenant, persistentSession.Tenant)
	assert.Equal(t, realm, persistentSession.Realm)
	assert.Equal(t, session.RunID, persistentSession.RunID)
	assert.Equal(t, session.SessionIdHash, persistentSession.SessionIDHash)
	assert.True(t, session.CreatedAt.Equal(persistentSession.CreatedAt))
	assert.True(t, session.ExpiresAt.Equal(persistentSession.ExpiresAt))
	assert.NotEmpty(t, persistentSession.SessionInformation)

	// Convert back to AuthenticationSession
	recoveredSession, err := persistentSession.ToAuthenticationSession()
	assert.NoError(t, err)
	assert.NotNil(t, recoveredSession)

	// Verify recovered session matches original
	assert.Equal(t, session.RunID, recoveredSession.RunID)
	assert.Equal(t, session.SessionIdHash, recoveredSession.SessionIdHash)
	assert.Equal(t, session.FlowId, recoveredSession.FlowId)
	assert.Equal(t, session.Current, recoveredSession.Current)
	assert.Equal(t, session.Context, recoveredSession.Context)
	assert.Equal(t, session.History, recoveredSession.History)
	assert.True(t, session.CreatedAt.Equal(recoveredSession.CreatedAt))
	assert.True(t, session.ExpiresAt.Equal(recoveredSession.ExpiresAt))
	assert.Equal(t, session.LoginUriNext, recoveredSession.LoginUriNext)
	assert.Equal(t, session.LoginUriBase, recoveredSession.LoginUriBase)
}
