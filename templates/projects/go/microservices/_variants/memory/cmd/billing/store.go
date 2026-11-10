package main

import (
	"context"

	"sample-app/internal/services/billing/adapters/memory"
)

func openStorage(context.Context) (storage, error) {
	return storage{invoices: memory.NewStore(), close: func() {}}, nil
}
