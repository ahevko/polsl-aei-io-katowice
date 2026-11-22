package memory

import (
	"context"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

// Implementacja in-memory repozytorium użytkowników.
type UserRepository struct{ store *Store }

func NewUserRepository(s *Store) *UserRepository { return &UserRepository{store: s} }

func (r *UserRepository) Add(_ context.Context, u *domain.User) error {
	return r.store.write(func(d *snapshot) error {
		for _, existing := range d.Users {
			if existing.Email == u.Email {
				return domain.ErrConflict
			}
		}
		d.Users[u.ID] = *u
		return nil
	})
}

func (r *UserRepository) GetByID(_ context.Context, id string) (*domain.User, error) {
	var (
		u  domain.User
		ok bool
	)
	r.store.read(func(d *snapshot) { u, ok = d.Users[id] })
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	var found *domain.User
	r.store.read(func(d *snapshot) {
		for _, u := range d.Users {
			if u.Email == email {
				c := u
				found = &c
				return
			}
		}
	})
	if found == nil {
		return nil, domain.ErrNotFound
	}
	return found, nil
}

func (r *UserRepository) Update(_ context.Context, u *domain.User) error {
	return r.store.write(func(d *snapshot) error {
		if _, ok := d.Users[u.ID]; !ok {
			return domain.ErrNotFound
		}
		d.Users[u.ID] = *u
		return nil
	})
}
