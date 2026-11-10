package order

import (
	"strings"
	"time"

	"sample-app/internal/domain/shared"
)

const DefaultCurrency = "BRL"

type Order struct {
	id           shared.ID
	customerName string
	status       Status
	createdAt    time.Time
	items        []Item
	events       []Event
}

func Create(customerName string, now time.Time) (*Order, error) {
	if strings.TrimSpace(customerName) == "" {
		return nil, shared.Rule("O nome do cliente é obrigatório.")
	}
	return &Order{id: shared.NewID(), customerName: strings.TrimSpace(customerName), status: Draft, createdAt: now}, nil
}

func Restore(id shared.ID, customerName string, status Status, createdAt time.Time, items []Item) *Order {
	return &Order{id: id, customerName: customerName, status: status, createdAt: createdAt, items: items}
}

func (o *Order) ID() shared.ID        { return o.id }
func (o *Order) CustomerName() string { return o.customerName }
func (o *Order) Status() Status       { return o.status }
func (o *Order) CreatedAt() time.Time { return o.createdAt }
func (o *Order) Items() []Item        { return append([]Item(nil), o.items...) }

func (o *Order) Total() Money {
	currency := DefaultCurrency
	if len(o.items) > 0 {
		currency = o.items[0].unitPrice.currency
	}
	total := Money{currency: currency}
	for _, item := range o.items {
		total.cents += item.Subtotal().cents
	}
	return total
}

func (o *Order) AddItem(productName string, quantity int, unitPrice Money) error {
	if o.status != Draft {
		return shared.Rule("Um pedido já realizado não pode ser alterado.")
	}
	if len(o.items) > 0 && o.items[0].unitPrice.currency != unitPrice.currency {
		return shared.Rule("Todos os itens do pedido devem usar a mesma moeda.")
	}
	item, err := newItem(productName, quantity, unitPrice)
	if err != nil {
		return err
	}
	o.items = append(o.items, item)
	return nil
}

func (o *Order) Place(now time.Time) error {
	if o.status != Draft {
		return shared.Rule("O pedido já foi realizado.")
	}
	if len(o.items) == 0 {
		return shared.Rule("Um pedido sem itens não pode ser realizado.")
	}
	o.status = Placed
	total := o.Total()
	o.events = append(o.events, PlacedEvent{OrderID: o.id, Total: total.cents, Currency: total.currency, At: now})
	return nil
}

func (o *Order) PullEvents() []Event {
	events := o.events
	o.events = nil
	return events
}
