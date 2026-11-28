package service

import (
	"context"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository"
)

// SubscriptionInput – dane wejściowe formularza subskrypcji (dodawanie/edycja).
type SubscriptionInput struct {
	Name            string
	Amount          float64
	Currency        string
	Cycle           string
	NextPaymentDate time.Time
	CategoryIDs     []string
}

// SubscriptionService – UC z diagramu „Zarządzanie subskrypcjami” (WF-01, WF-04).
type SubscriptionService struct {
	subs       repository.SubscriptionRepository
	categories repository.CategoryRepository
	payments   repository.PaymentRepository
	clock      Clock
}

func NewSubscriptionService(subs repository.SubscriptionRepository, categories repository.CategoryRepository, payments repository.PaymentRepository, clock Clock) *SubscriptionService {
	return &SubscriptionService{subs: subs, categories: categories, payments: payments, clock: clock}
}

// List – UC „Przeglądaj listę subskrypcji”.
func (s *SubscriptionService) List(ctx context.Context, userID string) ([]*domain.Subscription, error) {
	return s.subs.ListByUser(ctx, userID)
}

// Get zwraca subskrypcję, sprawdzając, czy należy do użytkownika
// (WF-05: użytkownik widzi wyłącznie własne dane).
func (s *SubscriptionService) Get(ctx context.Context, userID, id string) (*domain.Subscription, error) {
	sub, err := s.subs.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, domain.ErrNotFound // nie ujawniamy istnienia cudzych danych
	}
	return sub, nil
}

// Create – UC „Dodaj subskrypcję” (include: Przypisz kategorię,
// Wylicz datę następnej płatności).
func (s *SubscriptionService) Create(ctx context.Context, userID string, in SubscriptionInput) (*domain.Subscription, error) {
	price, cycle, err := s.parse(ctx, in)
	if err != nil {
		return nil, err
	}
	sub, err := domain.NewSubscription(userID, in.Name, price, cycle, in.NextPaymentDate, in.CategoryIDs, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := sub.RollForwardTo(s.clock.Now()); err != nil {
		return nil, err
	}
	return sub, s.subs.Add(ctx, sub)
}

// Update – UC „Edytuj subskrypcję”.
func (s *SubscriptionService) Update(ctx context.Context, userID, id string, in SubscriptionInput) (*domain.Subscription, error) {
	sub, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	price, cycle, err := s.parse(ctx, in)
	if err != nil {
		return nil, err
	}
	sub.Name, sub.Price, sub.Cycle = in.Name, price, cycle
	sub.NextPaymentDate = domain.DateOnly(in.NextPaymentDate)
	sub.CategoryIDs = append([]string(nil), in.CategoryIDs...)
	if err := sub.Validate(); err != nil {
		return nil, err
	}
	if err := sub.RollForwardTo(s.clock.Now()); err != nil {
		return nil, err
	}
	return sub, s.subs.Update(ctx, sub)
}

// Delete – UC „Usuń subskrypcję”: usuwa również jej historię płatności.
func (s *SubscriptionService) Delete(ctx context.Context, userID, id string) error {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return err
	}
	if err := s.payments.DeleteBySubscription(ctx, id); err != nil {
		return err
	}
	return s.subs.Delete(ctx, id)
}

// Pause – UC „Wstrzymaj subskrypcję”.
func (s *SubscriptionService) Pause(ctx context.Context, userID, id string) (*domain.Subscription, error) {
	sub, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := sub.Pause(); err != nil {
		return nil, err
	}
	return sub, s.subs.Update(ctx, sub)
}

// Resume – UC „Wznów subskrypcję” (include: Wylicz datę następnej płatności).
func (s *SubscriptionService) Resume(ctx context.Context, userID, id string, next time.Time) (*domain.Subscription, error) {
	sub, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := sub.Resume(next); err != nil {
		return nil, err
	}
	if err := sub.RollForwardTo(s.clock.Now()); err != nil {
		return nil, err
	}
	return sub, s.subs.Update(ctx, sub)
}

// parse waliduje pola wspólne dla dodawania i edycji (UC „Przypisz kategorię”).
func (s *SubscriptionService) parse(ctx context.Context, in SubscriptionInput) (domain.Money, domain.BillingCycle, error) {
	cur, err := domain.ParseCurrency(in.Currency)
	if err != nil {
		return domain.Money{}, "", err
	}
	price, err := domain.NewMoney(in.Amount, cur)
	if err != nil {
		return domain.Money{}, "", err
	}
	cycle, err := domain.ParseBillingCycle(in.Cycle)
	if err != nil {
		return domain.Money{}, "", err
	}
	if len(in.CategoryIDs) > 0 {
		ok, err := s.categories.ExistAll(ctx, in.CategoryIDs)
		if err != nil {
			return domain.Money{}, "", err
		}
		if !ok {
			return domain.Money{}, "", &domain.ValidationError{Field: "categoryIds", Message: "nieznana kategoria"}
		}
	}
	return price, cycle, nil
}

// CategoryService – udostępnia słownik kategorii (WF-04).
type CategoryService struct {
	categories repository.CategoryRepository
}

func NewCategoryService(categories repository.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

// List zwraca dostępne kategorie.
func (s *CategoryService) List(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}

// EnsureDefaults zasila słownik kategoriami domyślnymi (start systemu).
func (s *CategoryService) EnsureDefaults(ctx context.Context) error {
	existing, err := s.categories.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	for _, c := range domain.DefaultCategories() {
		if err := s.categories.Add(ctx, c); err != nil {
			return err
		}
	}
	return nil
}
