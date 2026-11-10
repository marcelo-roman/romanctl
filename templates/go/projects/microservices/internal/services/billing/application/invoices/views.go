package invoices

import (
	"time"

	"sample-app/internal/services/billing/domain/invoice"
)

type View struct {
	ID       string
	OrderID  string
	Amount   int64
	Currency string
	Status   string
	IssuedAt time.Time
}

func viewOf(i *invoice.Invoice) View {
	return View{ID: string(i.ID()), OrderID: string(i.OrderID()), Amount: i.Amount(), Currency: i.Currency(),
		Status: string(i.Status()), IssuedAt: i.IssuedAt()}
}
