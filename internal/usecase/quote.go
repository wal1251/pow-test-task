package usecase

import (
	"wisdom-server/internal/entity"
	"wisdom-server/internal/usecase/repo"
)

// QuoteService manages quotes
type QuoteService struct {
	storage repo.QuoteStorage
}

// NewQuoteService creates a new service
func NewQuoteService(storage repo.QuoteStorage) *QuoteService {
	return &QuoteService{storage: storage}
}

// GetRandomQuote returns a random quote
func (s *QuoteService) GetRandomQuote() *entity.Quote {
	return s.storage.Random()
}
