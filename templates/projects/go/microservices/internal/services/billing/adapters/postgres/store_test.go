package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	platform "sample-app/internal/platform/postgres"
	"sample-app/internal/services/billing/application/invoices"
	"sample-app/internal/services/billing/domain/invoice"
	"sample-app/internal/services/billing/domain/shared"
)

func openStore(t *testing.T) *Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set: romanctl up sample-app --no-run creates the database and the .env")
	}
	ctx := context.Background()
	pool, err := platform.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return NewStore(pool)
}

func TestInboxDeduplicatesEvents(t *testing.T) {
	store := openStore(t)
	ctx := context.Background()
	order := shared.NewID()
	event := string(shared.NewID())

	first, _ := invoice.Issue(string(order), 750, "BRL", time.Now().UTC().Truncate(time.Microsecond))
	second, _ := invoice.Issue(string(order), 750, "BRL", time.Now().UTC().Truncate(time.Microsecond))
	for _, i := range []*invoice.Invoice{first, second} {
		if err := store.Issue(ctx, event, i); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Issue(ctx, string(shared.NewID()), second); err != nil {
		t.Fatal(err)
	}
	got, err := store.ByOrder(ctx, order)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID() != first.ID() || got.Amount() != 750 || got.Status() != invoice.Issued {
		t.Fatalf("invoice %+v", got)
	}
	if _, err := store.ByOrder(ctx, shared.NewID()); !errors.Is(err, invoices.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
