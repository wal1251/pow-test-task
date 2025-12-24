package repo

import "context"

// ChallengeCache defines the interface for a cache that stores used challenge solutions
// to prevent replay attacks.
type ChallengeCache interface {
	// Add stores the given key in the cache for a configured TTL.
	Add(ctx context.Context, key string) error
	// Exists checks if the given key is present in the cache.
	Exists(ctx context.Context, key string) (bool, error)
}