package postgres

import (
	"context"
	"os"
	"sync"
	"testing"
)

func TestConcurrentMigrationsDoNotCollide(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set: romanctl up --no-run creates the database and the .env")
	}
	ctx := context.Background()
	pool, err := Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DROP SCHEMA IF EXISTS migration_probe CASCADE") })

	schema := `CREATE SCHEMA IF NOT EXISTS migration_probe;
		CREATE TABLE IF NOT EXISTS migration_probe.items (id integer PRIMARY KEY);`
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- Migrate(ctx, pool, schema)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}
