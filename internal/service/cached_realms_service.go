package service

import (
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/model"
)

const (
	// realmCacheTTL is the time-to-live for realm cache entries
	realmCacheTTL = 10 * time.Second
)

// cachedRealmService implements RealmService with caching
type cachedRealmService struct {
	realmService RealmService
	cache        CacheService
}

// NewCachedRealmService creates a new cached realm service
func NewCachedRealmService(realmService RealmService, cache CacheService) RealmService {
	return &cachedRealmService{
		realmService: realmService,
		cache:        cache,
	}
}

// getCacheKey returns a cache key in the format /<tenant>/<realm>/realm
func (s *cachedRealmService) getCacheKey(tenant, realm string) string {
	return fmt.Sprintf("/%s/%s/realm", tenant, realm)
}

func (s *cachedRealmService) GetRealm(tenant, realm string) (*LoadedRealm, bool) {
	// Try to get from cache first
	cacheKey := s.getCacheKey(tenant, realm)
	if cached, found := s.cache.Get(cacheKey); found && cached != nil {
		if loadedRealm, ok := cached.(*LoadedRealm); ok {
			return loadedRealm, true
		}
	}

	// If not in cache, get from service
	loadedRealm, found := s.realmService.GetRealm(tenant, realm)
	if found {
		// Cache the realm
		if err := s.cache.Cache(cacheKey, loadedRealm, realmCacheTTL, 1); err != nil {
			log := logger.GetLogger()
			log.Info().Err(err).Msg("failed to cache realm")
		}
	}
	return loadedRealm, found
}

func (s *cachedRealmService) GetAllRealms() (map[string]*LoadedRealm, error) {
	// Direct call to service without caching for admin operations
	return s.realmService.GetAllRealms()
}

func (s *cachedRealmService) CreateRealm(realm *model.Realm) error {
	// Create realm
	err := s.realmService.CreateRealm(realm)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(realm.Tenant, realm.Realm)
	return nil
}

func (s *cachedRealmService) UpdateRealm(realm *model.Realm) error {
	// Update realm
	err := s.realmService.UpdateRealm(realm)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(realm.Tenant, realm.Realm)
	return nil
}

func (s *cachedRealmService) DeleteRealm(tenant, realm string) error {
	// Delete realm
	err := s.realmService.DeleteRealm(tenant, realm)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(tenant, realm)
	return nil
}

// invalidateCaches invalidates all relevant cache entries
func (s *cachedRealmService) invalidateCaches(tenant, realm string) {
	// Invalidate specific realm cache
	realmKey := s.getCacheKey(tenant, realm)
	s.cache.Invalidate(realmKey)
}

// This is not cached
func (s *cachedRealmService) IsTenantNameAvailable(tenantName string) (bool, error) {
	return s.realmService.IsTenantNameAvailable(tenantName)
}
