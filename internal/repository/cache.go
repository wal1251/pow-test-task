package repository

import (

	"context"

	"sync"

	"time"



	"wisdom-server/internal/usecase/repo"

)



// MemoryCache is an in-memory, thread-safe cache with a TTL for storing challenge nonces.

type MemoryCache struct {

	mu      sync.RWMutex

	storage map[string]time.Time

	ttl     time.Duration

	stop    chan struct{}

}



// NewMemoryCache creates a new in-memory cache and starts a background cleanup goroutine.

func NewMemoryCache(ttl, cleanupInterval time.Duration) *MemoryCache {

	c := &MemoryCache{

		storage: make(map[string]time.Time),

		ttl:     ttl,

		stop:    make(chan struct{}),

	}

	go c.runCleanup(cleanupInterval)

	return c

}



// Add stores a key in the cache, marking it with an expiration time.

func (c *MemoryCache) Add(ctx context.Context, key string) error {

	c.mu.Lock()

	defer c.mu.Unlock()

	c.storage[key] = time.Now().Add(c.ttl)

	return nil

}



// Exists checks if a key exists and is not expired.

func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {

	c.mu.RLock()

	defer c.mu.RUnlock()



	expireTime, exists := c.storage[key]

	if !exists {

		return false, nil

	}



	if time.Now().After(expireTime) {

		// The cleanup goroutine will handle deletion, but for immediate consistency,

		// we can treat it as non-existent.

		return false, nil

	}



	return true, nil

}



// Stop terminates the background cleanup goroutine.

func (c *MemoryCache) Stop() {

	close(c.stop)

}



// runCleanup periodically scans the cache and removes expired items.

func (c *MemoryCache) runCleanup(interval time.Duration) {

	ticker := time.NewTicker(interval)

	defer ticker.Stop()



	for {

		select {

		case <-ticker.C:

			c.cleanupExpired()

		case <-c.stop:

			return

		}

	}

}



// cleanupExpired is the internal function that iterates and deletes expired keys.

func (c *MemoryCache) cleanupExpired() {

	c.mu.Lock()

	defer c.mu.Unlock()



	now := time.Now()

	for key, expireTime := range c.storage {

		if now.After(expireTime) {

			delete(c.storage, key)

		}

	}

}



// Compile-time check to ensure MemoryCache implements the ChallengeCache interface.

var _ repo.ChallengeCache = (*MemoryCache)(nil)
