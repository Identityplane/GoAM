package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

const (
	// signingKeyCacheTTL is the time-to-live for signing key cache entries
	signingKeyCacheTTL = 5 * time.Second
	// jwksCacheTTL is the time-to-live for JWKS cache entries
	jwksCacheTTL = 5 * time.Second
)

// cachedJWTService implements JWTService with caching
type cachedJWTService struct {
	jwtService services_interface.JWTService
	cache      services_interface.CacheService
}

// NewCachedJWTService creates a new cached JWT service
func NewCachedJWTService(jwtService services_interface.JWTService, cache services_interface.CacheService) services_interface.JWTService {
	return &cachedJWTService{
		jwtService: jwtService,
		cache:      cache,
	}
}

// getSigningKeyCacheKey returns a cache key in the format /<tenant>/<realm>/signing-key
func (s *cachedJWTService) getSigningKeyCacheKey(tenant, realm string) string {
	return fmt.Sprintf("/%s/%s/signing-key", tenant, realm)
}

// getJWKSCacheKey returns a cache key in the format /<tenant>/<realm>/jwks
func (s *cachedJWTService) getJWKSCacheKey(tenant, realm string) string {
	return fmt.Sprintf("/%s/%s/jwks", tenant, realm)
}

// LoadPublicKeys returns the JWKS for a given tenant and realm
func (s *cachedJWTService) LoadPublicKeys(tenant, realm string) (string, error) {
	// Try to get from cache first
	cacheKey := s.getJWKSCacheKey(tenant, realm)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if jwks, ok := cached.(string); ok {
			return jwks, nil
		}
	}

	// If not in cache, get from service
	jwks, err := s.jwtService.LoadPublicKeys(tenant, realm)
	if err != nil {
		return "", err
	}

	// Cache the result
	err = s.cache.Cache(cacheKey, jwks, jwksCacheTTL, 1)
	if err != nil {
		// Log error but continue - caching is not critical
		log := logger.GetLogger()
		log.Info().Err(err).Msg("failed to cache jwks")
	}

	return jwks, nil
}

// SignJWT signs a JWT token with the key for the given tenant and realm
func (s *cachedJWTService) SignJWT(tenant, realm string, claims map[string]interface{}) (string, error) {
	return s.jwtService.SignJWT(tenant, realm, claims)
}

// GenerateKey generates a new key for a tenant/realm
func (s *cachedJWTService) GenerateKey(tenant, realm string) error {
	err := s.jwtService.GenerateKey(tenant, realm)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(tenant, realm)
	return nil
}

// RotateKey generates a new key and disables the old one
func (s *cachedJWTService) RotateKey(tenant, realm string) error {
	err := s.jwtService.RotateKey(tenant, realm)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(tenant, realm)
	return nil
}

// invalidateCaches invalidates all relevant cache entries
func (s *cachedJWTService) invalidateCaches(tenant, realm string) {
	// Invalidate signing key cache
	keyCacheKey := s.getSigningKeyCacheKey(tenant, realm)
	s.cache.Invalidate(keyCacheKey)

	// Invalidate JWKS cache
	jwksCacheKey := s.getJWKSCacheKey(tenant, realm)
	s.cache.Invalidate(jwksCacheKey)
}

// GetActiveSigningKey returns an active signing key for the given tenant and realm
func (s *cachedJWTService) GetActiveSigningKey(ctx context.Context, tenant, realm string) (*model.SigningKey, error) {
	// Try to get from cache first
	cacheKey := s.getSigningKeyCacheKey(tenant, realm)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if key, ok := cached.(*model.SigningKey); ok {
			return key, nil
		}
	}

	// If not in cache, get from service
	key, err := s.jwtService.GetActiveSigningKey(ctx, tenant, realm)
	if err != nil {
		return nil, err
	}

	// Cache the result
	err = s.cache.Cache(cacheKey, key, signingKeyCacheTTL, 1)
	if err != nil {
		// Log error but continue - caching is not critical
		log := logger.GetLogger()
		log.Info().Err(err).Msg("failed to cache signing key")
	}

	return key, nil
}
