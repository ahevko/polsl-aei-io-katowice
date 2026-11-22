package api

import (
	"context"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

// Wzorzec ŁAŃCUCH ZOBOWIĄZAŃ / DEKORATOR: każde ogniwo
// opakowuje kolejny http.Handler i może przerwać przetwarzanie żądania.
type Middleware func(http.Handler) http.Handler

// Składa middleware w kolejności podania (pierwszy = zewnętrzny).
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}

// Zapisuje metodę, ścieżkę, status i czas obsługi żądania.
func Logging(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if strings.HasPrefix(r.URL.Path, "/api/") {
				logger.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Microsecond))
			}
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Zamienia panikę w odpowiedź 500 zamiast zatrzymania serwera.
func Recover(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					logger.Printf("panic: %v", v)
					writeJSON(w, http.StatusInternalServerError, errorDTO{Error: "błąd wewnętrzny serwera"})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS pozwala serwerowi deweloperskiemu Vite (na innym porcie) wywoływać API.
func CORS(allowedOrigin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type ctxKey int

const userKey ctxKey = 1

// Wymaga ważnego tokenu sesji (nagłówek Authorization: Bearer).
type AuthMiddleware struct {
	auth *service.AuthService
}

func NewAuthMiddleware(auth *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{auth: auth}
}

// Przepuszcza żądanie tylko dla zalogowanego użytkownika.
func (m *AuthMiddleware) RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := m.auth.Authenticate(r.Context(), bearerToken(r))
		if err != nil {
			writeError(w, domain.ErrUnauthorized)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
	}
}

// Odczytuje zalogowanego użytkownika z kontekstu żądania.
func currentUser(r *http.Request) *domain.User {
	u, _ := r.Context().Value(userKey).(*domain.User)
	return u
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if t, ok := strings.CutPrefix(h, "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return r.URL.Query().Get("token") // dla linku pobierania CSV
}
