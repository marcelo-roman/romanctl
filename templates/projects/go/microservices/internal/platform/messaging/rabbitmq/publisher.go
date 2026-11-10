package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"sample-app/internal/platform/messaging"
)

type Publisher struct {
	url      string
	exchange string
	mu       sync.Mutex
	conn     *amqp.Connection
	ch       *amqp.Channel
	returns  chan amqp.Return
}

func NewPublisher(url, exchange string) *Publisher {
	return &Publisher{url: url, exchange: exchange}
}

func (p *Publisher) Publish(ctx context.Context, e messaging.Envelope) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.connect(); err != nil {
		return err
	}
	confirm, err := p.ch.PublishWithDeferredConfirmWithContext(ctx, p.exchange, e.Type, true, false, amqp.Publishing{
		MessageId:    e.ID,
		Type:         e.Type,
		Timestamp:    e.OccurredAt,
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         e.Payload,
	})
	if err != nil {
		p.reset()
		return err
	}
	acked, err := confirm.WaitContext(ctx)
	if err != nil {
		p.reset()
		return err
	}
	if !acked {
		return fmt.Errorf("o RabbitMQ recusou a mensagem %s", e.ID)
	}
	return p.unroutable(e)
}

func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reset()
}

func (p *Publisher) unroutable(e messaging.Envelope) error {
	for {
		select {
		case r := <-p.returns:
			if r.MessageId == e.ID {
				return fmt.Errorf("nenhuma fila recebe %s em %s ainda", e.Type, p.exchange)
			}
		default:
			return nil
		}
	}
}

func (p *Publisher) connect() error {
	if p.conn != nil && !p.conn.IsClosed() && p.ch != nil && !p.ch.IsClosed() {
		return nil
	}
	p.reset()
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	if err := declareExchange(ch, p.exchange); err != nil {
		conn.Close()
		return err
	}
	if err := ch.Confirm(false); err != nil {
		conn.Close()
		return err
	}
	p.conn, p.ch, p.returns = conn, ch, ch.NotifyReturn(make(chan amqp.Return, 16))
	return nil
}

func (p *Publisher) reset() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.conn, p.ch, p.returns = nil, nil, nil
}
