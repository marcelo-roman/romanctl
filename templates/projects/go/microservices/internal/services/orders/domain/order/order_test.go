package order

import (
	"errors"
	"testing"
	"time"

	"sample-app/internal/services/orders/domain/shared"
)

var now = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func brl(t *testing.T, cents int64) Money {
	t.Helper()
	m, err := NewMoney(cents, "brl")
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func isRule(err error) bool {
	var r *shared.RuleError
	return errors.As(err, &r)
}

func TestCreateRequiresCustomer(t *testing.T) {
	if _, err := Create("  ", now); !isRule(err) {
		t.Fatalf("expected rule error, got %v", err)
	}
}

func TestTotalSumsItems(t *testing.T) {
	o, _ := Create("Ana", now)
	if err := o.AddItem("Caneta", 2, brl(t, 350)); err != nil {
		t.Fatal(err)
	}
	if err := o.AddItem("Caderno", 1, brl(t, 1200)); err != nil {
		t.Fatal(err)
	}
	if got := o.Total(); got.Cents() != 1900 || got.Currency() != "BRL" {
		t.Fatalf("total %d %s", got.Cents(), got.Currency())
	}
}

func TestItemsMustShareCurrency(t *testing.T) {
	o, _ := Create("Ana", now)
	_ = o.AddItem("Caneta", 1, brl(t, 100))
	usd, _ := NewMoney(100, "USD")
	if err := o.AddItem("Pen", 1, usd); !isRule(err) {
		t.Fatalf("expected rule error, got %v", err)
	}
}

func TestPlaceRaisesEventOnce(t *testing.T) {
	o, _ := Create("Ana", now)
	if err := o.Place(now); !isRule(err) {
		t.Fatalf("placing an empty order: %v", err)
	}
	_ = o.AddItem("Caneta", 1, brl(t, 100))
	if err := o.Place(now); err != nil {
		t.Fatal(err)
	}
	events := o.PullEvents()
	if len(events) != 1 || events[0].(PlacedEvent).Total != 100 {
		t.Fatalf("events %+v", events)
	}
	if len(o.PullEvents()) != 0 {
		t.Fatal("events were not cleared")
	}
	if err := o.AddItem("Lápis", 1, brl(t, 100)); !isRule(err) {
		t.Fatalf("changed a placed order: %v", err)
	}
}

func TestNegativeMoneyIsRejected(t *testing.T) {
	if _, err := NewMoney(-1, "BRL"); !isRule(err) {
		t.Fatalf("expected rule error, got %v", err)
	}
}
