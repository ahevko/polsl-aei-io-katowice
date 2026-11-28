package security

import (
	"testing"
	"time"
)

func TestPBKDF2Hasher(t *testing.T) {
	h := &PBKDF2Hasher{Iterations: 1000, SaltLen: 16, KeyLen: 32}
	a, _ := h.Hash("haslo1234")
	b, _ := h.Hash("haslo1234")
	if a == b {
		t.Error("ta sama wartość skrótu – sól nie działa")
	}
	if !h.Verify("haslo1234", a) || h.Verify("haslo1235", a) || h.Verify("haslo1234", "zepsuty") {
		t.Error("weryfikacja hasła działa niepoprawnie")
	}
}

func TestSessionExpiry(t *testing.T) {
	s := NewSessionStore(-time.Second) // sesja od razu wygasła
	tok, _ := s.Issue("u1")
	if _, ok := s.Resolve(tok); ok {
		t.Error("wygasła sesja nie powinna być ważna")
	}
}
