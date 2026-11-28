// Package domain zawiera model dziedzinowy systemu (encje, obiekty wartości
// i reguły biznesowe niezależne od infrastruktury).
package domain

import (
	"errors"
	"fmt"
)

// Błędy dziedzinowe – warstwa prezentacji mapuje je na kody HTTP.
var (
	ErrNotFound     = errors.New("nie znaleziono zasobu")
	ErrValidation   = errors.New("błąd walidacji")
	ErrConflict     = errors.New("konflikt danych")
	ErrUnauthorized = errors.New("brak autoryzacji")
	ErrInvalidState = errors.New("niedozwolona zmiana stanu")
)

// ValidationError opisuje niepoprawną wartość konkretnego pola.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return fmt.Sprintf("%s: %s", e.Field, e.Message) }

// Is pozwala używać errors.Is(err, ErrValidation).
func (e *ValidationError) Is(target error) bool { return target == ErrValidation }

func invalid(field, msg string) error { return &ValidationError{Field: field, Message: msg} }
