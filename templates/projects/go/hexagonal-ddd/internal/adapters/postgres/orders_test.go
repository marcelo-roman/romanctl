package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"sample-app/internal/application/orders"
	"sample-app/internal/domain/order"
	"sample-app/internal/domain/shared"
)

func TestOrderRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := NewOrderRepository(db)

	o, _ := order.Create("Ana", time.Now().UTC().Truncate(time.Microsecond))
	price, _ := order.NewMoney(350, "BRL")
	_ = o.AddItem("Caneta", 2, price)
	if err := repo.Save(ctx, o); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, o.ID())
	if err != nil {
		t.Fatal(err)
	}
	if got.Total().Cents() != 700 || len(got.Items()) != 1 || got.CustomerName() != "Ana" {
		t.Fatalf("loaded %+v", got)
	}
	if _, err := repo.Get(ctx, shared.NewID()); !errors.Is(err, orders.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
