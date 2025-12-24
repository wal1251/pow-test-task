package entity

import "context"

// Verifier defines the interface for creating and verifying challenges.
type Verifier interface {
	// NewChallenge creates a new Proof-of-Work challenge.
	NewChallenge(difficulty uint8) (*Challenge, error)
	// Verify checks if the nonce is a valid solution for the challenge.
	Verify(challenge *Challenge, nonce uint64) bool
}

// ChallengeCache defines the interface for a cache that stores used challenge solutions
// to prevent replay attacks.
type ChallengeCache interface {
	// Add stores the given key in the cache for a configured TTL.
	Add(ctx context.Context, key string) error
	// Exists checks if the given key is present in the cache.
	Exists(ctx context.Context, key string) (bool, error)
}

// RateLimiter defines the interface for a generic rate limiter.
type RateLimiter interface {
	// Allow checks if a request from the given identifier (e.g., IP address)
	// is allowed under the current rate-limiting policy.
	Allow(ctx context.Context, identifier string) (bool, error)
}