package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/currency"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/notify"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository/memory"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/security"
)

// fixture składa system tak jak main.go, ale z zegarem o stałej dacie.
type fixture struct {
	auth      *AuthService
	subs      *SubscriptionService
	analytics *AnalyticsService
	history   *PaymentHistoryService
	worker    *BillingWorker
	email     *notify.EmailSimulator
	api       *currency.ExchangeAPISimulator
	user      *domain.User
}

var today = time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)

func newFixture(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	store, _ := memory.NewStore("")
	users := memory.NewUserRepository(store)
	subs := memory.NewSubscriptionRepository(store)
	cats := memory.NewCategoryRepository(store)
	pays := memory.NewPaymentRepository(store)
	notifs := memory.NewNotificationRepository(store)
	clock := FixedClock{T: today}
	logger := log.New(io.Discard, "", 0)
	apiSim := currency.NewExchangeAPISimulator(0)
	conv := NewCurrencyConverter(currency.NewCachedRateProvider(apiSim, time.Hour))
	email := notify.NewEmailSimulator(logger, 50)
	hasher := &security.PBKDF2Hasher{Iterations: 1000, SaltLen: 16, KeyLen: 32}

	f := &fixture{
		auth:      NewAuthService(users, hasher, security.NewSessionStore(time.Hour), clock),
		subs:      NewSubscriptionService(subs, cats, pays, clock),
		analytics: NewAnalyticsService(subs, users, cats, conv),
		history:   NewPaymentHistoryService(pays, users, conv),
		email:     email,
		api:       apiSim,
	}
	f.worker = NewBillingWorker(subs, pays, NewNotificationService(subs, users, notifs, email, nil, clock), clock, logger)
	if err := NewCategoryService(cats).EnsureDefaults(ctx); err != nil {
		t.Fatal(err)
	}
	u, err := f.auth.Register(ctx, "Jan@Example.com", "tajnehaslo")
	if err != nil {
		t.Fatal(err)
	}
	f.user = u
	return f
}

func (f *fixture) add(t *testing.T, name string, amount float64, cur, cycle string, next time.Time) *domain.Subscription {
	t.Helper()
	s, err := f.subs.Create(context.Background(), f.user.ID, SubscriptionInput{
		Name: name, Amount: amount, Currency: cur, Cycle: cycle, NextPaymentDate: next, CategoryIDs: []string{"streaming"}})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAuth(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if _, err := f.auth.Register(ctx, "jan@example.com", "innehaslo"); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("duplikat e-mail: %v", err)
	}
	if _, _, err := f.auth.Login(ctx, "jan@example.com", "zlehaslo"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("złe hasło: %v", err)
	}
	token, _, err := f.auth.Login(ctx, "JAN@example.com", "tajnehaslo")
	if err != nil {
		t.Fatal(err)
	}
	if u, err := f.auth.Authenticate(ctx, token); err != nil || u.ID != f.user.ID {
		t.Errorf("authenticate: %v", err)
	}
	f.auth.Logout(token)
	if _, err := f.auth.Authenticate(ctx, token); err == nil {
		t.Error("token po wylogowaniu powinien być nieważny")
	}
}

func TestOwnershipIsolation(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	s := f.add(t, "Netflix", 49, "PLN", "MONTHLY", today.AddDate(0, 0, 5))
	other, _ := f.auth.Register(ctx, "obcy@example.com", "tajnehaslo")
	if _, err := f.subs.Get(ctx, other.ID, s.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("obcy użytkownik nie może widzieć subskrypcji: %v", err)
	}
	if err := f.subs.Delete(ctx, other.ID, s.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("obcy użytkownik nie może usuwać: %v", err)
	}
}

func TestCreateRollsPastDateForward(t *testing.T) {
	f := newFixture(t)
	s := f.add(t, "Spotify", 20, "PLN", "MONTHLY", time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC))
	if got := s.NextPaymentDate.Format("2006-01-02"); got != "2026-11-20" {
		t.Errorf("data z przeszłości powinna przejść na 2026-11-20, got %s", got)
	}
}

func TestSummaryWithCurrencyConversionAndCache(t *testing.T) {
	f := newFixture(t)
	f.add(t, "A", 100, "PLN", "MONTHLY", today.AddDate(0, 0, 5))
	f.add(t, "B", 10, "EUR", "MONTHLY", today.AddDate(0, 0, 5)) // 42.50 PLN
	f.add(t, "C", 120, "USD", "YEARLY", today.AddDate(0, 1, 0)) // 10 USD/mies. = 37.00 PLN
	paused := f.add(t, "D", 999, "PLN", "MONTHLY", today.AddDate(0, 0, 5))
	_, _ = f.subs.Pause(context.Background(), f.user.ID, paused.ID)

	sum, err := f.analytics.Summary(context.Background(), f.user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if sum.MonthlyTotal.Minor != 17950 || sum.BaseCurrency != domain.PLN {
		t.Errorf("monthly total = %v", sum.MonthlyTotal)
	}
	if sum.ActiveSubscriptions != 3 || sum.PausedSubscriptions != 1 {
		t.Errorf("liczniki: %+v", sum)
	}
	_, _ = f.analytics.Summary(context.Background(), f.user.ID)
	if f.api.Calls() != 1 {
		t.Errorf("cache: oczekiwano 1 wywołania API, było %d", f.api.Calls())
	}
}

func TestWorkerSettlesPaymentsAndSendsReminders(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.add(t, "Netflix", 49, "PLN", "MONTHLY", today.AddDate(0, 0, 3)) // 2026-11-20
	weekly := f.add(t, "Duolingo", 5, "EUR", "WEEKLY", today)         // płatność dziś

	rep, err := f.worker.RunOnce(ctx, today)
	if err != nil {
		t.Fatal(err)
	}
	if rep.PaymentsRecorded != 1 || rep.RemindersSent != 1 {
		t.Errorf("raport: %+v", rep)
	}
	// Ponowne uruchomienie tego samego dnia nie duplikuje niczego (idempotencja).
	rep, _ = f.worker.RunOnce(ctx, today)
	if rep.PaymentsRecorded != 0 || rep.RemindersSent != 0 {
		t.Errorf("drugi przebieg: %+v", rep)
	}
	got, _ := f.subs.Get(ctx, f.user.ID, weekly.ID)
	if got.NextPaymentDate.Format("2006-01-02") != "2026-11-24" {
		t.Errorf("następna płatność weekly = %s", got.NextPaymentDate)
	}
	// Worker nie działał przez 3 tygodnie – nadrabia zaległe cykle.
	rep, _ = f.worker.RunOnce(ctx, today.AddDate(0, 0, 21))
	if rep.PaymentsRecorded != 4 { // 3x Duolingo + 1x Netflix
		t.Errorf("nadrabianie: %+v", rep)
	}
	if !strings.Contains(f.email.Outbox()[len(f.email.Outbox())-1].Subject, "Netflix") {
		t.Errorf("pierwsze przypomnienie powinno dotyczyć Netflix: %+v", f.email.Outbox())
	}
}

func TestPausedSubscriptionIsSkippedByWorker(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	s := f.add(t, "Netflix", 49, "PLN", "MONTHLY", today.AddDate(0, 0, 1))
	_, _ = f.subs.Pause(ctx, f.user.ID, s.ID)
	rep, _ := f.worker.RunOnce(ctx, today.AddDate(0, 0, 1))
	if rep.PaymentsRecorded != 0 || rep.RemindersSent != 0 {
		t.Errorf("wstrzymana subskrypcja: %+v", rep)
	}
}

func TestDeleteRemovesHistoryAndCSVExport(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	s := f.add(t, "Netflix", 49, "PLN", "MONTHLY", today)
	f.add(t, "HBO Max", 29.99, "PLN", "MONTHLY", today)
	_, _ = f.worker.RunOnce(ctx, today)

	var buf bytes.Buffer
	if err := f.history.ExportCSV(ctx, f.user.ID, &buf); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 || !strings.Contains(buf.String(), "HBO Max;29.99;PLN") {
		t.Errorf("CSV:\n%s", buf.String())
	}
	if err := f.subs.Delete(ctx, f.user.ID, s.ID); err != nil {
		t.Fatal(err)
	}
	entries, _, _ := f.history.History(ctx, f.user.ID)
	if len(entries) != 1 {
		t.Errorf("po usunięciu subskrypcji historia powinna mieć 1 wpis, ma %d", len(entries))
	}
}
