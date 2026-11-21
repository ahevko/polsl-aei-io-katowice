package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

// Serwisy warstwy logiki wstrzykiwane do warstwy prezentacji.
type Dependencies struct {
	Auth          *service.AuthService
	Subscriptions *service.SubscriptionService
	Categories    *service.CategoryService
	Logger        *log.Logger
}

// Konfiguracja routera.
type Options struct {
	StaticDir  string // zbudowany frontend (frontend/dist); "" = brak
	CORSOrigin string // np. http://localhost:5173 dla `npm run dev`
}

// FASADA warstwy prezentacji: rejestruje wszystkie trasy REST API
// i zwraca jeden http.Handler opakowany łańcuchem middleware.
func NewRouter(d Dependencies, opt Options) http.Handler {
	authMW := NewAuthMiddleware(d.Auth)
	auth := &AuthHandler{auth: d.Auth}
	subs := &SubscriptionHandler{subs: d.Subscriptions}
	cats := &CategoryHandler{categories: d.Categories}
	authReqUser := authMW.RequireUser

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Zostawione z pierwszej wersji dla testów. TODO: usunąć w przyszłości.
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Pong"})
	})

	// Konto
	mux.HandleFunc("POST /api/auth/register", auth.Register)
	mux.HandleFunc("POST /api/auth/login", auth.Login)
	mux.HandleFunc("POST /api/auth/logout", authReqUser(auth.Logout))
	mux.HandleFunc("GET /api/me", authReqUser(auth.Me))
	mux.HandleFunc("PUT /api/me/currency", authReqUser(auth.SetCurrency))

	// Subskrypcje
	mux.HandleFunc("GET /api/categories", authReqUser(cats.List))
	mux.HandleFunc("GET /api/subscriptions", authReqUser(subs.List))
	mux.HandleFunc("POST /api/subscriptions", authReqUser(subs.Create))
	mux.HandleFunc("GET /api/subscriptions/{id}", authReqUser(subs.Get))
	mux.HandleFunc("PUT /api/subscriptions/{id}", authReqUser(subs.Update))
	mux.HandleFunc("DELETE /api/subscriptions/{id}", authReqUser(subs.Delete))
	mux.HandleFunc("POST /api/subscriptions/{id}/pause", authReqUser(subs.Pause))
	mux.HandleFunc("POST /api/subscriptions/{id}/resume", authReqUser(subs.Resume))

	if opt.StaticDir != "" {
		mux.Handle("/", spaHandler(opt.StaticDir))
	}

	return Chain(mux, Recover(d.Logger), Logging(d.Logger), CORS(opt.CORSOrigin))
}

// Serwuje pliki frontendu; nieznane ścieżki -> index.html (SPA).
func spaHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(p); err != nil {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}
