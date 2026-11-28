// Package memory – implementacja repozytoriów w pamięci operacyjnej
// z opcjonalnym zapisem migawki do pliku JSON (uproszczona „baza danych”
// prototypu #1).
package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

// snapshot – całość danych zapisywana do pliku.
type snapshot struct {
	Users         map[string]domain.User         `json:"users"`
	Subscriptions map[string]domain.Subscription `json:"subscriptions"`
	Categories    map[string]domain.Category     `json:"categories"`
	Payments      map[string]domain.Payment      `json:"payments"`
	Notifications map[string]domain.Notification `json:"notifications"`
}

// Store – współdzielony magazyn danych wszystkich repozytoriów in-memory.
// Dostęp chroniony muteksem (Worker działa równolegle z obsługą żądań HTTP).
type Store struct {
	mu       sync.RWMutex
	data     snapshot
	filePath string
}

// NewStore tworzy magazyn; gdy filePath != "" dane są wczytywane z pliku
// i zapisywane do niego po każdej modyfikacji.
func NewStore(filePath string) (*Store, error) {
	s := &Store{filePath: filePath, data: snapshot{
		Users:         map[string]domain.User{},
		Subscriptions: map[string]domain.Subscription{},
		Categories:    map[string]domain.Category{},
		Payments:      map[string]domain.Payment{},
		Notifications: map[string]domain.Notification{},
	}}
	if filePath == "" {
		return s, nil
	}
	raw, err := os.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &s.data); err != nil {
		return nil, fmt.Errorf("plik danych %s jest uszkodzony: %w", filePath, err)
	}
	return s, nil
}

// write wykonuje modyfikację pod blokadą i utrwala dane.
func (s *Store) write(fn func(d *snapshot) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fn(&s.data); err != nil {
		return err
	}
	return s.persistLocked()
}

// read wykonuje odczyt pod blokadą współdzieloną.
func (s *Store) read(fn func(d *snapshot)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fn(&s.data)
}

func (s *Store) persistLocked() error {
	if s.filePath == "" {
		return nil
	}
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(s.filePath); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	tmp := s.filePath + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.filePath) // zapis atomowy
}
