package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrRejected = errors.New("mensagem rejeitada")

type Envelope struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	OccurredAt time.Time       `json:"occurredAt"`
	Payload    json.RawMessage `json:"payload"`
}

type Publisher interface {
	Publish(ctx context.Context, e Envelope) error
}

type Handler func(ctx context.Context, e Envelope) error

func NewEnvelope(id, eventType string, occurredAt time.Time, payload any) (Envelope, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{ID: id, Type: eventType, OccurredAt: occurredAt.UTC(), Payload: body}, nil
}

func (e Envelope) Validate() error {
	if e.ID == "" || e.Type == "" {
		return Reject(errors.New("o evento precisa de id e tipo"))
	}
	return nil
}

func Reject(err error) error { return fmt.Errorf("%w: %w", ErrRejected, err) }
