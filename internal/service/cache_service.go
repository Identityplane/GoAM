package service

import (
	"fmt"
	"time"

	services_interface "github.com/Identityplane/GoAM/pkg/services"
	"github.com/dgraph-io/ristretto/v2"
)

// cacheServiceImpl implements CacheService
type cacheServiceImpl struct {
	cache *ristretto.Cache[string, interface{}]
}

// NewCacheService creates a new CacheService instance
func NewCacheService() (services_interface.CacheService, error) {
	// Configure Ristretto cache
	config := &ristretto.Config[string, interface{}]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
		Metrics:     true,
	}

	cache, err := ristretto.NewCache(config)
	if err != nil {
		return nil, err
	}

	return &cacheServiceImpl{
		cache: cache,
	}, nil
}

func (s *cacheServiceImpl) Cache(key string, value interface{}, ttl time.Duration, cost int64) error {
	success := s.cache.SetWithTTL(key, value, cost, ttl)
	if !success {
		return ErrCacheSetFailed
	}

	s.cache.Wait()

	return nil
}

func (s *cacheServiceImpl) Get(key string) (interface{}, bool) {
	value, found := s.cache.Get(key)
	return value, found
}

func (s *cacheServiceImpl) Invalidate(key string) error {
	s.cache.Del(key)
	return nil
}

func (s *cacheServiceImpl) GetMetrics() services_interface.CacheMetrics {

	ratio := s.cache.Metrics.Ratio()
	hits := s.cache.Metrics.Hits()
	misses := s.cache.Metrics.Misses()
	keysAdded := s.cache.Metrics.KeysAdded()

	return services_interface.CacheMetrics{
		Ratio:     ratio,
		Hits:      hits,
		Misses:    misses,
		KeysAdded: keysAdded,
	}
}

// ErrCacheSetFailed is returned when setting a value in the cache fails
var ErrCacheSetFailed = fmt.Errorf("failed to set value in cache")

// ErrCacheInvalidateFailed is returned when invalidating a key from the cache fails
var ErrCacheInvalidateFailed = fmt.Errorf("failed to invalidate key from cache")
