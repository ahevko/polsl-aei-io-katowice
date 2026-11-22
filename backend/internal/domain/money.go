package domain

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

// Kod waluty obsługiwanej przez system.
type Currency string

const (
	PLN Currency = "PLN"
	EUR Currency = "EUR"
	USD Currency = "USD"
)

// Słownik walut dostępnych w systemie.
var SupportedCurrencies = []Currency{PLN, EUR, USD}

// Sprawdza, czy waluta należy do słownika.
func (c Currency) IsValid() bool {
	return slices.Contains(SupportedCurrencies, c)
}

// Zamienia tekst (np. "eur") na walutę.
func ParseCurrency(s string) (Currency, error) {
	c := Currency(strings.ToUpper(strings.TrimSpace(s)))
	if !c.IsValid() {
		return "", invalid("currency", "nieobsługiwana waluta (dozwolone: PLN, EUR, USD)")
	}
	return c, nil
}

// Obiekt wartości: kwota w jednostkach podstawowych (grosze / centy),
// dzięki czemu unikamy błędów zaokrągleń typu float.
type Money struct {
	Minor    int64    `json:"minor"`
	Currency Currency `json:"currency"`
}

// Tworzy kwotę z wartości dziesiętnej (np. 49.99).
func NewMoney(amount float64, c Currency) (Money, error) {
	if !c.IsValid() {
		return Money{}, invalid("currency", "nieobsługiwana waluta")
	}
	if amount < 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return Money{}, invalid("amount", "kwota musi być liczbą nieujemną")
	}
	return Money{Minor: int64(math.Round(amount * 100)), Currency: c}, nil
}

// Zwraca kwotę jako liczbę dziesiętną.
func (m Money) Amount() float64 { return float64(m.Minor) / 100 }

// Dodaje dwie kwoty w tej samej walucie.
func (m Money) Add(o Money) (Money, error) {
	if m.Currency != o.Currency {
		return Money{}, fmt.Errorf("%w: różne waluty %s i %s", ErrValidation, m.Currency, o.Currency)
	}
	return Money{Minor: m.Minor + o.Minor, Currency: m.Currency}, nil
}

// Mnoży kwotę przez ułamek num/den z zaokrągleniem do grosza.
func (m Money) MulRatio(num, den int64) Money {
	return Money{Minor: int64(math.Round(float64(m.Minor) * float64(num) / float64(den))), Currency: m.Currency}
}

func (m Money) String() string { return fmt.Sprintf("%.2f %s", m.Amount(), m.Currency) }
