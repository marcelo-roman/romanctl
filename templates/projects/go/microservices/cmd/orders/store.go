package main

import (
	"context"

	platform "sample-app/internal/platform/postgres"
	"sample-app/internal/services/orders/adapters/postgres"
)

func openStorage(ctx context.Context) (storage, error) {
	pool, err := platform.Open(ctx, envOr("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5432/sample_app"))
	if err != nil {
		return storage{}, err
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		pool.Close()
		return storage{}, err
	}
	return storage{
		orders: postgres.NewStore(pool),
		// romanctl:repositories
		close: pool.Close,
	}, nil
}
