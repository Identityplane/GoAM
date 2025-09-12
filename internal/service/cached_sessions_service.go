package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

const (
	// sessionCacheTTL is the time-to-live for session cache entries
	sessionCacheTTL = 5 * time.Second
)

// cachedSessionsService implements SessionsService with caching
type cachedSessionsService struct {
	sessionsService services_interface.SessionsService
	cache           services_interface.CacheService
}

// NewCachedSessionsService creates a new cached sessions service
func NewCachedSessionsService(sessionsService services_interface.SessionsService, cache services_interface.CacheService) services_interface.SessionsService {
	return &cachedSessionsService{
		sessionsService: sessionsService,
		cache:           cache,
	}
}

// SetTimeProvider sets a custom time provider for testing
func (s *cachedSessionsService) SetTimeProvider(provider services_interface.TimeProvider) {
	s.sessionsService.SetTimeProvider(provider)
}

// CreateAuthSessionObject creates a new session object but does not store it
func (s *cachedSessionsService) CreateAuthSessionObject(tenant, realm, flowId, loginUri string) (*model.AuthenticationSession, string) {
	return s.sessionsService.CreateAuthSessionObject(tenant, realm, flowId, loginUri)
}

// CreateOrUpdateAuthenticationSession creates or updates an authentication session
func (s *cachedSessionsService) CreateOrUpdateAuthenticationSession(ctx context.Context, tenant, realm string, session model.AuthenticationSession) error {
	err := s.sessionsService.CreateOrUpdateAuthenticationSession(ctx, tenant, realm, session)
	if err == nil {
		// Cache the session
		cacheKey := fmt.Sprintf("auth_session:%s:%s:%s", tenant, realm, session.SessionIdHash)
		if err := s.cache.Cache(cacheKey, &session, sessionCacheTTL, 1); err != nil {
			log := logger.GetLogger()
			log.Error().Err(err).Msg("failed to cache auth session")
		}
	}
	return err
}

// GetAuthenticationSessionByID retrieves an authentication session by its ID
func (s *cachedSessionsService) GetAuthenticationSessionByID(ctx context.Context, tenant, realm, sessionID string) (*model.AuthenticationSession, bool) {
	return s.sessionsService.GetAuthenticationSessionByID(ctx, tenant, realm, sessionID)
}

// GetAuthenticationSession retrieves an authentication session by its hash
func (s *cachedSessionsService) GetAuthenticationSession(ctx context.Context, tenant, realm, sessionIDHash string) (*model.AuthenticationSession, bool) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("auth_session:%s:%s:%s", tenant, realm, sessionIDHash)
	if cached, found := s.cache.Get(cacheKey); found && cached != nil {
		if session, ok := cached.(*model.AuthenticationSession); ok {
			return session, true
		}
	}

	// If not in cache, get from service
	session, found := s.sessionsService.GetAuthenticationSession(ctx, tenant, realm, sessionIDHash)
	if found {
		// Cache the session
		if err := s.cache.Cache(cacheKey, session, sessionCacheTTL, 1); err != nil {
			log := logger.GetLogger()
			log.Error().Err(err).Msg("failed to cache auth session")
		}
	}
	return session, found
}

// DeleteAuthenticationSession removes an authentication session
func (s *cachedSessionsService) DeleteAuthenticationSession(ctx context.Context, tenant, realm, sessionIDHash string) error {
	err := s.sessionsService.DeleteAuthenticationSession(ctx, tenant, realm, sessionIDHash)
	if err == nil {
		// Remove from cache
		cacheKey := fmt.Sprintf("auth_session:%s:%s:%s", tenant, realm, sessionIDHash)
		if err := s.cache.Invalidate(cacheKey); err != nil {
			log := logger.GetLogger()
			log.Error().Err(err).Msg("failed to remove auth session from cache")
		}
	}
	return err
}

// CreateAuthCodeSession creates a new client session with an auth code
func (s *cachedSessionsService) CreateAuthCodeSession(ctx context.Context, tenant, realm, clientID, userID string, scope []string, grantType string, codeChallenge string, codeChallengeMethod string, loginSession *model.AuthenticationSession) (string, *model.ClientSession, error) {
	code, session, err := s.sessionsService.CreateAuthCodeSession(ctx, tenant, realm, clientID, userID, scope, grantType, codeChallenge, codeChallengeMethod, loginSession)
	if err == nil {
		// Cache the session
		cacheKey := fmt.Sprintf("auth_code_session:%s:%s:%s", tenant, realm, session.ClientSessionID)
		if err := s.cache.Cache(cacheKey, session, sessionCacheTTL, 1); err != nil {
			log := logger.GetLogger()
			log.Error().Err(err).Msg("failed to cache auth code session")
		}
	}
	return code, session, err
}

// CreateAccessTokenSession creates a new access token session
func (s *cachedSessionsService) CreateAccessTokenSession(ctx context.Context, tenant, realm, clientID, userID string, scope []string, grantType string, lifetime int) (string, *model.ClientSession, error) {
	token, session, err := s.sessionsService.CreateAccessTokenSession(ctx, tenant, realm, clientID, userID, scope, grantType, lifetime)
	if err == nil {
		// Cache the session
		cacheKey := fmt.Sprintf("access_token_session:%s:%s:%s", tenant, realm, session.ClientSessionID)
		if err := s.cache.Cache(cacheKey, session, sessionCacheTTL, 1); err != nil {
			log := logger.GetLogger()
			log.Error().Err(err).Msg("failed to cache access token session")
		}
	}
	return token, session, err
}

// CreateRefreshTokenSession creates a new refresh token session
func (s *cachedSessionsService) CreateRefreshTokenSession(ctx context.Context, tenant, realm, clientID, userID string, scope []string, grantType string, expiresIn int) (string, *model.ClientSession, error) {
	return s.sessionsService.CreateRefreshTokenSession(ctx, tenant, realm, clientID, userID, scope, grantType, expiresIn)
}

// GetClientSessionByAccessToken retrieves a client session by its access token
func (s *cachedSessionsService) GetClientSessionByAccessToken(ctx context.Context, tenant, realm, accessToken string) (*model.ClientSession, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("access_token:%s:%s:%s", tenant, realm, accessToken)
	if cached, found := s.cache.Get(cacheKey); found && cached != nil {
		if session, ok := cached.(*model.ClientSession); ok {
			return session, nil
		}
	}

	// If not in cache, get from service
	session, err := s.sessionsService.GetClientSessionByAccessToken(ctx, tenant, realm, accessToken)
	if err == nil && session != nil {
		// Cache the session
		if err := s.cache.Cache(cacheKey, session, sessionCacheTTL, 1); err != nil {
			log := logger.GetLogger()
			log.Error().Err(err).Msg("failed to cache access token session")
		}
	}
	return session, err
}

// LoadAndDeleteAuthCodeSession retrieves a client session by auth code and deletes it
// Not cached as we must delete it from the database after retrieving it
func (s *cachedSessionsService) LoadAndDeleteAuthCodeSession(ctx context.Context, tenant, realm, authCode string) (*model.ClientSession, *model.AuthenticationSession, error) {
	return s.sessionsService.LoadAndDeleteAuthCodeSession(ctx, tenant, realm, authCode)
}

// LoadAndDeleteRefreshTokenSession retrieves a client session by refresh token and deletes it
// Not cached as we must delete it from the database after retrieving it
func (s *cachedSessionsService) LoadAndDeleteRefreshTokenSession(ctx context.Context, tenant, realm, refreshToken string) (*model.ClientSession, error) {
	return s.sessionsService.LoadAndDeleteRefreshTokenSession(ctx, tenant, realm, refreshToken)
}

// cachedAuthSessionDB implements AuthSessionDB with caching
type cachedAuthSessionDB struct {
	authSessionDB db.AuthSessionDB
	cache         services_interface.CacheService
}

// getAuthSessionCacheKey returns a cache key for an auth session
func (c *cachedAuthSessionDB) getAuthSessionCacheKey(tenant, realm, sessionIDHash string) string {
	return fmt.Sprintf("/%s/%s/auth-session/%s", tenant, realm, sessionIDHash)
}

func (c *cachedAuthSessionDB) CreateOrUpdateAuthSession(ctx context.Context, session *model.PersistentAuthSession) error {
	// Create/update in database
	err := c.authSessionDB.CreateOrUpdateAuthSession(ctx, session)
	if err != nil {
		return err
	}

	// Cache the session
	cacheKey := c.getAuthSessionCacheKey(session.Tenant, session.Realm, session.SessionIDHash)
	err = c.cache.Cache(cacheKey, session, sessionCacheTTL, 1)
	if err != nil {
		// Log error but continue - caching is not critical
		log := logger.GetLogger()
		log.Error().Err(err).Msg("failed to cache auth session")
	}

	return nil
}

func (c *cachedAuthSessionDB) GetAuthSessionByID(ctx context.Context, tenant, realm, runID string) (*model.PersistentAuthSession, error) {
	// Get from database - no caching for ID lookup
	return c.authSessionDB.GetAuthSessionByID(ctx, tenant, realm, runID)
}

func (c *cachedAuthSessionDB) GetAuthSessionByHash(ctx context.Context, tenant, realm, sessionIDHash string) (*model.PersistentAuthSession, error) {
	// Try to get from cache first
	cacheKey := c.getAuthSessionCacheKey(tenant, realm, sessionIDHash)
	if cached, exists := c.cache.Get(cacheKey); exists {
		if session, ok := cached.(*model.PersistentAuthSession); ok {
			return session, nil
		}
	}

	// If not in cache, get from database
	session, err := c.authSessionDB.GetAuthSessionByHash(ctx, tenant, realm, sessionIDHash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	// Cache the result
	err = c.cache.Cache(cacheKey, session, sessionCacheTTL, 1)
	if err != nil {
		// Log error but continue - caching is not critical
		log := logger.GetLogger()
		log.Error().Err(err).Msg("failed to cache auth session")
	}

	return session, nil
}

func (c *cachedAuthSessionDB) ListAuthSessions(ctx context.Context, tenant, realm string) ([]model.PersistentAuthSession, error) {
	// Direct call to database - no caching for list operations
	return c.authSessionDB.ListAuthSessions(ctx, tenant, realm)
}

func (c *cachedAuthSessionDB) ListAllAuthSessions(ctx context.Context, tenant string) ([]model.PersistentAuthSession, error) {
	// Direct call to database - no caching for list operations
	return c.authSessionDB.ListAllAuthSessions(ctx, tenant)
}

func (c *cachedAuthSessionDB) DeleteAuthSession(ctx context.Context, tenant, realm, sessionIDHash string) error {
	// Delete from database
	err := c.authSessionDB.DeleteAuthSession(ctx, tenant, realm, sessionIDHash)
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := c.getAuthSessionCacheKey(tenant, realm, sessionIDHash)
	c.cache.Invalidate(cacheKey)

	return nil
}

func (c *cachedAuthSessionDB) DeleteExpiredAuthSessions(ctx context.Context, tenant, realm string) error {
	// Direct call to database - no caching for cleanup operations
	return c.authSessionDB.DeleteExpiredAuthSessions(ctx, tenant, realm)
}

// cachedClientSessionDB implements ClientSessionDB with caching for auth code and access token sessions
type cachedClientSessionDB struct {
	clientSessionDB db.ClientSessionDB
	cache           services_interface.CacheService
}

// getClientSessionCacheKey returns a cache key for a client session
func (c *cachedClientSessionDB) getClientSessionCacheKey(tenant, realm, tokenHash string, tokenType string) string {
	return fmt.Sprintf("/%s/%s/client-session/%s/%s", tenant, realm, tokenType, tokenHash)
}

func (c *cachedClientSessionDB) CreateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error {
	// Create in database
	err := c.clientSessionDB.CreateClientSession(ctx, tenant, realm, session)
	if err != nil {
		return err
	}

	// Cache if it's an auth code or access token session
	if session.AuthCodeHash != "" {
		cacheKey := c.getClientSessionCacheKey(tenant, realm, session.AuthCodeHash, "auth-code")
		err = c.cache.Cache(cacheKey, session, sessionCacheTTL, 1)
		if err != nil {
			log := logger.GetLogger()
			log.Info().Err(err).Msg("failed to cache auth code session")
		}
	}
	if session.AccessTokenHash != "" {
		cacheKey := c.getClientSessionCacheKey(tenant, realm, session.AccessTokenHash, "access-token")
		err = c.cache.Cache(cacheKey, session, sessionCacheTTL, 1)
		if err != nil {
			log := logger.GetLogger()
			log.Info().Err(err).Msg("failed to cache access token session")
		}
	}

	return nil
}

func (c *cachedClientSessionDB) GetClientSessionByID(ctx context.Context, tenant, realm, sessionID string) (*model.ClientSession, error) {
	// Direct call to database - no caching for ID lookup
	return c.clientSessionDB.GetClientSessionByID(ctx, tenant, realm, sessionID)
}

func (c *cachedClientSessionDB) GetClientSessionByAccessToken(ctx context.Context, tenant, realm, accessTokenHash string) (*model.ClientSession, error) {
	// Try to get from cache first
	cacheKey := c.getClientSessionCacheKey(tenant, realm, accessTokenHash, "access-token")
	if cached, exists := c.cache.Get(cacheKey); exists {
		if session, ok := cached.(*model.ClientSession); ok {
			return session, nil
		}
	}

	// If not in cache, get from database
	session, err := c.clientSessionDB.GetClientSessionByAccessToken(ctx, tenant, realm, accessTokenHash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	// Cache the result
	err = c.cache.Cache(cacheKey, session, sessionCacheTTL, 1)
	if err != nil {
		log := logger.GetLogger()
		log.Info().Err(err).Msg("failed to cache access token session")
	}

	return session, nil
}

func (c *cachedClientSessionDB) GetClientSessionByAuthCode(ctx context.Context, tenant, realm, authCodeHash string) (*model.ClientSession, error) {
	// Try to get from cache first
	cacheKey := c.getClientSessionCacheKey(tenant, realm, authCodeHash, "auth-code")
	if cached, exists := c.cache.Get(cacheKey); exists {
		if session, ok := cached.(*model.ClientSession); ok {
			return session, nil
		}
	}

	// If not in cache, get from database
	session, err := c.clientSessionDB.GetClientSessionByAuthCode(ctx, tenant, realm, authCodeHash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	// Cache the result
	err = c.cache.Cache(cacheKey, session, sessionCacheTTL, 1)
	if err != nil {
		log := logger.GetLogger()
		log.Info().Err(err).Msg("failed to cache auth code session")
	}

	return session, nil
}

func (c *cachedClientSessionDB) GetClientSessionByRefreshToken(ctx context.Context, tenant, realm, refreshTokenHash string) (*model.ClientSession, error) {
	// Direct call to database - no caching for refresh tokens
	return c.clientSessionDB.GetClientSessionByRefreshToken(ctx, tenant, realm, refreshTokenHash)
}

func (c *cachedClientSessionDB) ListClientSessions(ctx context.Context, tenant, realm, clientID string) ([]model.ClientSession, error) {
	// Direct call to database - no caching for list operations
	return c.clientSessionDB.ListClientSessions(ctx, tenant, realm, clientID)
}

func (c *cachedClientSessionDB) ListUserClientSessions(ctx context.Context, tenant, realm, userID string) ([]model.ClientSession, error) {
	// Direct call to database - no caching for list operations
	return c.clientSessionDB.ListUserClientSessions(ctx, tenant, realm, userID)
}

func (c *cachedClientSessionDB) UpdateClientSession(ctx context.Context, tenant, realm string, session *model.ClientSession) error {
	// Update in database
	err := c.clientSessionDB.UpdateClientSession(ctx, tenant, realm, session)
	if err != nil {
		return err
	}

	// Invalidate relevant caches
	if session.AuthCodeHash != "" {
		cacheKey := c.getClientSessionCacheKey(tenant, realm, session.AuthCodeHash, "auth-code")
		c.cache.Invalidate(cacheKey)
	}
	if session.AccessTokenHash != "" {
		cacheKey := c.getClientSessionCacheKey(tenant, realm, session.AccessTokenHash, "access-token")
		c.cache.Invalidate(cacheKey)
	}

	return nil
}

func (c *cachedClientSessionDB) DeleteClientSession(ctx context.Context, tenant, realm, sessionID string) error {
	// Get session first to know what to invalidate
	session, err := c.clientSessionDB.GetClientSessionByID(ctx, tenant, realm, sessionID)
	if err != nil {
		return err
	}

	// Delete from database
	err = c.clientSessionDB.DeleteClientSession(ctx, tenant, realm, sessionID)
	if err != nil {
		return err
	}

	// Invalidate relevant caches
	if session != nil {
		if session.AuthCodeHash != "" {
			cacheKey := c.getClientSessionCacheKey(tenant, realm, session.AuthCodeHash, "auth-code")
			c.cache.Invalidate(cacheKey)
		}
		if session.AccessTokenHash != "" {
			cacheKey := c.getClientSessionCacheKey(tenant, realm, session.AccessTokenHash, "access-token")
			c.cache.Invalidate(cacheKey)
		}
	}

	return nil
}

func (c *cachedClientSessionDB) DeleteExpiredClientSessions(ctx context.Context, tenant, realm string) error {
	// Direct call to database - no caching for cleanup operations
	return c.clientSessionDB.DeleteExpiredClientSessions(ctx, tenant, realm)
}
