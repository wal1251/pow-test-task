package tests

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wisdom-server/internal/config"
	"wisdom-server/internal/app"
	"wisdom-server/internal/client"
	"wisdom-server/internal/controller/tcp"
	"wisdom-server/internal/entity"
)

func setupTestServer(t *testing.T) (serverApp *app.App, serverAddr string, cancelServer context.CancelFunc) {
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

	logger := zerolog.Nop()
	serverCtx, cancelServer := context.WithCancel(context.Background())

	serverApp, err := app.NewApp(serverCtx, testCfg)
	require.NoError(t, err)

	go func() {
		err := serverApp.Run(serverCtx)
		if err != nil && err != net.ErrClosed {
			logger.Error().Err(err).Msg("server exited with error")
		}
	}()

	// Wait for the server to actually start listening
	time.Sleep(50 * time.Millisecond)

	serverAddr = serverApp.Server.Addr().String()

	return serverApp, serverAddr, cancelServer
}

func TestIntegration_FullFlow(t *testing.T) {
	_, serverAddr, cancelServer := setupTestServer(t)
	defer cancelServer()

	clientCfg := &config.ClientConfig{
		ServerAddr:     serverAddr,
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		SolveTimeout:   60 * time.Second,
	}

	clientApp := client.NewApp(clientCfg)
	err := clientApp.Run(context.Background())
	require.NoError(t, err)
}

func TestIntegration_InvalidSolution(t *testing.T) {
	_, serverAddr, cancelServer := setupTestServer(t)
	defer cancelServer()

	// This test needs to modify the client logic to send an invalid nonce.
	// Since client.App.Run encapsulates the logic, we need to either:
	// 1. Create a specific test client implementation
	// 2. Modify client.App.Run to accept a "nonce override" for testing invalid solutions.
	// For now, we will simulate the invalid solution by directly interacting
	// with the TCP connection as a client, similar to how TestHandler does.

	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	require.NoError(t, err)
	defer conn.Close()

	enc := tcp.NewEncoder(conn)
	dec := tcp.NewDecoder(conn)

	// 1. Receive challenge
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	msg, err := dec.Decode()
	require.NoError(t, err)
	require.Equal(t, tcp.MessageTypeChallenge, msg.Type)

	// 2. Send invalid solution
	require.NoError(t, conn.SetWriteDeadline(time.Now().Add(5*time.Second)))
	require.NoError(t, enc.Encode(tcp.NewSolutionMessage(999999)))

	// 3. Receive error
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	msg, err = dec.Decode()
	require.NoError(t, err)
	require.Equal(t, tcp.MessageTypeError, msg.Type)
	assert.Contains(t, msg.Error, entity.ErrInvalidSolution.Error())
}

func TestIntegration_MultipleClients(t *testing.T) {
	_, serverAddr, cancelServer := setupTestServer(t)
	defer cancelServer()

	clientCount := 5
	var wg sync.WaitGroup
	errs := make(chan error, clientCount)

	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clientCfg := &config.ClientConfig{
				ServerAddr:     serverAddr,
				ConnectTimeout: 5 * time.Second,
				ReadTimeout:    5 * time.Second,
				WriteTimeout:   5 * time.Second,
				SolveTimeout:   60 * time.Second,
			}
			clientApp := client.NewApp(clientCfg)
			err := clientApp.Run(context.Background())
			if err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		assert.NoError(t, err)
	}
}