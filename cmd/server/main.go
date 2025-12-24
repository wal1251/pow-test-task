package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"wisdom-server/internal/app"
	"wisdom-server/internal/config"
	"github.com/rs/zerolog"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	cfg, err := config.LoadServerConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize app")
	}

	if err := application.Run(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}
}
