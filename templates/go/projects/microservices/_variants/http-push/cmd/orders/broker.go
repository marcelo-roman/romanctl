package main

import (
	"strings"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/platform/messaging/httppush"
)

func openPublisher() (messaging.Publisher, func()) {
	return httppush.NewPublisher(strings.TrimSuffix(envOr("BILLING_URL", "http://localhost:8092"), "/") + "/events"), func() {}
}
