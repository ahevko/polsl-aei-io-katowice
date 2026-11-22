package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("nie znaleziono zasobu")
	ErrValidation   = errors.New("błąd walidacji")
	ErrConflict     = errors.New("konflikt danych")
	ErrUnauthorized = errors.New("brak autoryzacji")
	ErrInvalidState = errors.New("niedozwolona zmiana stanu")
)

// Opisuje niepoprawną wartość konkretnego pola.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return fmt.Sprintf("%s: %s", e.Field, e.Message) }

// Pozwala używać errors.Is(err, ErrValidation).
func (e *ValidationError) Is(target error) bool { return target == ErrValidation }

func invalid(field, msg string) error { return &ValidationError{Field: field, Message: msg} }
