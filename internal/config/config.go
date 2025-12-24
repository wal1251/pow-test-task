package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"wisdom-server/internal/entity"
)

// ========== SERVER CONFIG ==========

type ServerConfig struct {
	Addr            string
	Difficulty      uint8
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	RateLimit       int
	RateBurst       int
	CacheTTL        time.Duration
}

func LoadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{
		Addr:            getEnv("SERVER_ADDR", ":8080"),
		Difficulty:      uint8(getEnvInt("POW_DIFFICULTY", 20)),
		ReadTimeout:     getEnvDuration("READ_TIMEOUT_SEC", 60),
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT_SEC", 10),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT_SEC", 15),
		RateLimit:       getEnvInt("RATE_LIMIT", 10),
		RateBurst:       getEnvInt("RATE_BURST", 5),
		CacheTTL:        getEnvDuration("CACHE_TTL_SEC", 30),
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *ServerConfig) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("%w: server address is required", entity.ErrInvalidConfig)
	}
	if c.Difficulty == 0 || c.Difficulty > 32 {
		return fmt.Errorf("%w: difficulty must be between 1 and 32", entity.ErrInvalidConfig)
	}
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("%w: read timeout must be positive", entity.ErrInvalidConfig)
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("%w: write timeout must be positive", entity.ErrInvalidConfig)
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("%w: shutdown timeout must be positive", entity.ErrInvalidConfig)
	}
	if c.RateLimit <= 0 {
		return fmt.Errorf("%w: rate limit must be positive", entity.ErrInvalidConfig)
	}
	if c.RateBurst <= 0 {
		return fmt.Errorf("%w: rate burst must be positive", entity.ErrInvalidConfig)
	}
	if c.CacheTTL <= 0 {
		return fmt.Errorf("%w: cache TTL must be positive", entity.ErrInvalidConfig)
	}
	return nil
}

// ========== CLIENT CONFIG ==========

type ClientConfig struct {
	ServerAddr     string
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	SolveTimeout   time.Duration
}

func LoadClientConfig() (*ClientConfig, error) {
	cfg := &ClientConfig{
		ServerAddr:     getEnv("SERVER_ADDR", "localhost:8080"),
		ConnectTimeout: getEnvDuration("CONNECT_TIMEOUT_SEC", 10),
		ReadTimeout:    getEnvDuration("READ_TIMEOUT_SEC", 10),
		WriteTimeout:   getEnvDuration("WRITE_TIMEOUT_SEC", 10),
		SolveTimeout:   getEnvDuration("SOLVE_TIMEOUT_SEC", 60),
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *ClientConfig) Validate() error {
	if c.ServerAddr == "" {
		return fmt.Errorf("%w: server address is required", entity.ErrInvalidConfig)
	}
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("%w: connect timeout must be positive", entity.ErrInvalidConfig)
	}
	return nil
}

// ========== ENV HELPERS ==========

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultSeconds int) time.Duration {
	seconds := getEnvInt(key, defaultSeconds)
	return time.Duration(seconds) * time.Second
}
