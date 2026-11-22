package memory

import (
	"context"
	"sort"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

// Implementacja in-memory słownika kategorii.
type CategoryRepository struct{ store *Store }

func NewCategoryRepository(s *Store) *CategoryRepository { return &CategoryRepository{store: s} }

func (r *CategoryRepository) List(_ context.Context) ([]domain.Category, error) {
	var out []domain.Category
	r.store.read(func(d *snapshot) {
		for _, c := range d.Categories {
			out = append(out, c)
		}
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (r *CategoryRepository) Add(_ context.Context, c domain.Category) error {
	return r.store.write(func(d *snapshot) error {
		d.Categories[c.ID] = c
		return nil
	})
}

func (r *CategoryRepository) ExistAll(_ context.Context, ids []string) (bool, error) {
	ok := true
	r.store.read(func(d *snapshot) {
		for _, id := range ids {
			if _, found := d.Categories[id]; !found {
				ok = false
				return
			}
		}
	})
	return ok, nil
}
