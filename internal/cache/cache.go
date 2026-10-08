package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
	fetchedAt time.Time
}

type TTLCache struct {
	mu sync.RWMutex
	m  map[string]entry
}

func New() *TTLCache {
	return &TTLCache{
		m: make(map[string]entry),
	}
}

func (c *TTLCache) Get(key string) (value any, fetchedAt time.Time, ok bool) {
	c.mu.RLock()
	item, exists := c.m[key]

	if !exists {
		c.mu.RUnlock()
		return nil, time.Time{}, false
	}

	if time.Now().After(item.expiresAt) {
		c.mu.RUnlock()

		c.mu.Lock()
		item, exists = c.m[key]
		if exists && time.Now().After(item.expiresAt) {
			delete(c.m, key)
		}
		c.mu.Unlock()
		return nil, time.Time{}, false
	}

	fetchedAt = item.fetchedAt
	value = item.value
	ok = true
	c.mu.RUnlock()
	return value, fetchedAt, ok
}

func (c *TTLCache) Set(key string, value any, ttl time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	c.m[key] = entry{
		value:     value,
		expiresAt: now.Add(ttl),
		fetchedAt: now,
	}

	return now
}
