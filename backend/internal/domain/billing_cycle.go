package domain

import (
	"strings"
	"time"
)

// BillingCycle – cykl rozliczeniowy subskrypcji.
type BillingCycle string

const (
	Weekly  BillingCycle = "WEEKLY"
	Monthly BillingCycle = "MONTHLY"
	Yearly  BillingCycle = "YEARLY"
)

// ParseBillingCycle zamienia tekst na cykl rozliczeniowy.
func ParseBillingCycle(s string) (BillingCycle, error) {
	c := BillingCycle(strings.ToUpper(strings.TrimSpace(s)))
	switch c {
	case Weekly, Monthly, Yearly:
		return c, nil
	}
	return "", invalid("cycle", "dozwolone cykle: WEEKLY, MONTHLY, YEARLY")
}

// CycleStrategy – wzorzec STRATEGIA: każdy cykl rozliczeniowy inaczej
// wylicza datę kolejnej płatności i liczbę płatności w roku.
type CycleStrategy interface {
	// Next zwraca datę płatności następującej po from.
	Next(from time.Time) time.Time
	// PaymentsPerYear służy do normalizacji kosztów (miesięczny / roczny).
	PaymentsPerYear() int64
}

// WeeklyCycle – płatność co 7 dni.
type WeeklyCycle struct{}

func (WeeklyCycle) Next(from time.Time) time.Time { return from.AddDate(0, 0, 7) }
func (WeeklyCycle) PaymentsPerYear() int64        { return 52 }

// MonthlyCycle – płatność co miesiąc (31.01 -> 28/29.02).
type MonthlyCycle struct{}

func (MonthlyCycle) Next(from time.Time) time.Time { return addMonthsClamped(from, 1) }
func (MonthlyCycle) PaymentsPerYear() int64        { return 12 }

// YearlyCycle – płatność co rok (29.02 -> 28.02).
type YearlyCycle struct{}

func (YearlyCycle) Next(from time.Time) time.Time { return addMonthsClamped(from, 12) }
func (YearlyCycle) PaymentsPerYear() int64        { return 1 }

// StrategyFor – prosta FABRYKA strategii dla danego cyklu.
func StrategyFor(c BillingCycle) (CycleStrategy, error) {
	switch c {
	case Weekly:
		return WeeklyCycle{}, nil
	case Monthly:
		return MonthlyCycle{}, nil
	case Yearly:
		return YearlyCycle{}, nil
	}
	return nil, invalid("cycle", "nieznany cykl rozliczeniowy")
}

// addMonthsClamped dodaje miesiące, przycinając dzień do ostatniego dnia
// miesiąca docelowego (time.AddDate przenosiłby 31.01 na 03.03).
func addMonthsClamped(t time.Time, months int) time.Time {
	y, m, d := t.Date()
	firstOfTarget := time.Date(y, m+time.Month(months), 1, 0, 0, 0, 0, t.Location())
	lastDay := firstOfTarget.AddDate(0, 1, -1).Day()
	if d > lastDay {
		d = lastDay
	}
	return time.Date(firstOfTarget.Year(), firstOfTarget.Month(), d, 0, 0, 0, 0, t.Location())
}

// DateOnly obcina czas do północy UTC – system operuje na datach kalendarzowych.
func DateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
