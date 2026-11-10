package rabbitmq

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"sample-app/internal/platform/messaging"
)

type Topology struct {
	Exchange string
	Queue    string
	Bindings []string
}

type Consumer struct {
	url        string
	topology   Topology
	prefetch   int
	RetryDelay time.Duration
}

func NewConsumer(url string, topology Topology) *Consumer {
	return &Consumer{url: url, topology: topology, prefetch: 10, RetryDelay: time.Second}
}

func (c *Consumer) Run(ctx context.Context, handler messaging.Handler) error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := c.declare(ch); err != nil {
		return err
	}
	deliveries, err := ch.ConsumeWithContext(ctx, c.topology.Queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case d, open := <-deliveries:
			if !open {
				return errors.New("a conexão com o RabbitMQ caiu")
			}
			if err := c.deliver(ctx, handler, d); err != nil {
				return err
			}
		}
	}
}

func (c *Consumer) deliver(ctx context.Context, handler messaging.Handler, d amqp.Delivery) error {
	e := envelopeOf(d)
	if err := e.Validate(); err != nil {
		return d.Nack(false, false)
	}
	err := handler(ctx, e)
	if err == nil {
		return d.Ack(false)
	}
	if errors.Is(err, messaging.ErrRejected) {
		return d.Nack(false, false)
	}
	select {
	case <-ctx.Done():
	case <-time.After(c.RetryDelay):
	}
	return d.Nack(false, true)
}

func (c *Consumer) declare(ch *amqp.Channel) error {
	if err := declareExchange(ch, c.topology.Exchange); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(c.topology.Queue, true, false, false, false, nil); err != nil {
		return err
	}
	for _, key := range c.topology.Bindings {
		if err := ch.QueueBind(c.topology.Queue, key, c.topology.Exchange, false, nil); err != nil {
			return err
		}
	}
	return ch.Qos(c.prefetch, 0, false)
}

func declareExchange(ch *amqp.Channel, name string) error {
	return ch.ExchangeDeclare(name, "topic", true, false, false, false, nil)
}

func envelopeOf(d amqp.Delivery) messaging.Envelope {
	eventType := d.Type
	if eventType == "" {
		eventType = d.RoutingKey
	}
	return messaging.Envelope{ID: d.MessageId, Type: eventType, OccurredAt: d.Timestamp, Payload: d.Body}
}
