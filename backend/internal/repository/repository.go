// Package repository definiuje kontrakty warstwy dostępu do danych
// (wzorzec REPOZYTORIUM). Warstwa logiki biznesowej zależy wyłącznie od tych
// interfejsów, a nie od konkretnej technologii składowania danych – w
// prototypie używamy implementacji in-memory (pakiet memory), docelowo można
// dodać implementację dla relacyjnej bazy danych bez zmian w serwisach.
package repository

import (
	"context"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

// UserRepository – dostęp do kont użytkowników.
type UserRepository interface {
	Add(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
}

// SubscriptionRepository – dostęp do subskrypcji.
type SubscriptionRepository interface {
	Add(ctx context.Context, s *domain.Subscription) error
	GetByID(ctx context.Context, id string) (*domain.Subscription, error)
	ListByUser(ctx context.Context, userID string) ([]*domain.Subscription, error)
	ListActive(ctx context.Context) ([]*domain.Subscription, error)
	ListDue(ctx context.Context, day time.Time) ([]*domain.Subscription, error)
	Update(ctx context.Context, s *domain.Subscription) error
	Delete(ctx context.Context, id string) error
}

// CategoryRepository – dostęp do słownika kategorii.
type CategoryRepository interface {
	List(ctx context.Context) ([]domain.Category, error)
	Add(ctx context.Context, c domain.Category) error
	ExistAll(ctx context.Context, ids []string) (bool, error)
}

// PaymentRepository – dostęp do historii płatności.
type PaymentRepository interface {
	Add(ctx context.Context, p *domain.Payment) error
	ListByUser(ctx context.Context, userID string) ([]*domain.Payment, error)
	DeleteBySubscription(ctx context.Context, subscriptionID string) error
}

// NotificationRepository – rejestr wysłanych przypomnień.
type NotificationRepository interface {
	Add(ctx context.Context, n *domain.Notification) error
	Exists(ctx context.Context, key string) (bool, error)
}
