package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChallenge_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "not expired - future time",
			expiresAt: time.Now().Add(time.Minute),
			expected:  false,
		},
		{
			name:      "expired - past time",
			expiresAt: time.Now().Add(-time.Second),
			expected:  true,
		},
		{
			name:      "expired - zero time",
			expiresAt: time.Time{},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Challenge{
				Rand:       "test",
				Difficulty: 10,
				ExpiresAt:  tt.expiresAt,
			}
			assert.Equal(t, tt.expected, c.IsExpired())
		})
	}
}
