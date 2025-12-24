package tcp_test

import (
	"net"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wisdom-server/internal/controller/tcp"
	"wisdom-server/internal/entity"
	"wisdom-server/internal/repository"
	"wisdom-server/internal/usecase"
	"wisdom-server/pkg/hasher"
)

func setupTestHandler(difficulty uint8) (*tcp.Handler, net.Listener) {
	logger := zerolog.Nop()

	quoteStorage := repository.NewInMemoryQuoteStorage()
	challengeCache := repository.NewChallengeCache()
	verifier := hasher.NewSHA256Verifier()

	quoteService := usecase.NewQuoteService(quoteStorage)
	challengeService := usecase.NewChallengeService(verifier, challengeCache)

	handler := tcp.NewHandler(challengeService, quoteService, 5*time.Second, 5*time.Second, logger, difficulty)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	return handler, listener
}

func TestHandler_SuccessfulFlow(t *testing.T) {
	handler, listener := setupTestHandler(4)
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		handler.Handle(conn)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	enc := tcp.NewEncoder(conn)
	dec := tcp.NewDecoder(conn)

	// 1. Receive challenge
	msg, err := dec.Decode()
	require.NoError(t, err)
	assert.Equal(t, tcp.MessageTypeChallenge, msg.Type)

	challenge := msg.ToChallenge()
	challenge.ExpiresAt = time.Now().Add(time.Minute) // Ensure challenge is not expired for Solve()
	nonce, ok := hasher.Solve(challenge)
	require.True(t, ok)

	// 2. Send solution
	require.NoError(t, enc.Encode(tcp.NewSolutionMessage(nonce)))

	// 3. Receive quote
	msg, err = dec.Decode()
	require.NoError(t, err)
	assert.Equal(t, tcp.MessageTypeQuote, msg.Type)
	assert.NotEmpty(t, msg.Text)
	assert.NotEmpty(t, msg.Author)
}

func TestHandler_InvalidSolution(t *testing.T) {
	handler, listener := setupTestHandler(8)
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		handler.Handle(conn)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	enc := tcp.NewEncoder(conn)
	dec := tcp.NewDecoder(conn)

	// 1. Receive challenge
	_, err = dec.Decode()
	require.NoError(t, err)

	// 2. Send invalid solution
	require.NoError(t, enc.Encode(tcp.NewSolutionMessage(999999)))

	// 3. Receive error
	msg, err := dec.Decode()
	require.NoError(t, err)
	assert.Equal(t, tcp.MessageTypeError, msg.Type)
	assert.Contains(t, msg.Error, entity.ErrInvalidSolution.Error())
}

func TestHandler_ProtocolViolation(t *testing.T) {
	handler, listener := setupTestHandler(4)
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		handler.Handle(conn)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	enc := tcp.NewEncoder(conn)
	dec := tcp.NewDecoder(conn)

	// 1. Receive challenge
	_, err = dec.Decode()
	require.NoError(t, err)

	// 2. Send a quote message instead of solution
	require.NoError(t, enc.Encode(tcp.NewQuoteMessage(&entity.Quote{Text: "wrong", Author: "type"})))

	// 3. Receive error
	msg, err := dec.Decode()
	require.NoError(t, err)
	assert.Equal(t, tcp.MessageTypeError, msg.Type)
	assert.Contains(t, msg.Error, entity.ErrProtocolViolation.Error())
}
