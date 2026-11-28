package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/notify"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

// Dependencies – serwisy warstwy logiki wstrzykiwane do warstwy prezentacji.
type Dependencies struct {
	Auth          *service.AuthService
	Subscriptions *service.SubscriptionService
	Categories    *service.CategoryService
	Analytics     *service.AnalyticsService
	History       *service.PaymentHistoryService
	Worker        *service.BillingWorker
	Email         *notify.EmailSimulator
	Logger        *log.Logger
}

// Options – konfiguracja routera.
type Options struct {
	StaticDir  string // zbudowany frontend (frontend/dist); "" = brak
	CORSOrigin string // np. http://localhost:5173 dla `npm run dev`
	EnableDev  bool   // endpointy demonstracyjne /api/dev/*
}

// NewRouter – FASADA warstwy prezentacji: rejestruje wszystkie trasy REST API
// i zwraca jeden http.Handler opakowany łańcuchem middleware.
func NewRouter(d Dependencies, opt Options) http.Handler {
	authMW := NewAuthMiddleware(d.Auth)
	auth := &AuthHandler{auth: d.Auth}
	subs := &SubscriptionHandler{subs: d.Subscriptions}
	cats := &CategoryHandler{categories: d.Categories}
	an := &AnalyticsHandler{analytics: d.Analytics, history: d.History}
	u := authMW.RequireUser

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Konto (diagram „Uwierzytelnianie i konto użytkownika”)
	mux.HandleFunc("POST /api/auth/register", auth.Register)
	mux.HandleFunc("POST /api/auth/login", auth.Login)
	mux.HandleFunc("POST /api/auth/logout", u(auth.Logout))
	mux.HandleFunc("GET /api/me", u(auth.Me))
	mux.HandleFunc("PUT /api/me/currency", u(auth.SetCurrency))

	// Subskrypcje (diagram „Zarządzanie subskrypcjami”)
	mux.HandleFunc("GET /api/categories", u(cats.List))
	mux.HandleFunc("GET /api/subscriptions", u(subs.List))
	mux.HandleFunc("POST /api/subscriptions", u(subs.Create))
	mux.HandleFunc("GET /api/subscriptions/{id}", u(subs.Get))
	mux.HandleFunc("PUT /api/subscriptions/{id}", u(subs.Update))
	mux.HandleFunc("DELETE /api/subscriptions/{id}", u(subs.Delete))
	mux.HandleFunc("POST /api/subscriptions/{id}/pause", u(subs.Pause))
	mux.HandleFunc("POST /api/subscriptions/{id}/resume", u(subs.Resume))

	// Analiza (diagram „Analiza wydatków i historia płatności”)
	mux.HandleFunc("GET /api/analytics/summary", u(an.Summary))
	mux.HandleFunc("GET /api/payments", u(an.Payments))
	mux.HandleFunc("GET /api/payments/export.csv", u(an.ExportCSV))

	if opt.EnableDev {
		dev := &DevHandler{worker: d.Worker, email: d.Email}
		mux.HandleFunc("POST /api/dev/worker/run", u(dev.RunWorker))
		mux.HandleFunc("GET /api/dev/outbox", u(dev.Outbox))
	}

	if opt.StaticDir != "" {
		mux.Handle("/", spaHandler(opt.StaticDir))
	}

	return Chain(mux, Recover(d.Logger), Logging(d.Logger), CORS(opt.CORSOrigin))
}

// spaHandler serwuje pliki frontendu; nieznane ścieżki → index.html (SPA).
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
