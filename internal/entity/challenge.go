package entity

import "time"

type Challenge struct {
	Rand       string
	Difficulty uint8
	ExpiresAt  time.Time
}

func (c *Challenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}
