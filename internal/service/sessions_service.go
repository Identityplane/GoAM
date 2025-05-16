package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"goiam/internal/db"
	"goiam/internal/lib"
	"goiam/internal/logger"
	"goiam/internal/model"

	"github.com/google/uuid"
)

// TimeProvider defines an interface for getting the current time
type TimeProvider interface {
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the system clock
type RealTimeProvider struct{}

func (r *RealTimeProvider) Now() time.Time {
	return time.Now()
}

type SessionsService struct {
	mu              sync.RWMutex
	clientSessionDB db.ClientSessionDB
	authSessionDB   db.AuthSessionDB
	timeProvider    TimeProvider
}

func NewSessionsService(clientSessionDB db.ClientSessionDB, authSessionDB db.AuthSessionDB) *SessionsService {
	return &SessionsService{
		clientSessionDB: clientSessionDB,
		authSessionDB:   authSessionDB,
		timeProvider:    &RealTimeProvider{},
	}
}

// SetTimeProvider sets a custom time provider for testing
func (s *SessionsService) SetTimeProvider(provider TimeProvider) {
	s.timeProvider = provider
}

// Creates a new session object but does not store it in the database yet
// This is to optimize performance so that only one database call is made when the session is created
// returns the session object and session id
func (s *SessionsService) CreateAuthSessionObject(tenant, realm, flowId, loginUri string) (*model.AuthenticationSession, string) {
	sessionID := lib.GenerateSecureSessionID()

	session := &model.AuthenticationSession{
		RunID:                    uuid.New().String(),
		FlowId:                   flowId,
		SessionIdHash:            lib.HashString(sessionID),
		Context:                  make(map[string]string),
		History:                  make([]string, 0),
		Prompts:                  make(map[string]string),
		Oauth2SessionInformation: nil,
		CreatedAt:                s.timeProvider.Now(),
		ExpiresAt:                s.timeProvider.Now().Add(30 * time.Minute), // 30 minutes expiration TODO make this variable by realm
		LoginUri:                 loginUri,
	}

	return session, sessionID
}

// CreateOrUpdateAuthenticationSession creates a new authentication session or updates an existing one
func (s *SessionsService) CreateOrUpdateAuthenticationSession(ctx context.Context, tenant, realm string, session model.AuthenticationSession) error {
	persistentSession, err := model.NewPersistentAuthSession(tenant, realm, &session)
	if err != nil {
		return fmt.Errorf("failed to create persistent session: %w", err)
	}

	return s.authSessionDB.CreateOrUpdateAuthSession(ctx, persistentSession)
}

func (s *SessionsService) GetAuthenticationSessionByID(ctx context.Context, tenant, realm, sessionID string) (*model.AuthenticationSession, bool) {
	// Hash the session id
	sessionIDHash := lib.HashString(sessionID)
	return s.GetAuthenticationSession(ctx, tenant, realm, sessionIDHash)
}

// GetAuthenticationSession retrieves an authentication session by its hash
func (s *SessionsService) GetAuthenticationSession(ctx context.Context, tenant, realm, sessionIDHash string) (*model.AuthenticationSession, bool) {
	persistentSession, err := s.authSessionDB.GetAuthSessionByHash(ctx, tenant, realm, sessionIDHash)
	if err != nil {
		logger.ErrorNoContext("Failed to get auth session: %v", err)
		return nil, false
	}
	if persistentSession == nil {
		return nil, false
	}

	// Check if session has expired
	if s.timeProvider.Now().After(persistentSession.ExpiresAt) {
		// Delete expired session
		err := s.authSessionDB.DeleteAuthSession(ctx, tenant, realm, sessionIDHash)
		if err != nil {
			logger.ErrorNoContext("Failed to delete expired session: %v", err)
		}
		return nil, false
	}

	session, err := persistentSession.ToAuthenticationSession()
	if err != nil {
		logger.ErrorNoContext("Failed to convert persistent session to auth session: %v", err)
		return nil, false
	}

	return session, true
}

// DeleteAuthenticationSession removes an authentication session
func (s *SessionsService) DeleteAuthenticationSession(ctx context.Context, tenant, realm, sessionIDHash string) error {
	return s.authSessionDB.DeleteAuthSession(ctx, tenant, realm, sessionIDHash)
}

// CreateClientSession creates a new client session
func (s *SessionsService) CreateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error {

	panic("not implemented")

	return s.clientSessionDB.CreateClientSession(ctx, tenant, realm, session)
}

// CreateAuthCodeSession creates a new client session with an auth code
func (s *SessionsService) CreateAuthCodeSession(ctx context.Context, tenant, realm, clientID, userID string, scope []string, grantType string, codeChallenge string, codeChallengeMethod string, loginSession *model.AuthenticationSession) (string, error) {
	// Generate a new auth code
	sessionID := lib.GenerateSecureSessionID()
	authCode := lib.GenerateSecureSessionID()
	authCodeHash := lib.HashString(authCode)

	// json encode the login session
	loginSessionJSON, err := json.Marshal(loginSession)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login session: %w", err)
	}

	// Create a new client session
	session := &model.ClientSession{
		Tenant:              tenant,
		Realm:               realm,
		ClientSessionID:     sessionID,
		ClientID:            clientID,
		GrantType:           grantType,
		AuthCodeHash:        authCodeHash,
		UserID:              userID,
		Scope:               strings.Join(scope, " "),
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		LoginSessionJson:    string(loginSessionJSON),
		Created:             time.Now(),
		Expire:              time.Now().Add(10 * time.Minute), // Auth codes typically expire in 10 minutes recommended by OAuth 2.1
	}

	// Store the session in the database
	err = s.clientSessionDB.CreateClientSession(ctx, tenant, realm, session)
	if err != nil {
		return "", err
	}

	return authCode, nil
}

func (s *SessionsService) CreateAccessTokenSession(ctx context.Context, tenant, realm, clientID, userID string, scope []string, grantType string, lifetime int) (string, error) {

	sessionID := lib.GenerateSecureSessionID()
	accessToken := lib.GenerateSecureSessionID()
	accessTokenHash := lib.HashString(accessToken)

	session := &model.ClientSession{
		Tenant:          tenant,
		Realm:           realm,
		ClientSessionID: sessionID,
		ClientID:        clientID,
		GrantType:       grantType,
		AccessTokenHash: accessTokenHash,
		UserID:          userID,
		Scope:           strings.Join(scope, " "),
		Created:         time.Now(),
		Expire:          time.Now().Add(time.Duration(lifetime) * time.Second),
	}

	err := s.clientSessionDB.CreateClientSession(ctx, tenant, realm, session)
	if err != nil {
		return "", fmt.Errorf("failed to create access token session: %w", err)
	}

	logger.InfoNoContext("Creating client access token session tenant=%s realm=%s id=%s hash=%s", tenant, realm, sessionID[:8], accessTokenHash[:8])

	return accessToken, nil
}

// CreateRefreshTokenSession creates a new refresh token session
func (s *SessionsService) CreateRefreshTokenSession(context context.Context, tenant, realm, clientID, userID string, scope []string, grantType string, expiresIn int) (string, error) {

	sessionID := lib.GenerateSecureSessionID()
	refreshToken := lib.GenerateSecureSessionID()
	refreshTokenHash := lib.HashString(refreshToken)

	session := &model.ClientSession{
		Tenant:           tenant,
		Realm:            realm,
		ClientSessionID:  sessionID,
		ClientID:         clientID,
		GrantType:        grantType,
		RefreshTokenHash: refreshTokenHash,
		UserID:           userID,
		Scope:            strings.Join(scope, " "),
		Created:          time.Now(),
		Expire:           time.Now().Add(time.Duration(expiresIn) * time.Second),
	}

	err := s.clientSessionDB.CreateClientSession(context, tenant, realm, session)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token session: %w", err)
	}

	logger.InfoNoContext("Creating client refresh token session tenant=%s realm=%s id=%s hash=%s", tenant, realm, sessionID[:8], refreshTokenHash[:8])

	return refreshToken, nil
}

// GetClientSessionByAccessToken retrieves a client session by its access token
func (s *SessionsService) GetClientSessionByAccessToken(ctx context.Context, tenant, realm, accessToken string) (*model.ClientSession, error) {
	accessTokenHash := lib.HashString(accessToken)

	logger.DebugNoContext("Getting client session by access token tenant=%s realm=%s hash=%s", tenant, realm, accessTokenHash[:8])

	session, err := s.clientSessionDB.GetClientSessionByAccessToken(ctx, tenant, realm, accessTokenHash)
	if err != nil {
		return nil, err
	}

	// Check if session has expired
	if s.timeProvider.Now().After(session.Expire) {
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// LoadAndDeleteAuthCodeSession retrieves a client session by auth code and deletes it
func (s *SessionsService) LoadAndDeleteAuthCodeSession(ctx context.Context, tenant, realm, authCode string) (*model.ClientSession, *model.AuthenticationSession, error) {
	// Hash the auth code
	authCodeHash := lib.HashString(authCode)

	// Get the session by auth code hash
	session, err := s.clientSessionDB.GetClientSessionByAuthCode(ctx, tenant, realm, authCodeHash)
	if err != nil {
		return nil, nil, err
	}
	if session == nil {
		return nil, nil, nil
	}

	// Delete the session
	err = s.clientSessionDB.DeleteClientSession(ctx, tenant, realm, session.ClientSessionID)
	if err != nil {
		return nil, nil, err
	}

	// Check if the session has expired
	if s.timeProvider.Now().After(session.Expire) {
		return nil, nil, fmt.Errorf("session expired")
	}

	// json decode the login session
	var loginSession model.AuthenticationSession
	err = json.Unmarshal([]byte(session.LoginSessionJson), &loginSession)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal login session: %w", err)
	}

	return session, &loginSession, nil
}

// LoadAndDeleteRefreshTokenSession retrieves a client session by refresh token and deletes it
func (s *SessionsService) LoadAndDeleteRefreshTokenSession(ctx context.Context, tenant, realm, refreshToken string) (*model.ClientSession, error) {

	// Hash the refresh token
	refreshTokenHash := lib.HashString(refreshToken)

	// Get the session by refresh token hash
	session, err := s.clientSessionDB.GetClientSessionByRefreshToken(ctx, tenant, realm, refreshTokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get client session by refresh token: %w", err)
	}

	if session == nil {
		return nil, nil
	}

	err = s.clientSessionDB.DeleteClientSession(ctx, tenant, realm, session.ClientSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete client session: %w", err)
	}

	// Check if the session has expired
	if time.Now().After(session.Expire) {
		return nil, nil
	}

	return session, nil
}
