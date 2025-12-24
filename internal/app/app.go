package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"wisdom-server/internal/config"
	"wisdom-server/internal/controller/tcp"
	"wisdom-server/internal/repository"
	"wisdom-server/internal/usecase"
	"wisdom-server/pkg/hasher"
	"github.com/rs/zerolog"
)

type App struct {
	Server *Server
	logger zerolog.Logger
}

func NewApp(ctx context.Context, cfg *config.ServerConfig) (*App, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()

	quoteStorage := repository.NewInMemoryQuoteStorage()
	challengeCache := repository.NewChallengeCache()
	verifier := hasher.NewSHA256Verifier()

	quoteService := usecase.NewQuoteService(quoteStorage)
	challengeService := usecase.NewChallengeService(verifier, challengeCache)

	handler := tcp.NewHandler(challengeService, quoteService, cfg.ReadTimeout, cfg.WriteTimeout, logger, cfg.Difficulty)
	server := NewServer(cfg.Addr, handler, logger)

	return &App{
		Server: server,
		logger: logger,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	return a.Server.Start(ctx)
}

type Server struct {
	addr     string
	handler  *tcp.Handler
	listener net.Listener
	wg       sync.WaitGroup
	logger   zerolog.Logger
}

func NewServer(addr string, handler *tcp.Handler, logger zerolog.Logger) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
		logger:  logger,
	}
}

func (s *Server) Addr() net.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

func (s *Server) Start(ctx context.Context) error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.logger.Info().Str("addr", s.addr).Msg("server started")

	go func() {
		<-ctx.Done()
		if err := s.listener.Close(); err != nil {
			s.logger.Error().Err(err).Msg("failed to close listener")
		}
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				s.wg.Wait()
				return nil
			default:
				s.logger.Error().Err(err).Msg("accept failed")
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handler.Handle(conn)
		}()
	}
}
