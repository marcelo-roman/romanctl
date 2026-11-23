# Messaging

RabbitMQ, with `github.com/rabbitmq/amqp091-go`, read from `RABBITMQ_URL`. `romanctl up sample-app` starts RabbitMQ and
writes the variable to `.env`.

| Piece | Name |
|---|---|
| Exchange | `sample-app.events`, of type `topic` and durable |
| Event key | `orders.order-placed` |
| Billing queue | `billing.order-placed`, durable |

`orders` publishes with broker confirms and with `mandatory`: while no queue is bound to the key, the message comes
back, the relay treats it as a failure and it stays in the outbox. So nothing is lost if `orders` starts before
`billing`.

`billing` declares the queue, consumes with manual acks and reconnects on its own if the connection drops. A transient
error returns the message to the queue; an event that will never be accepted, such as invalid JSON, is discarded.

The test in `internal/platform/messaging/rabbitmq` does the round trip on a real RabbitMQ and is skipped without
`RABBITMQ_URL`.
