package service

import (
	"context"
	"math"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/currency"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

// Konwerter walut. Współpracuje z RateProvider, który pobiera aktualne kursy walut.
type CurrencyConverter struct {
	rates currency.RateProvider
}

func NewCurrencyConverter(rates currency.RateProvider) *CurrencyConverter {
	return &CurrencyConverter{rates: rates}
}

// Przelicza kwotę na walutę docelową (przez PLN jako walutę pośrednią).
func (c *CurrencyConverter) Convert(ctx context.Context, m domain.Money, to domain.Currency) (domain.Money, error) {
	if m.Currency == to {
		return m, nil
	}
	t, err := c.rates.CurrentRates(ctx)
	if err != nil {
		return domain.Money{}, err
	}
	from, err := t.PLNPer(m.Currency)
	if err != nil {
		return domain.Money{}, err
	}
	target, err := t.PLNPer(to)
	if err != nil {
		return domain.Money{}, err
	}
	return domain.Money{Minor: int64(math.Round(float64(m.Minor) * from / target)), Currency: to}, nil
}
