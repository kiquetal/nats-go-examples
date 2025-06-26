// Package cache provides caching functionality for tokens
package cache

import (
	"sync"
	"time"
)

// TokenCache provides a thread-safe cache for storing tokens with expiration
type TokenCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	token      string
	expiration time.Time
}

// NewTokenCache creates a new TokenCache
func NewTokenCache() *TokenCache {
	// Initialize a new cache
	cache := &TokenCache{
		items: make(map[string]*cacheItem),
	}

	// Start a goroutine to clean expired items periodically
	go cache.cleanExpired()

	return cache
}

// cleanExpired removes expired items from the cache every minute
func (c *TokenCache) cleanExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.removeExpired()
	}
}

// removeExpired removes all expired items from the cache
func (c *TokenCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if item.expiration.Before(now) {
			delete(c.items, key)
		}
	}
}

// Set adds or updates a token in the cache with a specified TTL
func (c *TokenCache) Set(clientID string, token string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[clientID] = &cacheItem{
		token:      token,
		expiration: time.Now().Add(ttl),
	}
}

// Get retrieves a token from the cache if it exists and is not expired
func (c *TokenCache) Get(clientID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[clientID]
	if !exists {
		return "", false
	}

	// Check if the item has expired
	if time.Now().After(item.expiration) {
		return "", false
	}

	return item.token, true
}

// Delete removes a token from the cache
func (c *TokenCache) Delete(clientID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, clientID)
}

// Clear removes all items from the cache
func (c *TokenCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
}
