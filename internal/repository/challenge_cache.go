package repository

import (
	"context"
	"sync"
	"time"

	"wisdom-server/internal/usecase/repo"
)

const (
	// TTL should match Challenge expiration time (30s)
	challengeTTL = 30 * time.Second

	// Frequency of cleanup for expired entries
	cleanupInterval = 10 * time.Second
)

// ChallengeCache stores used challenges for replay attack protection
type ChallengeCache struct {
	mu   sync.RWMutex
	data map[string]time.Time // challengeID -> expiry time
	done chan struct{}
}

// NewChallengeCache creates a new in-memory cache with auto-cleanup
func NewChallengeCache() repo.ChallengeCache {
	c := &ChallengeCache{
		data: make(map[string]time.Time),
		done: make(chan struct{}),
	}

	// Start background cleanup
	go c.cleanupLoop()

	return c
}

// Add marks a challenge as used
func (c *ChallengeCache) Add(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = time.Now().Add(challengeTTL)
	return nil
}

// Exists checks if a challenge has already been used
func (c *ChallengeCache) Exists(_ context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expiry, exists := c.data[key]
	if !exists {
		return false, nil
	}

	// Check if TTL has expired
	return time.Now().Before(expiry), nil
}

// Close stops the background cleanup
func (c *ChallengeCache) Close() {
	close(c.done)
}

// cleanupLoop periodically removes expired entries
func (c *ChallengeCache) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.done:
			return
		}
	}
}

func (c *ChallengeCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, expiry := range c.data {
		if now.After(expiry) {
			delete(c.data, id)
		}
	}
}
