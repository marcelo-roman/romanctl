package order

import (
	"strings"

	"sample-app/internal/domain/shared"
)

type Money struct {
	cents    int64
	currency string
}

func NewMoney(cents int64, currency string) (Money, error) {
	if cents < 0 {
		return Money{}, shared.Rule("O valor não pode ser negativo.")
	}
	if len(strings.TrimSpace(currency)) != 3 {
		return Money{}, shared.Rule("A moeda deve ter 3 letras, como BRL.")
	}
	return Money{cents: cents, currency: strings.ToUpper(currency)}, nil
}

func (m Money) Cents() int64     { return m.cents }
func (m Money) Currency() string { return m.currency }

func (m Money) Multiply(quantity int) Money {
	return Money{cents: m.cents * int64(quantity), currency: m.currency}
}

func (m Money) Add(other Money) (Money, error) {
	if other.currency != m.currency {
		return Money{}, shared.Rule("Não é possível somar moedas diferentes.")
	}
	return Money{cents: m.cents + other.cents, currency: m.currency}, nil
}
