package invoice

import (
	"strings"
	"time"

	"sample-app/internal/services/billing/domain/shared"
)

type Status string

const Issued Status = "Issued"

type Invoice struct {
	id       shared.ID
	orderID  shared.ID
	amount   int64
	currency string
	status   Status
	issuedAt time.Time
}

func Issue(orderID string, amountCents int64, currency string, now time.Time) (*Invoice, error) {
	order, err := shared.ParseID(orderID)
	if err != nil {
		return nil, shared.Rule("An invoice needs a valid order.")
	}
	if amountCents < 0 {
		return nil, shared.Rule("Invoice amount cannot be negative.")
	}
	if len(strings.TrimSpace(currency)) != 3 {
		return nil, shared.Rule("Currency must have 3 letters, like BRL.")
	}
	return &Invoice{id: shared.NewID(), orderID: order, amount: amountCents, currency: strings.ToUpper(strings.TrimSpace(currency)),
		status: Issued, issuedAt: now}, nil
}

func Restore(id, orderID shared.ID, amountCents int64, currency string, status Status, issuedAt time.Time) *Invoice {
	return &Invoice{id: id, orderID: orderID, amount: amountCents, currency: currency, status: status, issuedAt: issuedAt}
}

func (i *Invoice) ID() shared.ID       { return i.id }
func (i *Invoice) OrderID() shared.ID  { return i.orderID }
func (i *Invoice) Amount() int64       { return i.amount }
func (i *Invoice) Currency() string    { return i.currency }
func (i *Invoice) Status() Status      { return i.status }
func (i *Invoice) IssuedAt() time.Time { return i.issuedAt }
