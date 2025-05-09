package service

import (
	"testing"
	"time"
)

func TestCacheService(t *testing.T) {
	// Create a new cache service instance
	cache, err := NewCacheService()
	if err != nil {
		t.Fatalf("Failed to create cache service: %v", err)
	}

	// Test case 1: Get non-existing key
	t.Run("Get non-existing key", func(t *testing.T) {
		value, exists := cache.Get("non-existing-key")
		if exists {
			t.Errorf("Expected non-existing key to return false, got true")
		}
		if value != nil {
			t.Errorf("Expected non-existing key to return nil value, got %v", value)
		}
	})

	// Test case 2: Add and retrieve an object
	t.Run("Add and retrieve object", func(t *testing.T) {
		testKey := "test-key"
		testValue := "test-value"

		// Cache the value
		err := cache.Cache(testKey, testValue, 1*time.Minute, 1)
		if err != nil {
			t.Errorf("Failed to cache value: %v", err)
		}

		// Retrieve the value
		value, exists := cache.Get(testKey)
		if !exists {
			t.Errorf("Expected key to exist, got false")
		}
		if value != testValue {
			t.Errorf("Expected value %v, got %v", testValue, value)
		}
	})

	// Test case 3: TTL expiry
	t.Run("TTL expiry", func(t *testing.T) {
		testKey := "expiry-key"
		testValue := "expiry-value"

		// Cache the value with a short TTL
		err := cache.Cache(testKey, testValue, 100*time.Millisecond, 1)
		if err != nil {
			t.Errorf("Failed to cache value: %v", err)
		}

		// Verify value exists immediately
		value, exists := cache.Get(testKey)
		if !exists {
			t.Errorf("Expected key to exist immediately after caching")
		}
		if value != testValue {
			t.Errorf("Expected value %v, got %v", testValue, value)
		}

		// Wait for TTL to expire
		time.Sleep(200 * time.Millisecond)

		// Verify value no longer exists
		value, exists = cache.Get(testKey)
		if exists {
			t.Errorf("Expected key to be expired, but it still exists")
		}
		if value != nil {
			t.Errorf("Expected expired key to return nil value, got %v", value)
		}
	})

	// Test case 4: Invalidate key
	t.Run("Invalidate key", func(t *testing.T) {
		testKey := "invalidate-key"
		testValue := "invalidate-value"

		// Cache the value
		err := cache.Cache(testKey, testValue, 1*time.Minute, 1)
		if err != nil {
			t.Errorf("Failed to cache value: %v", err)
		}

		// Verify value exists
		value, exists := cache.Get(testKey)
		if !exists {
			t.Errorf("Expected key to exist before invalidation")
		}
		if value != testValue {
			t.Errorf("Expected value %v, got %v", testValue, value)
		}

		// Invalidate the key
		err = cache.Invalidate(testKey)
		if err != nil {
			t.Errorf("Failed to invalidate key: %v", err)
		}

		// Verify value no longer exists
		value, exists = cache.Get(testKey)
		if exists {
			t.Errorf("Expected key to be invalidated, but it still exists")
		}
		if value != nil {
			t.Errorf("Expected invalidated key to return nil value, got %v", value)
		}
	})
}
