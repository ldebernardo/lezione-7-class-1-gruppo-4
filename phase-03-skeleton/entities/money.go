package entities

import "regexp"

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type Money struct {
	amountCents int64
	currency    string
}

func NewMoney(amountCents int64, currency string) (Money, error) {
	if amountCents < 0 || !currencyPattern.MatchString(currency) {
		return Money{}, ErrInvalidMoney
	}
	return Money{amountCents: amountCents, currency: currency}, nil
}

func (m Money) AmountCents() int64 {
	return m.amountCents
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Equal(other Money) bool {
	return m.amountCents == other.amountCents && m.currency == other.currency
}
