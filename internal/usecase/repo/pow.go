package repo

import "wisdom-server/internal/entity"

// Verifier defines the interface for creating and verifying challenges.
type Verifier interface {
	// NewChallenge creates a new Proof-of-Work challenge.
	NewChallenge(difficulty uint8) (*entity.Challenge, error)
	// Verify checks if the nonce is a valid solution for the challenge.
	Verify(challenge *entity.Challenge, nonce uint64) bool
}
