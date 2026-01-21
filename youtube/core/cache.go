package core

import (
	"sync"
	"time"
)

// Cache provides in-memory caching with TTL (time-to-live) support.
// It is safe for concurrent use.
type Cache struct {
	mu       sync.RWMutex
	items    map[string]*cacheItem
	defaultTTL time.Duration
	maxItems int
	now      func() time.Time // For testing
}

// cacheItem represents a cached value with expiration.
type cacheItem struct {
	value     any
	expiresAt time.Time
}

// CacheOption configures a Cache.
type CacheOption func(*Cache)

// NewCache creates a new cache with the given options.
func NewCache(opts ...CacheOption) *Cache {
	c := &Cache{
		items:      make(map[string]*cacheItem),
		defaultTTL: 5 * time.Minute,
		maxItems:   1000,
		now:        time.Now,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithDefaultTTL sets the default TTL for cache entries.
func WithDefaultTTL(ttl time.Duration) CacheOption {
	return func(c *Cache) { c.defaultTTL = ttl }
}

// WithMaxItems sets the maximum number of items in the cache.
// When exceeded, oldest items are evicted.
func WithMaxItems(max int) CacheOption {
	return func(c *Cache) { c.maxItems = max }
}

// WithTimeFunc sets a custom time function (for testing).
func WithTimeFunc(fn func() time.Time) CacheOption {
	return func(c *Cache) { c.now = fn }
}

// Get retrieves a value from the cache.
// Returns the value and true if found and not expired, nil and false otherwise.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if c.now().After(item.expiresAt) {
		// Item expired, clean it up
		c.Delete(key)
		return nil, false
	}

	return item.value, true
}

// Set stores a value in the cache with the default TTL.
func (c *Cache) Set(key string, value any) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL.
func (c *Cache) SetWithTTL(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict items
	if len(c.items) >= c.maxItems {
		c.evictExpiredLocked()
		// If still at max, evict oldest
		if len(c.items) >= c.maxItems {
			c.evictOldestLocked()
		}
	}

	c.items[key] = &cacheItem{
		value:     value,
		expiresAt: c.now().Add(ttl),
	}
}

// Delete removes a value from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all values from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*cacheItem)
}

// Len returns the number of items in the cache (including expired).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Keys returns all keys in the cache (including expired).
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}

// Cleanup removes all expired items from the cache.
func (c *Cache) Cleanup() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.evictExpiredLocked()
}

// evictExpiredLocked removes expired items. Must be called with lock held.
func (c *Cache) evictExpiredLocked() int {
	now := c.now()
	count := 0
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
			count++
		}
	}
	return count
}

// evictOldestLocked removes the oldest item. Must be called with lock held.
func (c *Cache) evictOldestLocked() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expiresAt
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// CacheStats contains cache statistics.
type CacheStats struct {
	// Items is the total number of items in the cache.
	Items int

	// Expired is the number of expired items (not yet cleaned up).
	Expired int

	// Active is the number of non-expired items.
	Active int
}

// Stats returns cache statistics.
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := c.now()
	expired := 0
	for _, item := range c.items {
		if now.After(item.expiresAt) {
			expired++
		}
	}

	return CacheStats{
		Items:   len(c.items),
		Expired: expired,
		Active:  len(c.items) - expired,
	}
}

// GetOrSet retrieves a value from the cache, or sets it using the provided
// function if not found or expired. This is atomic - only one goroutine
// will call the function even if multiple goroutines call GetOrSet concurrently.
func (c *Cache) GetOrSet(key string, fn func() (any, error)) (any, error) {
	return c.GetOrSetWithTTL(key, c.defaultTTL, fn)
}

// GetOrSetWithTTL is like GetOrSet but with a custom TTL.
func (c *Cache) GetOrSetWithTTL(key string, ttl time.Duration, fn func() (any, error)) (any, error) {
	// Try to get existing value first
	if value, ok := c.Get(key); ok {
		return value, nil
	}

	// Need to compute the value
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if item, ok := c.items[key]; ok && c.now().Before(item.expiresAt) {
		return item.value, nil
	}

	// Compute the value
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Check if we need to evict items
	if len(c.items) >= c.maxItems {
		c.evictExpiredLocked()
		if len(c.items) >= c.maxItems {
			c.evictOldestLocked()
		}
	}

	c.items[key] = &cacheItem{
		value:     value,
		expiresAt: c.now().Add(ttl),
	}

	return value, nil
}
