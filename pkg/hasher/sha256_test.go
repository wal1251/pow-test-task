package hasher_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wisdom-server/internal/entity"
	"wisdom-server/pkg/hasher"
)

func TestSHA256Verifier_NewChallenge(t *testing.T) {
	v := hasher.NewSHA256Verifier()

	challenge, err := v.NewChallenge(20)
	require.NoError(t, err)

	assert.NotEmpty(t, challenge.Rand)
	assert.Len(t, challenge.Rand, 32)
	assert.Equal(t, uint8(20), challenge.Difficulty)
	assert.True(t, challenge.ExpiresAt.After(time.Now()))
}

func TestSHA256Verifier_Verify(t *testing.T) {
	v := hasher.NewSHA256Verifier()

	challenge := &entity.Challenge{
		Rand:       "test",
		Difficulty: 8,
		ExpiresAt:  time.Now().Add(time.Minute),
	}

	nonce, ok := hasher.Solve(challenge)
	require.True(t, ok)

	assert.True(t, v.Verify(challenge, nonce))
	assert.False(t, v.Verify(challenge, nonce+1))
}

func TestSHA256Verifier_VerifyExpired(t *testing.T) {
	v := hasher.NewSHA256Verifier()

	challenge := &entity.Challenge{
		Rand:       "test",
		Difficulty: 4,
		ExpiresAt:  time.Now().Add(-time.Second),
	}

	assert.False(t, v.Verify(challenge, 0))
}

func TestSolve(t *testing.T) {
	challenge := &entity.Challenge{
		Rand:       "test123",
		Difficulty: 12,
		ExpiresAt:  time.Now().Add(time.Minute),
	}

	nonce, ok := hasher.Solve(challenge)
	require.True(t, ok)

	v := hasher.NewSHA256Verifier()
	assert.True(t, v.Verify(challenge, nonce))
}
