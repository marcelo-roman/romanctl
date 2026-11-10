package main

import (
	"context"

	"sample-app/internal/services/orders/adapters/memory"
)

func openStorage(context.Context) (storage, error) {
	return storage{
		orders: memory.NewStore(),
		// romanctl:repositories
		close: func() {},
	}, nil
}
