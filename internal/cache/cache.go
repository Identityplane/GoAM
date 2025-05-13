package cache

import "time"

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(key string) (interface{}, bool)

	// Cache stores a value in the cache with the given TTL and priority
	Cache(key string, value interface{}, ttl time.Duration, priority int) error

	// Invalidate removes a value from the cache
	Invalidate(key string)
}
