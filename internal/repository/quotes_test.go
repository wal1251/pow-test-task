package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryQuoteStorage_Random(t *testing.T) {
	storage := NewInMemoryQuoteStorage()

	quote := storage.Random()

	require.NotNil(t, quote)
	assert.NotEmpty(t, quote.Text)
	assert.NotEmpty(t, quote.Author)
}

func TestInMemoryQuoteStorage_RandomDistribution(t *testing.T) {
	storage := NewInMemoryQuoteStorage()
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		quote := storage.Random()
		seen[quote.Text] = true
	}

	assert.Greater(t, len(seen), 1, "should return different quotes")
}
