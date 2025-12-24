package repository

import (
	"crypto/rand"
	"encoding/binary"

	"wisdom-server/internal/entity"
	"wisdom-server/internal/usecase/repo"
)

var defaultQuotes = []entity.Quote{
	{Text: "The only true wisdom is in knowing you know nothing.", Author: "Socrates"},
	{Text: "In the middle of difficulty lies opportunity.", Author: "Albert Einstein"},
	{Text: "Knowledge speaks, but wisdom listens.", Author: "Jimi Hendrix"},
	{Text: "The fool doth think he is wise, but the wise man knows himself to be a fool.", Author: "William Shakespeare"},
	{Text: "Wisdom is not a product of schooling but of the lifelong attempt to acquire it.", Author: "Albert Einstein"},
	{Text: "Turn your wounds into wisdom.", Author: "Oprah Winfrey"},
	{Text: "The only thing I know is that I know nothing.", Author: "Socrates"},
	{Text: "It is the mark of an educated mind to be able to entertain a thought without accepting it.", Author: "Aristotle"},
	{Text: "The unexamined life is not worth living.", Author: "Socrates"},
	{Text: "Knowing yourself is the beginning of all wisdom.", Author: "Aristotle"},
}

type inMemoryQuoteStorage struct {
	quotes []entity.Quote
}

func NewInMemoryQuoteStorage() repo.QuoteStorage {
	return &inMemoryQuoteStorage{quotes: defaultQuotes}
}

func (s *inMemoryQuoteStorage) Random() *entity.Quote {
	var b [8]byte
	_, _ = rand.Read(b[:])
	idx := binary.LittleEndian.Uint64(b[:]) % uint64(len(s.quotes))
	return &s.quotes[idx]
}
