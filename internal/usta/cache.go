package usta

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// cacheEntry holds a cached value with expiration
type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// cache provides simple in-memory caching with TTL
type cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

// newCache creates a new cache with the given TTL
func newCache(ttl time.Duration) *cache {
	return &cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

// get retrieves a value from cache if it exists and hasn't expired
func (c *cache) get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.value, true
}

// set stores a value in cache with TTL
func (c *cache) set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

// Global caches with 10-minute TTL
var (
	teamCache = newCache(10 * time.Minute)
	orgCache  = newCache(10 * time.Minute)
)

// Global singleflight groups for request deduplication
var (
	teamGroup singleflight.Group
	orgGroup  singleflight.Group
)
