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

var WorkoutCache = NewCache()

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

// delete item from cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// cleanup expired cache item(by time)
func (c *InMemoryCache) Cleanup(interval time.Duration, stopChan chan struct{}) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.mu.Lock()
				for key, item := range c.data {
					if time.Now().After(item.ExpiresAt) {
						delete(c.data, key)
					}
				}
				c.mu.Unlock()
			case <-stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}
