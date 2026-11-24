package notify

import (
	"context"
	"log"
	"sync"
	"time"
)

// Wiadomość wysyłana do użytkownika.
type Message struct {
	To      string    `json:"to"`
	Subject string    `json:"subject"`
	Body    string    `json:"body"`
	SentAt  time.Time `json:"sentAt"`
}

// Kontrakt kanału powiadomień.
// Serwisy zależą od interfejsu, więc symulator można zastąpić prawdziwym
// klientem SMTP bez zmian w logice biznesowej.
type Notifier interface {
	Send(ctx context.Context, m Message) error
}

// SYMULATOR serwisu e-mail: zamiast wysyłać wiadomość,
// zapisuje ją do logu i do skrzynki nadawczej w pamięci (podgląd w UI).
type EmailSimulator struct {
	mu       sync.Mutex
	outbox   []Message
	capacity int
	logger   *log.Logger
}

// Tworzy symulator przechowujący ostatnie capacity wiadomości.
func NewEmailSimulator(logger *log.Logger, capacity int) *EmailSimulator {
	return &EmailSimulator{logger: logger, capacity: capacity}
}

func (e *EmailSimulator) Send(_ context.Context, m Message) error {
	if m.SentAt.IsZero() {
		m.SentAt = time.Now()
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.outbox = append(e.outbox, m)
	if len(e.outbox) > e.capacity {
		e.outbox = e.outbox[len(e.outbox)-e.capacity:]
	}
	e.logger.Printf("[e-mail] do=%s temat=%q", m.To, m.Subject)
	return nil
}

// Zwraca kopię wysłanych wiadomości (najnowsze na początku).
func (e *EmailSimulator) Outbox() []Message {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]Message, len(e.outbox))
	for i, m := range e.outbox {
		out[len(e.outbox)-1-i] = m
	}
	return out
}
