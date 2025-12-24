package tcp

import (
	"bufio"
	"encoding/json"
	"io"
	"time"

	"wisdom-server/internal/entity"
)

type MessageType string

const (
	MessageTypeChallenge MessageType = "challenge"
	MessageTypeSolution  MessageType = "solution"
	MessageTypeQuote     MessageType = "quote"
	MessageTypeError     MessageType = "error"
)

type Message struct {
	Type       MessageType `json:"type"`
	Rand       string      `json:"rand,omitempty"`
	Difficulty uint8       `json:"difficulty,omitempty"`
	Nonce      uint64      `json:"nonce,omitempty"`
	ExpiresAt  time.Time   `json:"expires_at,omitempty"` // Add ExpiresAt field
	Text       string      `json:"text,omitempty"`
	Author     string      `json:"author,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func NewChallengeMessage(challenge *entity.Challenge) *Message {
	return &Message{
		Type:       MessageTypeChallenge,
		Rand:       challenge.Rand,
		Difficulty: challenge.Difficulty,
		ExpiresAt:  challenge.ExpiresAt, // Populate ExpiresAt
	}
}

func NewSolutionMessage(nonce uint64) *Message {
	return &Message{
		Type:  MessageTypeSolution,
		Nonce: nonce,
	}
}

func NewQuoteMessage(quote *entity.Quote) *Message {
	return &Message{
		Type:   MessageTypeQuote,
		Text:   quote.Text,
		Author: quote.Author,
	}
}

func NewErrorMessage(err error) *Message {
	return &Message{
		Type:  MessageTypeError,
		Error: err.Error(),
	}
}

func (m *Message) ToChallenge() *entity.Challenge {
	return &entity.Challenge{
		Rand:       m.Rand,
		Difficulty: m.Difficulty,
		ExpiresAt:  m.ExpiresAt, // Reconstruct ExpiresAt
	}
}

func (m *Message) ToQuote() *entity.Quote {
	return &entity.Quote{
		Text:   m.Text,
		Author: m.Author,
	}
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = e.w.Write(data)
	return err
}

const maxMessageSize = 4096

type Decoder struct {
	scanner *bufio.Scanner
}

func NewDecoder(r io.Reader) *Decoder {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, maxMessageSize), maxMessageSize)
	return &Decoder{scanner: scanner}
}

func (d *Decoder) Decode() (*Message, error) {
	if !d.scanner.Scan() {
		if err := d.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	var msg Message
	if err := json.Unmarshal(d.scanner.Bytes(), &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
