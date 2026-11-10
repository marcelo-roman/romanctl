package invoice

import (
	"errors"
	"testing"
	"time"

	"sample-app/internal/services/billing/domain/shared"
)

func isRule(err error) bool {
	var r *shared.RuleError
	return errors.As(err, &r)
}

func TestIssueCreatesAnIssuedInvoice(t *testing.T) {
	order := shared.NewID()
	i, err := Issue(string(order), 750, " brl ", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if i.OrderID() != order || i.Amount() != 750 || i.Currency() != "BRL" || i.Status() != Issued || i.ID() == "" {
		t.Fatalf("invoice %+v", i)
	}
}

func TestIssueRules(t *testing.T) {
	order := string(shared.NewID())
	cases := map[string]func() error{
		"invalid order":   func() error { _, err := Issue("x", 1, "BRL", time.Now()); return err },
		"negative amount": func() error { _, err := Issue(order, -1, "BRL", time.Now()); return err },
		"bad currency":    func() error { _, err := Issue(order, 1, "R$", time.Now()); return err },
	}
	for name, issue := range cases {
		if err := issue(); !isRule(err) {
			t.Errorf("%s: expected rule error, got %v", name, err)
		}
	}
}
