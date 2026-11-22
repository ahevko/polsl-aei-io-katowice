package memory

import (
	"context"
	"sort"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

// Implementacja in-memory historii płatności.
type PaymentRepository struct{ store *Store }

func NewPaymentRepository(s *Store) *PaymentRepository { return &PaymentRepository{store: s} }

func (r *PaymentRepository) Add(_ context.Context, p *domain.Payment) error {
	return r.store.write(func(d *snapshot) error {
		d.Payments[p.ID] = *p
		return nil
	})
}

func (r *PaymentRepository) ListByUser(_ context.Context, userID string) ([]*domain.Payment, error) {
	var out []*domain.Payment
	r.store.read(func(d *snapshot) {
		for _, p := range d.Payments {
			if p.UserID == userID {
				c := p
				out = append(out, &c)
			}
		}
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	return out, nil
}

func (r *PaymentRepository) DeleteBySubscription(_ context.Context, subscriptionID string) error {
	return r.store.write(func(d *snapshot) error {
		for id, p := range d.Payments {
			if p.SubscriptionID == subscriptionID {
				delete(d.Payments, id)
			}
		}
		return nil
	})
}
