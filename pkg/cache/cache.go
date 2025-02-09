package cache

import (
	"sync"
	"time"
)

type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

type InMemoryCache struct {
	data map[string]CacheItem
	mu   sync.RWMutex
}

func NewCache() *InMemoryCache {
	return &InMemoryCache{
		data: make(map[string]CacheItem),
	}
}

// set adds data to cache with TTL(time to live)
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// get cached data if it exist and not expired
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists || time.Now().After(item.ExpiresAt) {
		return nil, false
	}
	return item.Data, true
}
