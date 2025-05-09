package service

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// CacheService defines the interface for cache operations
// This is currently in memory only but can be extended to other cache backend such as Redis
type CacheService interface {
	// Cache stores a value in the cache with the specified TTL
	Cache(key string, value interface{}, ttl time.Duration, cost int64) error

	// Get retrieves a value from the cache by its key
	Get(key string) (interface{}, bool)

	// Invalidate removes a key from the cache
	Invalidate(key string) error

	// GetMetrics returns the metrics of the cache
	GetMetrics() CacheMetrics
}

// cacheServiceImpl implements CacheService
type cacheServiceImpl struct {
	cache *ristretto.Cache[string, interface{}]
}

// NewCacheService creates a new CacheService instance
func NewCacheService() (CacheService, error) {
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

type CacheMetrics struct {
	Ratio     float64
	Hits      uint64
	Misses    uint64
	KeysAdded uint64
}

func (s *cacheServiceImpl) GetMetrics() CacheMetrics {

	ratio := s.cache.Metrics.Ratio()
	hits := s.cache.Metrics.Hits()
	misses := s.cache.Metrics.Misses()
	keysAdded := s.cache.Metrics.KeysAdded()

	return CacheMetrics{
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
