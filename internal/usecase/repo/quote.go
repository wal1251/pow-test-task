package repo

import "wisdom-server/internal/entity"

type QuoteStorage interface {
	Random() *entity.Quote
}
