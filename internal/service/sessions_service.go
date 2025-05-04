package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"goiam/internal/db"
	"goiam/internal/lib"
	"goiam/internal/model"

	"github.com/google/uuid"
)

type SessionsService struct {
	mu              sync.RWMutex
	sessions        map[string]*model.AuthenticationSession
	clientSessionDB db.ClientSessionDB
}

func NewSessionsService(clientSessionDB db.ClientSessionDB) *SessionsService {
	return &SessionsService{
		sessions:        make(map[string]*model.AuthenticationSession),
		clientSessionDB: clientSessionDB,
	}
}

// Creates a new session object but does not store it in the database yet
// This is to optimize performance so that only one database call is made when the session is created
// returns the session object and session id
func (s *SessionsService) CreateSessionObject(tenant, realm, flowId, loginUri string) (*model.AuthenticationSession, string) {

	sessionID := lib.GenerateSessionID()

	session := &model.AuthenticationSession{
		RunID:                    uuid.New().String(),
		FlowId:                   flowId,
		SessionIdHash:            lib.HashString(sessionID),
		Context:                  make(map[string]string),
		History:                  make([]string, 0),
		Prompts:                  make(map[string]string),
		Oauth2SessionInformation: nil,
		ExpiresAt:                time.Now().Add(30 * time.Minute), // 30 minutes expiration TODO make this variable by realm
		LoginUri:                 loginUri,
	}

	return session, sessionID
}

// CreateOrUpdateAuthenticationSession creates a new authentication session or updates an existing one
func (s *SessionsService) CreateOrUpdateAuthenticationSession(tenant, realm string, session model.AuthenticationSession) error {

	cashKey := tenant + ":" + realm + ":" + session.SessionIdHash
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[cashKey] = &session

	return nil
}

func (s *SessionsService) GetAuthenticationSessionByID(tenant, realm, sessionID string) (*model.AuthenticationSession, bool) {

	// Hash the session id
	sessionIDHash := lib.HashString(sessionID)

	return s.GetAuthenticationSession(tenant, realm, sessionIDHash)
}

// GetAuthenticationSession retrieves an authentication session by its hash
func (s *SessionsService) GetAuthenticationSession(tenant, realm, sessionIDHash string) (*model.AuthenticationSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cashKey := tenant + ":" + realm + ":" + sessionIDHash
	session, ok := s.sessions[cashKey]
	if !ok {
		return nil, false
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, cashKey)
		s.mu.Unlock()
		return nil, false
	}

	return session, true
}

// DeleteAuthenticationSession removes an authentication session
func (s *SessionsService) DeleteAuthenticationSession(sessionIDHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionIDHash)
}

// CreateClientSession creates a new client session
func (s *SessionsService) CreateClientSession(ctx context.Context, session *model.ClientSession) error {
	return s.clientSessionDB.CreateClientSession(ctx, session)
}

// CreateAuthCodeSession creates a new client session with an auth code
func (s *SessionsService) CreateAuthCodeSession(ctx context.Context, tenant, realm, clientID, userID string, scope []string, grantType string) (string, error) {
	// Generate a new auth code
	sessionID := lib.GenerateSessionID()
	authCode := lib.GenerateSessionID()
	authCodeHash := lib.HashString(authCode)

	// Create a new client session
	session := &model.ClientSession{
		Tenant:          tenant,
		Realm:           realm,
		ClientSessionID: sessionID,
		ClientID:        clientID,
		GrantType:       grantType,
		AuthCodeHash:    authCodeHash,
		UserID:          userID,
		Scope:           strings.Join(scope, " "),
		Created:         time.Now(),
		Expire:          time.Now().Add(10 * time.Minute), // Auth codes typically expire in 10 minutes recommended by OAuth 2.1
	}

	// Store the session in the database
	err := s.clientSessionDB.CreateClientSession(ctx, session)
	if err != nil {
		return "", err
	}

	return authCode, nil
}

// LoadAndDeleteAuthCodeSession retrieves a client session by auth code and deletes it
func (s *SessionsService) LoadAndDeleteAuthCodeSession(ctx context.Context, authCode string) (*model.ClientSession, error) {
	// Hash the auth code
	authCodeHash := lib.HashString(authCode)

	// Get the session by auth code hash
	session, err := s.clientSessionDB.GetClientSessionByAuthCode(ctx, authCodeHash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	// Delete the session
	err = s.clientSessionDB.DeleteClientSession(ctx, session.Tenant, session.Realm, session.ClientSessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}
