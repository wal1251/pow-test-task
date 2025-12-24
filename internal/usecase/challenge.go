package usecase

import (
	"context"

	"wisdom-server/internal/entity"
	"wisdom-server/internal/usecase/repo"
)

// ChallengeService coordinates challenge creation and verification
type ChallengeService struct {
	verifier repo.Verifier
	cache    repo.ChallengeCache
}

// NewChallengeService creates a new service
func NewChallengeService(verifier repo.Verifier, cache repo.ChallengeCache) *ChallengeService {
	return &ChallengeService{
		verifier: verifier,
		cache:    cache,
	}
}

// CreateChallenge creates a new challenge
func (s *ChallengeService) CreateChallenge(difficulty uint8) (*entity.Challenge, error) {
	return s.verifier.NewChallenge(difficulty)
}

// VerifySolution verifies PoW solution with replay attack protection
func (s *ChallengeService) VerifySolution(ctx context.Context, challenge *entity.Challenge, nonce uint64) error {
	// 1. Check expiration
	if challenge.IsExpired() {
		return entity.ErrChallengeExpired
	}

	// 2. Check reuse (replay attack protection)
	used, err := s.cache.Exists(ctx, challenge.Rand)
	if err != nil {
		return err
	}
	if used {
		return entity.ErrChallengeReused
	}

	// 3. Verify PoW
	if !s.verifier.Verify(challenge, nonce) {
		return entity.ErrInvalidSolution
	}

	// 4. Mark as used
	if err := s.cache.Add(ctx, challenge.Rand); err != nil {
		return err
	}

	return nil
}
