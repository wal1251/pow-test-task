package hasher

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"wisdom-server/internal/entity"
	"wisdom-server/internal/usecase/repo"
)

const (
	randBytes         = 16
	challengeLifetime = 30 * time.Second
)

type sha256Verifier struct{}

func NewSHA256Verifier() repo.Verifier {
	return &sha256Verifier{}
}

func (v *sha256Verifier) NewChallenge(difficulty uint8) (*entity.Challenge, error) {
	randData := make([]byte, randBytes)
	if _, err := io.ReadFull(rand.Reader, randData); err != nil {
		return nil, fmt.Errorf("generate random: %w", err)
	}

	return &entity.Challenge{
		Rand:       hex.EncodeToString(randData),
		Difficulty: difficulty,
		ExpiresAt:  time.Now().Add(challengeLifetime),
	}, nil
}

func (v *sha256Verifier) Verify(challenge *entity.Challenge, nonce uint64) bool {
	if challenge.IsExpired() {
		return false
	}

	data := fmt.Sprintf("%s:%d:%d", challenge.Rand, challenge.Difficulty, nonce)
	hash := sha256.Sum256([]byte(data))

	return hasLeadingZeroBits(hash[:], challenge.Difficulty)
}

func hasLeadingZeroBits(hash []byte, bits uint8) bool {
	fullBytes := bits / 8
	remainingBits := bits % 8

	for i := uint8(0); i < fullBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}

	if remainingBits > 0 {
		mask := byte(0xFF << (8 - remainingBits))
		if hash[fullBytes]&mask != 0 {
			return false
		}
	}

	return true
}

const expiryCheckInterval = 10000

func Solve(challenge *entity.Challenge) (uint64, bool) {
	data := fmt.Sprintf("%s:%d:", challenge.Rand, challenge.Difficulty)

	for nonce := uint64(0); ; nonce++ {
		if nonce%expiryCheckInterval == 0 && challenge.IsExpired() {
			return 0, false
		}

		fullData := fmt.Sprintf("%s%d", data, nonce)
		hash := sha256.Sum256([]byte(fullData))

		if hasLeadingZeroBits(hash[:], challenge.Difficulty) {
			return nonce, true
		}
	}
}
