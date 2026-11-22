package memory

import (
	"context"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

// Implementacja in-memory rejestru przypomnień.
type NotificationRepository struct{ store *Store }

func NewNotificationRepository(s *Store) *NotificationRepository {
	return &NotificationRepository{store: s}
}

func (r *NotificationRepository) Add(_ context.Context, n *domain.Notification) error {
	return r.store.write(func(d *snapshot) error {
		d.Notifications[n.Key()] = *n
		return nil
	})
}

func (r *NotificationRepository) Exists(_ context.Context, key string) (bool, error) {
	var ok bool
	r.store.read(func(d *snapshot) { _, ok = d.Notifications[key] })
	return ok, nil
}
