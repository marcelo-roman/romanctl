package order

import (
	"strings"

	"sample-app/internal/domain/shared"
)

type Item struct {
	id          shared.ID
	productName string
	quantity    int
	unitPrice   Money
}

func newItem(productName string, quantity int, unitPrice Money) (Item, error) {
	if strings.TrimSpace(productName) == "" {
		return Item{}, shared.Rule("O nome do produto é obrigatório.")
	}
	if quantity <= 0 {
		return Item{}, shared.Rule("A quantidade deve ser maior que zero.")
	}
	return Item{id: shared.NewID(), productName: strings.TrimSpace(productName), quantity: quantity, unitPrice: unitPrice}, nil
}

func RestoreItem(id shared.ID, productName string, quantity int, unitPrice Money) Item {
	return Item{id: id, productName: productName, quantity: quantity, unitPrice: unitPrice}
}

func (i Item) ID() shared.ID       { return i.id }
func (i Item) ProductName() string { return i.productName }
func (i Item) Quantity() int       { return i.quantity }
func (i Item) UnitPrice() Money    { return i.unitPrice }
func (i Item) Subtotal() Money     { return i.unitPrice.Multiply(i.quantity) }
