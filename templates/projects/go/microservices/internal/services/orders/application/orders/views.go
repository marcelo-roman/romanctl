package orders

import (
	"time"

	"sample-app/internal/services/orders/domain/order"
)

type ItemView struct {
	ProductName string
	Quantity    int
	UnitPrice   int64
	Subtotal    int64
}

type OrderView struct {
	ID           string
	CustomerName string
	Status       string
	CreatedAt    time.Time
	Total        int64
	Currency     string
	Items        []ItemView
}

func viewOf(o *order.Order) OrderView {
	total := o.Total()
	v := OrderView{ID: string(o.ID()), CustomerName: o.CustomerName(), Status: string(o.Status()),
		CreatedAt: o.CreatedAt(), Total: total.Cents(), Currency: total.Currency(), Items: []ItemView{}}
	for _, item := range o.Items() {
		v.Items = append(v.Items, ItemView{ProductName: item.ProductName(), Quantity: item.Quantity(),
			UnitPrice: item.UnitPrice().Cents(), Subtotal: item.Subtotal().Cents()})
	}
	return v
}
