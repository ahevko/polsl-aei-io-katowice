// Punkt wejścia serwera – KORZEŃ KOMPOZYCJI (Composition Root):
// tu i tylko tu tworzone są konkretne implementacje i wstrzykiwane do
// kolejnych warstw (Dependency Injection przez konstruktory).
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/api"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository/memory"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/security"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

func main() {
	addr := flag.String("addr", ":8080", "adres nasłuchiwania HTTP")
	dataFile := flag.String("data", "data/data.json", "plik danych (pusty = tylko pamięć)")
	staticDir := flag.String("static", "../frontend/dist", "katalog zbudowanego frontendu")
	cors := flag.String("cors", "http://localhost:5173", "dozwolony Origin (serwer Vite)")
	seed := flag.Bool("seed", true, "utwórz konto demo@example.com / demo1234")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- Warstwa dostępu do danych i integracji ---
	store, err := memory.NewStore(*dataFile)
	if err != nil {
		logger.Fatal(err)
	}
	users := memory.NewUserRepository(store)
	subsRepo := memory.NewSubscriptionRepository(store)
	catRepo := memory.NewCategoryRepository(store)
	payRepo := memory.NewPaymentRepository(store)

	// --- Warstwa logiki biznesowej ---
	clock := service.SystemClock{}
	authSvc := service.NewAuthService(users, security.NewPBKDF2Hasher(), security.NewSessionStore(24*time.Hour), clock)
	subSvc := service.NewSubscriptionService(subsRepo, catRepo, payRepo, clock)
	catSvc := service.NewCategoryService(catRepo)

	if err := catSvc.EnsureDefaults(ctx); err != nil {
		logger.Fatal(err)
	}
	if *seed {
		seedDemo(ctx, logger, authSvc, subSvc)
	}

	// --- Warstwa prezentacji ---
	staticPath := *staticDir
	if _, err := os.Stat(staticPath); err != nil {
		staticPath = ""
	}
	handler := api.NewRouter(api.Dependencies{
		Auth: authSvc, Subscriptions: subSvc, Categories: catSvc, Logger: logger,
	}, api.Options{StaticDir: staticPath, CORSOrigin: *cors})

	srv := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	logger.Printf("Serwer działa na http://localhost%s (frontend: %q)", *addr, staticPath)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal(err)
	}
	logger.Println("Zatrzymano serwer.")
}

// seedDemo tworzy konto demonstracyjne z kilkoma subskrypcjami.
func seedDemo(ctx context.Context, logger *log.Logger, auth *service.AuthService, subs *service.SubscriptionService) {
	u, err := auth.Register(ctx, "demo@example.com", "demo1234")
	if errors.Is(err, domain.ErrConflict) {
		return // konto już istnieje (dane z pliku)
	}
	if err != nil {
		logger.Printf("seed: %v", err)
		return
	}
	today := domain.DateOnly(time.Now())
	samples := []service.SubscriptionInput{
		{Name: "Netflix", Amount: 49, Currency: "PLN", Cycle: "MONTHLY", NextPaymentDate: today.AddDate(0, 0, 3), CategoryIDs: []string{"streaming"}},
		{Name: "Spotify", Amount: 23.99, Currency: "PLN", Cycle: "MONTHLY", NextPaymentDate: today.AddDate(0, 0, 10), CategoryIDs: []string{"streaming"}},
		{Name: "JetBrains All Products", Amount: 289, Currency: "EUR", Cycle: "YEARLY", NextPaymentDate: today.AddDate(0, 4, 0), CategoryIDs: []string{"software", "education"}},
		{Name: "iCloud+ 200 GB", Amount: 2.99, Currency: "USD", Cycle: "MONTHLY", NextPaymentDate: today.AddDate(0, 0, 1), CategoryIDs: []string{"cloud"}},
		{Name: "Duolingo", Amount: 5.5, Currency: "EUR", Cycle: "WEEKLY", NextPaymentDate: today.AddDate(0, 0, 5), CategoryIDs: []string{"education"}},
	}
	for _, in := range samples {
		if _, err := subs.Create(ctx, u.ID, in); err != nil {
			logger.Printf("seed %s: %v", in.Name, err)
		}
	}
	logger.Println("Utworzono konto demo: demo@example.com / demo1234")
}
