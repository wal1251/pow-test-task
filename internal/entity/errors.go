package entity

import "errors"

var (
	ErrChallengeExpired     = errors.New("challenge expired")
	ErrInvalidSolution      = errors.New("invalid solution")
	ErrProtocolViolation    = errors.New("protocol violation")
	ErrInvalidConfig        = errors.New("invalid config")
	ErrUnexpectedMessage    = errors.New("unexpected message type")
	ErrSolveTimeout         = errors.New("failed to solve challenge")
	ErrChallengeReused      = errors.New("challenge already used")
)
