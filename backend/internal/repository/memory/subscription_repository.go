package memory

import (
	"context"
	"sort"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

// SubscriptionRepository – implementacja in-memory repozytorium subskrypcji.
type SubscriptionRepository struct{ store *Store }

func NewSubscriptionRepository(s *Store) *SubscriptionRepository {
	return &SubscriptionRepository{store: s}
}

func cloneSub(s domain.Subscription) *domain.Subscription {
	s.CategoryIDs = append([]string(nil), s.CategoryIDs...)
	return &s
}

func (r *SubscriptionRepository) Add(_ context.Context, s *domain.Subscription) error {
	return r.store.write(func(d *snapshot) error {
		d.Subscriptions[s.ID] = *cloneSub(*s)
		return nil
	})
}

func (r *SubscriptionRepository) GetByID(_ context.Context, id string) (*domain.Subscription, error) {
	var (
		s  domain.Subscription
		ok bool
	)
	r.store.read(func(d *snapshot) { s, ok = d.Subscriptions[id] })
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cloneSub(s), nil
}

func (r *SubscriptionRepository) filter(pred func(domain.Subscription) bool) []*domain.Subscription {
	var out []*domain.Subscription
	r.store.read(func(d *snapshot) {
		for _, s := range d.Subscriptions {
			if pred(s) {
				out = append(out, cloneSub(s))
			}
		}
	})
	sort.Slice(out, func(i, j int) bool {
		if !out[i].NextPaymentDate.Equal(out[j].NextPaymentDate) {
			return out[i].NextPaymentDate.Before(out[j].NextPaymentDate)
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (r *SubscriptionRepository) ListByUser(_ context.Context, userID string) ([]*domain.Subscription, error) {
	return r.filter(func(s domain.Subscription) bool { return s.UserID == userID }), nil
}

func (r *SubscriptionRepository) ListActive(_ context.Context) ([]*domain.Subscription, error) {
	return r.filter(func(s domain.Subscription) bool { return s.IsActive() }), nil
}

func (r *SubscriptionRepository) ListDue(_ context.Context, day time.Time) ([]*domain.Subscription, error) {
	return r.filter(func(s domain.Subscription) bool { return s.IsDue(day) }), nil
}

func (r *SubscriptionRepository) Update(_ context.Context, s *domain.Subscription) error {
	return r.store.write(func(d *snapshot) error {
		if _, ok := d.Subscriptions[s.ID]; !ok {
			return domain.ErrNotFound
		}
		d.Subscriptions[s.ID] = *cloneSub(*s)
		return nil
	})
}

func (r *SubscriptionRepository) Delete(_ context.Context, id string) error {
	return r.store.write(func(d *snapshot) error {
		if _, ok := d.Subscriptions[id]; !ok {
			return domain.ErrNotFound
		}
		delete(d.Subscriptions, id)
		return nil
	})
}
