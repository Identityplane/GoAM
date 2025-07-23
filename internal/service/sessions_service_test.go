package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTimeProvider is a mock implementation of TimeProvider
type mockTimeProvider struct {
	currentTime time.Time
}

func newMockTimeProvider() *mockTimeProvider {
	return &mockTimeProvider{
		currentTime: time.Now(),
	}
}

func (m *mockTimeProvider) Now() time.Time {
	return m.currentTime
}

func (m *mockTimeProvider) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

// mockClientSessionDB is a mock implementation of db.ClientSessionDB
type mockClientSessionDB struct {
	sessions map[string]*model.ClientSession
}

func newMockClientSessionDB() *mockClientSessionDB {
	return &mockClientSessionDB{
		sessions: make(map[string]*model.ClientSession),
	}
}

func (m *mockClientSessionDB) CreateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error {
	key := tenant + ":" + realm + ":" + session.ClientSessionID
	m.sessions[key] = session
	return nil
}

func (m *mockClientSessionDB) GetClientSessionByID(ctx context.Context, tenant, realm, sessionID string) (*model.ClientSession, error) {
	key := tenant + ":" + realm + ":" + sessionID
	return m.sessions[key], nil
}

func (m *mockClientSessionDB) GetClientSessionByAccessToken(ctx context.Context, tenant, realm, accessTokenHash string) (*model.ClientSession, error) {
	for _, session := range m.sessions {
		if session.AccessTokenHash == accessTokenHash {
			return session, nil
		}
	}
	return nil, nil
}

func (m *mockClientSessionDB) GetClientSessionByRefreshToken(ctx context.Context, tenant, realm, refreshTokenHash string) (*model.ClientSession, error) {
	for _, session := range m.sessions {
		if session.RefreshTokenHash == refreshTokenHash {
			return session, nil
		}
	}
	return nil, nil
}

func (m *mockClientSessionDB) GetClientSessionByAuthCode(ctx context.Context, tenant, realm, authCodeHash string) (*model.ClientSession, error) {
	for _, session := range m.sessions {
		if session.AuthCodeHash == authCodeHash {
			return session, nil
		}
	}
	return nil, nil
}

func (m *mockClientSessionDB) ListClientSessions(ctx context.Context, tenant, realm, clientID string) ([]model.ClientSession, error) {
	var sessions []model.ClientSession
	for _, session := range m.sessions {
		if session.ClientID == clientID {
			sessions = append(sessions, *session)
		}
	}
	return sessions, nil
}

func (m *mockClientSessionDB) ListUserClientSessions(ctx context.Context, tenant, realm, userID string) ([]model.ClientSession, error) {
	var sessions []model.ClientSession
	for _, session := range m.sessions {
		if session.UserID == userID {
			sessions = append(sessions, *session)
		}
	}
	return sessions, nil
}

func (m *mockClientSessionDB) UpdateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error {
	key := tenant + ":" + realm + ":" + session.ClientSessionID
	m.sessions[key] = session
	return nil
}

func (m *mockClientSessionDB) DeleteClientSession(ctx context.Context, tenant, realm, sessionID string) error {
	key := tenant + ":" + realm + ":" + sessionID
	delete(m.sessions, key)
	return nil
}

func (m *mockClientSessionDB) DeleteExpiredClientSessions(ctx context.Context, tenant, realm string) error {
	now := time.Now()
	for key, session := range m.sessions {
		if session.Expire.Before(now) {
			delete(m.sessions, key)
		}
	}
	return nil
}

type mockAuthSessionDB struct {
	sessions map[string]*model.PersistentAuthSession
	mu       sync.RWMutex
}

func newMockAuthSessionDB() *mockAuthSessionDB {
	return &mockAuthSessionDB{
		sessions: make(map[string]*model.PersistentAuthSession),
	}
}

func (m *mockAuthSessionDB) CreateAuthSession(ctx context.Context, session *model.PersistentAuthSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := session.Tenant + ":" + session.Realm + ":" + session.SessionIDHash
	m.sessions[key] = session
	return nil
}

func (m *mockAuthSessionDB) GetAuthSessionByID(ctx context.Context, tenant, realm, runID string) (*model.PersistentAuthSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		if session.Tenant == tenant && session.Realm == realm && session.RunID == runID {
			return session, nil
		}
	}
	return nil, nil
}

func (m *mockAuthSessionDB) GetAuthSessionByHash(ctx context.Context, tenant, realm, sessionIDHash string) (*model.PersistentAuthSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := tenant + ":" + realm + ":" + sessionIDHash
	if session, ok := m.sessions[key]; ok {
		return session, nil
	}
	return nil, nil
}

func (m *mockAuthSessionDB) ListAuthSessions(ctx context.Context, tenant, realm string) ([]model.PersistentAuthSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []model.PersistentAuthSession
	for _, session := range m.sessions {
		if session.Tenant == tenant && session.Realm == realm {
			sessions = append(sessions, *session)
		}
	}
	return sessions, nil
}

func (m *mockAuthSessionDB) ListAllAuthSessions(ctx context.Context, tenant string) ([]model.PersistentAuthSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []model.PersistentAuthSession
	for _, session := range m.sessions {
		if session.Tenant == tenant {
			sessions = append(sessions, *session)
		}
	}
	return sessions, nil
}

func (m *mockAuthSessionDB) DeleteAuthSession(ctx context.Context, tenant, realm, sessionIDHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := tenant + ":" + realm + ":" + sessionIDHash
	delete(m.sessions, key)
	return nil
}

func (m *mockAuthSessionDB) DeleteExpiredAuthSessions(ctx context.Context, tenant, realm string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, session := range m.sessions {
		if session.Tenant == tenant && session.Realm == realm && session.ExpiresAt.Before(now) {
			delete(m.sessions, key)
		}
	}
	return nil
}

func (m *mockAuthSessionDB) CreateOrUpdateAuthSession(ctx context.Context, session *model.PersistentAuthSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := session.Tenant + ":" + session.Realm + ":" + session.SessionIDHash
	m.sessions[key] = session
	return nil
}

func TestSessionsService(t *testing.T) {
	ctx := context.Background()
	mockDB := newMockClientSessionDB()
	mockAuthSessionDB := newMockAuthSessionDB()
	service := NewSessionsService(mockDB, mockAuthSessionDB)
	mockTime := newMockTimeProvider()
	service.SetTimeProvider(mockTime)

	testTenant := "test-tenant"
	testRealm := "test-realm"
	testClientID := "test-client"
	testUserID := "test-user"
	testScope := []string{"openid", "profile"}

	t.Run("CreateSessionObject", func(t *testing.T) {
		flowID := "test-flow"
		loginURI := "/login"
		session, sessionID := service.CreateAuthSessionObject(testTenant, testRealm, flowID, loginURI)

		assert.NotEmpty(t, sessionID)
		assert.Equal(t, flowID, session.FlowId)
		assert.Equal(t, loginURI, session.LoginUri)
		assert.NotEmpty(t, session.RunID)
		assert.NotEmpty(t, session.SessionIdHash)
		assert.NotNil(t, session.Context)
		assert.NotNil(t, session.History)
		assert.NotNil(t, session.Prompts)
		assert.True(t, session.ExpiresAt.After(time.Now()))
	})

	t.Run("CreateAndGetAuthenticationSession", func(t *testing.T) {
		flowID := "test-flow"
		loginURI := "/login"
		session, sessionID := service.CreateAuthSessionObject(testTenant, testRealm, flowID, loginURI)

		err := service.CreateOrUpdateAuthenticationSession(ctx, testTenant, testRealm, *session)
		require.NoError(t, err)

		retrievedSession, exists := service.GetAuthenticationSessionByID(ctx, testTenant, testRealm, sessionID)
		assert.True(t, exists)
		assert.Equal(t, session.SessionIdHash, retrievedSession.SessionIdHash)
		assert.Equal(t, session.FlowId, retrievedSession.FlowId)
	})

	t.Run("CreateAndGetClientSessionByAccessToken", func(t *testing.T) {
		accessToken, _, err := service.CreateAccessTokenSession(ctx, testTenant, testRealm, testClientID, testUserID, testScope, "client_credentials", 3600)
		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)

		session, err := service.GetClientSessionByAccessToken(ctx, testTenant, testRealm, accessToken)
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, testClientID, session.ClientID)
		assert.Equal(t, testUserID, session.UserID)
		assert.Equal(t, "openid profile", session.Scope)
	})

	t.Run("CreateAndGetClientSessionByRefreshToken", func(t *testing.T) {
		refreshToken, _, err := service.CreateRefreshTokenSession(ctx, testTenant, testRealm, testClientID, testUserID, testScope, "refresh_token", 86400)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshToken)

		session, err := service.LoadAndDeleteRefreshTokenSession(ctx, testTenant, testRealm, refreshToken)
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, testClientID, session.ClientID)
		assert.Equal(t, testUserID, session.UserID)
		assert.Equal(t, "openid profile", session.Scope)

		// Check that the session is deleted
		session, err = service.LoadAndDeleteRefreshTokenSession(ctx, testTenant, testRealm, refreshToken)
		assert.NoError(t, err)
		assert.Nil(t, session)
	})

	t.Run("CreateAndGetAuthCodeSession", func(t *testing.T) {
		flowID := "test-flow"
		loginURI := "/login"
		loginSession, _ := service.CreateAuthSessionObject(testTenant, testRealm, flowID, loginURI)

		authCode, _, err := service.CreateAuthCodeSession(
			ctx,
			testTenant,
			testRealm,
			testClientID,
			testUserID,
			testScope,
			"authorization_code",
			"test-challenge",
			"S256",
			loginSession,
		)
		require.NoError(t, err)
		assert.NotEmpty(t, authCode)

		session, retrievedLoginSession, err := service.LoadAndDeleteAuthCodeSession(ctx, testTenant, testRealm, authCode)
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.NotNil(t, retrievedLoginSession)
		assert.Equal(t, testClientID, session.ClientID)
		assert.Equal(t, testUserID, session.UserID)
		assert.Equal(t, "openid profile", session.Scope)
		assert.Equal(t, loginSession.FlowId, retrievedLoginSession.FlowId)
	})

	t.Run("SessionExpiration", func(t *testing.T) {
		// Create a session that expires in 1 second
		accessToken, _, err := service.CreateAccessTokenSession(ctx, testTenant, testRealm, testClientID, testUserID, testScope, "client_credentials", 1)
		require.NoError(t, err)

		// Advance time by 2 seconds
		mockTime.Advance(2 * time.Second)

		// Try to get the expired session
		session, err := service.GetClientSessionByAccessToken(ctx, testTenant, testRealm, accessToken)
		assert.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("SessionNotExpired", func(t *testing.T) {
		// Create a session that expires in 10 seconds
		accessToken, _, err := service.CreateAccessTokenSession(ctx, testTenant, testRealm, testClientID, testUserID, testScope, "client_credentials", 10)
		require.NoError(t, err)

		// Advance time by 5 seconds
		mockTime.Advance(5 * time.Second)

		// Try to get the session - should still be valid
		session, err := service.GetClientSessionByAccessToken(ctx, testTenant, testRealm, accessToken)
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, testClientID, session.ClientID)
		assert.Equal(t, testUserID, session.UserID)
	})

	t.Run("CreateAndUpdateAuthenticationSession", func(t *testing.T) {
		flowID := "test-flow"
		loginURI := "/login"
		session, sessionID := service.CreateAuthSessionObject(testTenant, testRealm, flowID, loginURI)

		// Create initial session
		err := service.CreateOrUpdateAuthenticationSession(ctx, testTenant, testRealm, *session)
		require.NoError(t, err)

		// Verify initial session
		retrievedSession, exists := service.GetAuthenticationSessionByID(ctx, testTenant, testRealm, sessionID)
		assert.True(t, exists)
		assert.Equal(t, session.SessionIdHash, retrievedSession.SessionIdHash)
		assert.Equal(t, session.FlowId, retrievedSession.FlowId)

		// Update session with new flow ID
		newFlowID := "updated-flow"
		session.FlowId = newFlowID
		err = service.CreateOrUpdateAuthenticationSession(ctx, testTenant, testRealm, *session)
		require.NoError(t, err)

		// Verify updated session
		retrievedSession, exists = service.GetAuthenticationSessionByID(ctx, testTenant, testRealm, sessionID)
		assert.True(t, exists)
		assert.Equal(t, session.SessionIdHash, retrievedSession.SessionIdHash)
		assert.Equal(t, newFlowID, retrievedSession.FlowId)

		// Update session with new context
		session.Context["new_key"] = "new_value"
		err = service.CreateOrUpdateAuthenticationSession(ctx, testTenant, testRealm, *session)
		require.NoError(t, err)

		// Verify context update
		retrievedSession, exists = service.GetAuthenticationSessionByID(ctx, testTenant, testRealm, sessionID)
		assert.True(t, exists)
		assert.Equal(t, "new_value", retrievedSession.Context["new_key"])
	})
}
