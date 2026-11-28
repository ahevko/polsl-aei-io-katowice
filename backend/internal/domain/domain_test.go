package domain

import (
	"errors"
	"testing"
	"time"
)

func d(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestCycleStrategies(t *testing.T) {
	cases := []struct {
		cycle BillingCycle
		from  string
		want  string
	}{
		{Weekly, "2026-11-17", "2026-11-24"},
		{Monthly, "2026-11-17", "2026-12-17"},
		{Monthly, "2026-01-31", "2026-02-28"}, // przycięcie do końca miesiąca
		{Monthly, "2026-12-15", "2027-01-15"},
		{Yearly, "2028-02-29", "2029-02-28"},
	}
	for _, c := range cases {
		st, err := StrategyFor(c.cycle)
		if err != nil {
			t.Fatal(err)
		}
		if got := st.Next(d(c.from)).Format("2006-01-02"); got != c.want {
			t.Errorf("%s od %s: got %s, want %s", c.cycle, c.from, got, c.want)
		}
	}
}

func newSub(t *testing.T, amount float64, cycle BillingCycle) *Subscription {
	t.Helper()
	m, err := NewMoney(amount, PLN)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSubscription("u1", "Netflix", m, cycle, d("2026-12-01"), []string{"streaming"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestCostNormalization(t *testing.T) {
	s := newSub(t, 12, Weekly)
	m, _ := s.MonthlyCost()
	y, _ := s.YearlyCost()
	if m.Minor != 5200 || y.Minor != 62400 { // 12*52/12 = 52.00, 12*52 = 624.00
		t.Errorf("weekly: monthly=%v yearly=%v", m, y)
	}
	s = newSub(t, 120, Yearly)
	m, _ = s.MonthlyCost()
	if m.Minor != 1000 {
		t.Errorf("yearly->monthly = %v", m)
	}
}

func TestPauseResume(t *testing.T) {
	s := newSub(t, 49, Monthly)
	if err := s.Pause(); err != nil {
		t.Fatal(err)
	}
	if s.IsDue(d("2027-01-01")) {
		t.Error("wstrzymana subskrypcja nie powinna być wymagalna")
	}
	if err := s.Pause(); !errors.Is(err, ErrInvalidState) {
		t.Errorf("ponowne wstrzymanie: %v", err)
	}
	if err := s.Resume(d("2027-02-10")); err != nil {
		t.Fatal(err)
	}
	if !s.IsActive() || !s.NextPaymentDate.Equal(d("2027-02-10")) {
		t.Errorf("po wznowieniu: %+v", s)
	}
}

func TestValidation(t *testing.T) {
	m, _ := NewMoney(10, EUR)
	_, err := NewSubscription("u1", "  ", m, Monthly, d("2026-12-01"), []string{"x"}, time.Now())
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "name" {
		t.Errorf("oczekiwano błędu pola name, got %v", err)
	}
	if _, err := NewMoney(-1, PLN); !errors.Is(err, ErrValidation) {
		t.Error("ujemna kwota powinna być odrzucona")
	}
	if _, err := ParseCurrency("GBP"); err == nil {
		t.Error("GBP nie jest obsługiwane")
	}
}

func TestRollForward(t *testing.T) {
	s := newSub(t, 10, Weekly)
	s.NextPaymentDate = d("2026-11-01")
	_ = s.RollForwardTo(d("2026-11-17"))
	if got := s.NextPaymentDate.Format("2006-01-02"); got != "2026-11-22" {
		t.Errorf("got %s", got)
	}
}
