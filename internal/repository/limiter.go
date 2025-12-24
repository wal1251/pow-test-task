package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"wisdom-server/internal/usecase/repo"
)

// Limiter is a rate limiter implementation based on the ulule/limiter library.
type Limiter struct {
	instance *limiter.Limiter
}

// NewLimiter creates a new rate limiter with the specified rate.
func NewLimiter(requests int, period time.Duration) (*Limiter, error) {
	rate := limiter.Rate{
		Period: period,
		Limit:  int64(requests),
	}

	store := memory.NewStore()

	instance := limiter.New(store, rate)
	if instance == nil {
		return nil, fmt.Errorf("could not create limiter instance")
	}

	return &Limiter{
		instance: instance,
	}, nil
}

// Allow checks if a request for the given key is permitted by the rate limiter.
func (l *Limiter) Allow(ctx context.Context, key string) (bool, error) {
	res, err := l.instance.Get(ctx, key)
	if err != nil {
		return false, fmt.Errorf("limiter check failed: %w", err)
	}

	return !res.Reached, nil
}

// Compile-time check to ensure Limiter implements the RateLimiter interface.
var _ repo.RateLimiter = (*Limiter)(nil)