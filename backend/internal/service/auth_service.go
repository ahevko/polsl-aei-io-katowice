// Package service – warstwa logiki biznesowej. Serwisy realizują przypadki
// użycia, korzystając z modelu domenowego i kontraktów repozytoriów.
package service

import (
	"context"
	"errors"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/security"
)

// AuthService – UC: Zarejestruj konto, Zaloguj się, Wyloguj się,
// Ustaw walutę domyślną (WF-05).
type AuthService struct {
	users    repository.UserRepository
	hasher   security.PasswordHasher
	sessions *security.SessionStore
	clock    Clock
}

// NewAuthService – wstrzykiwanie zależności przez konstruktor.
func NewAuthService(users repository.UserRepository, hasher security.PasswordHasher, sessions *security.SessionStore, clock Clock) *AuthService {
	return &AuthService{users: users, hasher: hasher, sessions: sessions, clock: clock}
}

// Register tworzy konto; zajęty e-mail jest odrzucany (ErrConflict).
func (s *AuthService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	email = domain.NormalizeEmail(email)
	if err := domain.ValidateEmail(email); err != nil { // UC „Waliduj dane konta”
		return nil, err
	}
	if err := domain.ValidatePassword(password); err != nil {
		return nil, err
	}
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, domain.ErrConflict
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	u, err := domain.NewUser(email, hash, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.users.Add(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login weryfikuje dane i wydaje token sesji.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	u, err := s.users.GetByEmail(ctx, domain.NormalizeEmail(email))
	if errors.Is(err, domain.ErrNotFound) || (err == nil && !s.hasher.Verify(password, u.PasswordHash)) {
		return "", nil, domain.ErrUnauthorized
	}
	if err != nil {
		return "", nil, err
	}
	token, err := s.sessions.Issue(u.ID)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

// Logout unieważnia token.
func (s *AuthService) Logout(token string) { s.sessions.Revoke(token) }

// Authenticate zwraca zalogowanego użytkownika na podstawie tokenu.
func (s *AuthService) Authenticate(ctx context.Context, token string) (*domain.User, error) {
	id, ok := s.sessions.Resolve(token)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	return u, nil
}

// SetBaseCurrency – UC „Ustaw walutę domyślną”.
func (s *AuthService) SetBaseCurrency(ctx context.Context, userID string, c domain.Currency) (*domain.User, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := u.ChangeBaseCurrency(c); err != nil {
		return nil, err
	}
	return u, s.users.Update(ctx, u)
}
