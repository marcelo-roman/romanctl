# Messaging

No broker: the `orders` relay delivers each event with a `POST` to `billing`'s `/events`, at the `BILLING_URL` address
(default `http://localhost:8092`). An in-process bus would not cross from one service to the other.

The outbox and the inbox still apply: the event only leaves the outbox when `billing` answers 2xx, and while `billing`
is down the relay retries with growing backoff. `billing` answers 204 when it processed the event, 503 on a transient
failure and 422 for an event that will never be accepted.

To switch to RabbitMQ, generate the project with `--with rabbitmq`.
