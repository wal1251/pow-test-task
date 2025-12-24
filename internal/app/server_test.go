package app_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wisdom-server/config"
	"wisdom-server/internal/app"
	"wisdom-server/internal/controller/tcp"
	"wisdom-server/internal/repository"
	"wisdom-server/internal/usecase"
	"wisdom-server/pkg/hasher"
)

func TestServer_StartAndShutdown(t *testing.T) {
	testCfg := &config.ServerConfig{
		Addr:            "127.0.0.1:0", // Listen on a random available port
		Difficulty:      4,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		ShutdownTimeout: 1 * time.Second,
		RateLimit:       10,
		RateBurst:       5,
		CacheTTL:        30 * time.Second,
	}

	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	quoteStorage := repository.NewInMemoryQuoteStorage()
	challengeCache := repository.NewChallengeCache()
	verifier := hasher.NewSHA256Verifier()

	quoteService := usecase.NewQuoteService(quoteStorage)
	challengeService := usecase.NewChallengeService(verifier, challengeCache)

	handler := tcp.NewHandler(challengeService, quoteService, testCfg.ReadTimeout, testCfg.WriteTimeout, logger, testCfg.Difficulty)
	server := app.NewServer(testCfg.Addr, handler, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start(ctx)
	}()

	// Wait for the server to actually start listening
	time.Sleep(50 * time.Millisecond)

	// Attempt to connect to verify it's listening
	conn, err := net.Dial("tcp", server.Addr().String())
	require.NoError(t, err)
	conn.Close()

	cancel() // Trigger server shutdown

	select {
	case err := <-serverErr:
		assert.NoError(t, err)
	case <-time.After(testCfg.ShutdownTimeout + 500*time.Millisecond): // Give a bit more time for graceful shutdown
		t.Fatal("server did not shut down gracefully")
	}
}

func TestServer_InvalidAddress(t *testing.T) {
	testCfg := &config.ServerConfig{
		Addr:            "invalid:address:99999", // Invalid address
		Difficulty:      4,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		ShutdownTimeout: 1 * time.Second,
		RateLimit:       10,
		RateBurst:       5,
		CacheTTL:        30 * time.Second,
	}
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	quoteStorage := repository.NewInMemoryQuoteStorage()
	challengeCache := repository.NewChallengeCache()
	verifier := hasher.NewSHA256Verifier()

	quoteService := usecase.NewQuoteService(quoteStorage)
	challengeService := usecase.NewChallengeService(verifier, challengeCache)

	handler := tcp.NewHandler(challengeService, quoteService, testCfg.ReadTimeout, testCfg.WriteTimeout, logger, testCfg.Difficulty)
	server := app.NewServer(testCfg.Addr, handler, logger)

	ctx := context.Background()
	err := server.Start(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listen")
}