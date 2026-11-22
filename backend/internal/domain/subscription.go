package domain

import (
	"strings"
	"time"
)

// Status subskrypcji (słownik: "aktywna" / "wstrzymana").
type SubscriptionStatus string

const (
	StatusActive SubscriptionStatus = "ACTIVE"
	StatusPaused SubscriptionStatus = "PAUSED"
)

// Główna encja systemu.
type Subscription struct {
	ID              string             `json:"id"`
	UserID          string             `json:"userId"`
	Name            string             `json:"name"`
	Price           Money              `json:"price"`
	Cycle           BillingCycle       `json:"cycle"`
	NextPaymentDate time.Time          `json:"nextPaymentDate"`
	CategoryIDs     []string           `json:"categoryIds"`
	Status          SubscriptionStatus `json:"status"`
	CreatedAt       time.Time          `json:"createdAt"`
}

// Tworzy aktywną subskrypcję i waliduje jej dane.
func NewSubscription(userID, name string, price Money, cycle BillingCycle, next time.Time, categoryIDs []string, now time.Time) (*Subscription, error) {
	s := &Subscription{
		ID:              NewID(),
		UserID:          userID,
		Name:            strings.TrimSpace(name),
		Price:           price,
		Cycle:           cycle,
		NextPaymentDate: DateOnly(next),
		CategoryIDs:     append([]string(nil), categoryIDs...),
		Status:          StatusActive,
		CreatedAt:       now,
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return s, nil
}

// Sprawdza niezmienniki encji.
func (s *Subscription) Validate() error {
	if s.Name == "" {
		return invalid("name", "nazwa jest wymagana")
	}
	if len(s.Name) > 100 {
		return invalid("name", "nazwa może mieć maksymalnie 100 znaków")
	}
	if s.Price.Minor <= 0 {
		return invalid("amount", "koszt musi być większy od zera")
	}
	if !s.Price.Currency.IsValid() {
		return invalid("currency", "nieobsługiwana waluta")
	}
	if _, err := StrategyFor(s.Cycle); err != nil {
		return err
	}
	if s.NextPaymentDate.IsZero() {
		return invalid("nextPaymentDate", "data następnej płatności jest wymagana")
	}
	if len(s.CategoryIDs) == 0 {
		return invalid("categoryIds", "wybierz co najmniej jedną kategorię")
	}
	return nil
}

// Czy subskrypcja generuje płatności i powiadomienia.
func (s *Subscription) IsActive() bool { return s.Status == StatusActive }

// Wstrzymuje aktywną subskrypcję.
func (s *Subscription) Pause() error {
	if s.Status != StatusActive {
		return ErrInvalidState
	}
	s.Status = StatusPaused
	return nil
}

// Wznawia subskrypcję z nową datą następnej płatności.
func (s *Subscription) Resume(next time.Time) error {
	if s.Status != StatusPaused {
		return ErrInvalidState
	}
	if next.IsZero() {
		return invalid("nextPaymentDate", "podaj datę następnej płatności")
	}
	s.NextPaymentDate = DateOnly(next)
	s.Status = StatusActive
	return nil
}

// Czy w danym dniu przypada (lub minęła) płatność.
func (s *Subscription) IsDue(day time.Time) bool {
	return s.IsActive() && !s.NextPaymentDate.After(DateOnly(day))
}

// Przesuwa datę następnej płatności o jeden cykl.
func (s *Subscription) AdvanceBillingDate() error {
	st, err := StrategyFor(s.Cycle)
	if err != nil {
		return err
	}
	s.NextPaymentDate = st.Next(s.NextPaymentDate)
	return nil
}

// Przesuwa datę z przeszłości na najbliższy termin >= day
// bez generowania historii (używane przy dodawaniu/wznawianiu).
func (s *Subscription) RollForwardTo(day time.Time) error {
	day = DateOnly(day)
	for s.NextPaymentDate.Before(day) {
		if err := s.AdvanceBillingDate(); err != nil {
			return err
		}
	}
	return nil
}

// Koszt znormalizowany do jednego miesiąca.
func (s *Subscription) MonthlyCost() (Money, error) {
	st, err := StrategyFor(s.Cycle)
	if err != nil {
		return Money{}, err
	}
	return s.Price.MulRatio(st.PaymentsPerYear(), 12), nil
}

// Koszt w skali roku.
func (s *Subscription) YearlyCost() (Money, error) {
	st, err := StrategyFor(s.Cycle)
	if err != nil {
		return Money{}, err
	}
	return s.Price.MulRatio(st.PaymentsPerYear(), 1), nil
}
