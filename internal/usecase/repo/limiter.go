package repo

import "context"

// RateLimiter defines the interface for a generic rate limiter.
type RateLimiter interface {
	// Allow checks if a request from the given identifier (e.g., IP address)
	// is allowed under the current rate-limiting policy.
	Allow(ctx context.Context, identifier string) (bool, error)
}
