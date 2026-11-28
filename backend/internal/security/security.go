// Package security – haszowanie haseł i sesje użytkowników.
package security

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PasswordHasher – kontrakt jednokierunkowego haszowania haseł (wymaganie
// pozafunkcjonalne: bezpieczeństwo).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encoded string) bool
}

// PBKDF2Hasher – PBKDF2-HMAC-SHA256 z unikatową solą (biblioteka standardowa Go).
// Format: pbkdf2-sha256$<iteracje>$<sól b64>$<skrót b64>
type PBKDF2Hasher struct {
	Iterations int
	SaltLen    int
	KeyLen     int
}

// NewPBKDF2Hasher – parametry zgodne z zaleceniami OWASP dla SHA-256.
func NewPBKDF2Hasher() *PBKDF2Hasher {
	return &PBKDF2Hasher{Iterations: 600_000, SaltLen: 16, KeyLen: 32}
}

func (h *PBKDF2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, h.Iterations, h.KeyLen)
	if err != nil {
		return "", err
	}
	enc := base64.RawStdEncoding
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", h.Iterations, enc.EncodeToString(salt), enc.EncodeToString(key)), nil
}

func (h *PBKDF2Hasher) Verify(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	enc := base64.RawStdEncoding
	salt, err1 := enc.DecodeString(parts[2])
	want, err2 := enc.DecodeString(parts[3])
	if err1 != nil || err2 != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, iter, len(want))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}

// SessionStore – sesje logowania (token okaziciela → identyfikator użytkownika),
// przechowywane w pamięci z czasem wygaśnięcia.
type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]session
	ttl      time.Duration
}

type session struct {
	userID  string
	expires time.Time
}

// NewSessionStore tworzy magazyn sesji o zadanym czasie życia.
func NewSessionStore(ttl time.Duration) *SessionStore {
	return &SessionStore{sessions: map[string]session{}, ttl: ttl}
}

// Issue tworzy nową sesję i zwraca losowy token.
func (s *SessionStore) Issue(userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	s.mu.Lock()
	s.sessions[token] = session{userID: userID, expires: time.Now().Add(s.ttl)}
	s.mu.Unlock()
	return token, nil
}

// Resolve zwraca użytkownika przypisanego do ważnego tokenu.
func (s *SessionStore) Resolve(token string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[token]
	if !ok {
		return "", false
	}
	if time.Now().After(sess.expires) {
		delete(s.sessions, token)
		return "", false
	}
	return sess.userID, true
}

// Revoke unieważnia sesję (wylogowanie).
func (s *SessionStore) Revoke(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}
