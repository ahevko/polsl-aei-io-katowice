package domain

import (
	"net/mail"
	"strings"
	"time"
)

// Rola użytkownika w systemie.
type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

// Konto użytkownika.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"passwordHash"`
	BaseCurrency Currency  `json:"baseCurrency"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Login porównujemy bez względu na wielkość liter.
func NormalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

// Sprawdza format adresu e-mail.
func ValidateEmail(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || !strings.Contains(email, ".") {
		return invalid("email", "niepoprawny adres e-mail")
	}
	return nil
}

// Minimalna polityka haseł prototypu.
func ValidatePassword(pw string) error {
	if len(pw) < 8 {
		return invalid("password", "hasło musi mieć co najmniej 8 znaków")
	}
	return nil
}

// Tworzy konto zwykłego użytkownika z walutą bazową PLN.
func NewUser(email, passwordHash string, now time.Time) (*User, error) {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}
	return &User{
		ID:           NewID(),
		Email:        email,
		PasswordHash: passwordHash,
		BaseCurrency: PLN,
		Role:         RoleUser,
		CreatedAt:    now,
	}, nil
}

// Ustawia walutę bazową użytkownika (do konwersji kosztów subskrypcji).
func (u *User) ChangeBaseCurrency(c Currency) error {
	if !c.IsValid() {
		return invalid("currency", "nieobsługiwana waluta")
	}
	u.BaseCurrency = c
	return nil
}
