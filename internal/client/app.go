package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"wisdom-server/config"
	"wisdom-server/internal/controller/tcp"
	"wisdom-server/pkg/hasher"
	"github.com/rs/zerolog"
)

type App struct {
	cfg    *config.ClientConfig
	logger zerolog.Logger
}

func NewApp(cfg *config.ClientConfig) *App {
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()

	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	conn, err := net.DialTimeout("tcp", a.cfg.ServerAddr, a.cfg.ConnectTimeout)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	defer conn.Close()

	a.logger.Info().Str("addr", a.cfg.ServerAddr).Msg("connected to server")

	enc := tcp.NewEncoder(conn)
	dec := tcp.NewDecoder(conn)

	// 1. Receive challenge
	if err := conn.SetReadDeadline(time.Now().Add(a.cfg.ReadTimeout)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}
	msg, err := dec.Decode()
	if err != nil {
		return fmt.Errorf("receive challenge: %w", err)
	}
	if msg.Type != tcp.MessageTypeChallenge {
		return fmt.Errorf("unexpected message type: %s", msg.Type)
	}

	challenge := msg.ToChallenge()
	a.logger.Info().Msg("challenge received")

	// 2. Solve challenge
	start := time.Now()
	nonce, ok := hasher.Solve(challenge)
	if !ok {
		return fmt.Errorf("failed to solve challenge")
	}
	elapsed := time.Since(start)
	a.logger.Info().Uint64("nonce", nonce).Dur("elapsed", elapsed).Msg("challenge solved")

	// 3. Send solution
	if err := conn.SetWriteDeadline(time.Now().Add(a.cfg.WriteTimeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	if err := enc.Encode(tcp.NewSolutionMessage(nonce)); err != nil {
		return fmt.Errorf("send solution: %w", err)
	}

	// 4. Receive quote
	if err := conn.SetReadDeadline(time.Now().Add(a.cfg.ReadTimeout)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}
	msg, err = dec.Decode()
	if err != nil {
		return fmt.Errorf("receive quote: %w", err)
	}

	if msg.Type == tcp.MessageTypeError {
		return fmt.Errorf("server error: %s", msg.Error)
	}

	if msg.Type != tcp.MessageTypeQuote {
		return fmt.Errorf("unexpected message type: %s", msg.Type)
	}

	quote := msg.ToQuote()
	a.logger.Info().Str("author", quote.Author).Msgf("quote: %q", quote.Text)

	return nil
}
