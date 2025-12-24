package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"wisdom-server/config"
	"wisdom-server/internal/client"
	"github.com/rs/zerolog"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	cfg, err := config.LoadClientConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	app := client.NewApp(cfg)

	if err := app.Run(ctx); err != nil {
		logger.Fatal().Err(err).Msg("client error")
	}
}
