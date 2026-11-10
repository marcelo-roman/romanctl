package main

import (
	"sample-app/internal/platform/messaging"
	"sample-app/internal/platform/messaging/rabbitmq"
)

const exchange = "sample-app.events"

func openPublisher() (messaging.Publisher, func()) {
	publisher := rabbitmq.NewPublisher(envOr("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672"), exchange)
	return publisher, publisher.Close
}
