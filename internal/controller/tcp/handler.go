package tcp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"

	"wisdom-server/internal/entity"
	"wisdom-server/internal/usecase"
)

type AppService interface {
	CreateChallenge(difficulty uint8) (*entity.Challenge, error)
	VerifySolution(ctx context.Context, challenge *entity.Challenge, nonce uint64) error
	GetRandomQuote() *entity.Quote
}

type Handler struct {
	challengeUseCase *usecase.ChallengeService
	quoteUseCase     *usecase.QuoteService
	readTimeout      time.Duration
	writeTimeout     time.Duration
	logger           zerolog.Logger
	difficulty       uint8
}

func NewHandler(
	challengeUseCase *usecase.ChallengeService,
	quoteUseCase *usecase.QuoteService,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	logger zerolog.Logger,
	difficulty uint8,
) *Handler {
	return &Handler{
		challengeUseCase: challengeUseCase,
		quoteUseCase:     quoteUseCase,
		readTimeout:      readTimeout,
		writeTimeout:     writeTimeout,
		logger:           logger,
		difficulty:       difficulty,
	}
}

func (h *Handler) Handle(conn net.Conn) {
	defer conn.Close()

	ctx := context.Background()
	remoteAddr := conn.RemoteAddr().String()
	log := h.logger.With().Str("addr", remoteAddr).Logger()

	log.Info().Msg("new connection")

	enc := NewEncoder(conn)
	dec := NewDecoder(conn)

	challenge, err := h.challengeUseCase.CreateChallenge(h.difficulty)
	if err != nil {
		log.Error().Err(err).Msg("failed to create challenge")
		return
	}

	if err := conn.SetWriteDeadline(time.Now().Add(h.writeTimeout)); err != nil {
		log.Error().Err(err).Msg("failed to set write deadline")
		return
	}

	if err := enc.Encode(NewChallengeMessage(challenge)); err != nil {
		log.Error().Err(err).Msg("failed to send challenge")
		return
	}

	if err := conn.SetReadDeadline(time.Now().Add(h.readTimeout)); err != nil {
		log.Error().Err(err).Msg("failed to set read deadline")
		return
	}

	msg, err := dec.Decode()
	if err != nil {
		log.Error().Err(err).Msg("failed to read solution")
		return
	}

	if msg.Type != MessageTypeSolution {
		log.Warn().Str("type", string(msg.Type)).Msg("unexpected message type")
		if err := conn.SetWriteDeadline(time.Now().Add(h.writeTimeout)); err != nil {
			log.Error().Err(err).Msg("failed to set write deadline")
			return
		}
		protocolErr := fmt.Errorf("%w: expected %s, got %s", entity.ErrProtocolViolation, MessageTypeSolution, msg.Type)
		if err := enc.Encode(NewErrorMessage(protocolErr)); err != nil {
			log.Error().Err(err).Msg("failed to send error")
		}
		return
	}

	if err := h.challengeUseCase.VerifySolution(ctx, challenge, msg.Nonce); err != nil {
		log.Warn().Err(err).Msg("verification failed")
		if err := conn.SetWriteDeadline(time.Now().Add(h.writeTimeout)); err != nil {
			log.Error().Err(err).Msg("failed to set write deadline")
			return
		}
		if err := enc.Encode(NewErrorMessage(err)); err != nil {
			log.Error().Err(err).Msg("failed to send error")
		}
		return
	}

	quote := h.quoteUseCase.GetRandomQuote()

	if err := conn.SetWriteDeadline(time.Now().Add(h.writeTimeout)); err != nil {
		log.Error().Err(err).Msg("failed to set write deadline")
		return
	}

	if err := enc.Encode(NewQuoteMessage(quote)); err != nil {
		log.Error().Err(err).Msg("failed to send quote")
		return
	}

	log.Info().Str("author", quote.Author).Msg("quote sent")
}
