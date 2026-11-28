package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/currency"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/notify"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository/memory"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/security"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

// Test integracyjny: pełny przepływ przez wszystkie warstwy
// (HTTP → handler → serwis → repozytorium in-memory).
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	store, _ := memory.NewStore("")
	users := memory.NewUserRepository(store)
	subs := memory.NewSubscriptionRepository(store)
	cats := memory.NewCategoryRepository(store)
	pays := memory.NewPaymentRepository(store)
	notifs := memory.NewNotificationRepository(store)
	clock := service.SystemClock{}
	logger := log.New(io.Discard, "", 0)
	conv := service.NewCurrencyConverter(currency.NewCachedRateProvider(currency.NewExchangeAPISimulator(0), time.Hour))
	email := notify.NewEmailSimulator(logger, 50)
	catSvc := service.NewCategoryService(cats)
	_ = catSvc.EnsureDefaults(context.Background())
	notifSvc := service.NewNotificationService(subs, users, notifs, email, nil, clock)
	h := NewRouter(Dependencies{
		Auth:          service.NewAuthService(users, &security.PBKDF2Hasher{Iterations: 1000, SaltLen: 16, KeyLen: 32}, security.NewSessionStore(time.Hour), clock),
		Subscriptions: service.NewSubscriptionService(subs, cats, pays, clock),
		Categories:    catSvc,
		Analytics:     service.NewAnalyticsService(subs, users, cats, conv),
		History:       service.NewPaymentHistoryService(pays, users, conv),
		Worker:        service.NewBillingWorker(subs, pays, notifSvc, clock, logger),
		Email:         email,
		Logger:        logger,
	}, Options{EnableDev: true})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func call(t *testing.T, srv *httptest.Server, method, path, token string, body any, out any) int {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, srv.URL+path, rd)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

func TestEndToEndFlow(t *testing.T) {
	srv := newTestServer(t)
	creds := map[string]string{"email": "anna@example.com", "password": "bezpieczne1"}

	if c := call(t, srv, "POST", "/api/auth/register", "", creds, nil); c != http.StatusCreated {
		t.Fatalf("register: %d", c)
	}
	if c := call(t, srv, "POST", "/api/auth/register", "", creds, nil); c != http.StatusConflict {
		t.Errorf("duplikat: %d", c)
	}
	if c := call(t, srv, "GET", "/api/subscriptions", "", nil, nil); c != http.StatusUnauthorized {
		t.Errorf("bez tokenu: %d", c)
	}
	var login loginResponse
	if c := call(t, srv, "POST", "/api/auth/login", "", creds, &login); c != http.StatusOK || login.Token == "" {
		t.Fatalf("login: %d", c)
	}
	tok := login.Token

	next := time.Now().AddDate(0, 0, 3).Format(dateLayout)
	var created subscriptionDTO
	c := call(t, srv, "POST", "/api/subscriptions", tok, subscriptionRequest{
		Name: "Netflix", Amount: 49, Currency: "PLN", Cycle: "MONTHLY", NextPaymentDate: next,
		CategoryIDs: []string{"streaming"}}, &created)
	if c != http.StatusCreated || created.ID == "" {
		t.Fatalf("create: %d %+v", c, created)
	}
	var bad errorDTO
	if c := call(t, srv, "POST", "/api/subscriptions", tok, subscriptionRequest{
		Name: "X", Amount: 10, Currency: "GBP", Cycle: "MONTHLY", NextPaymentDate: next,
		CategoryIDs: []string{"streaming"}}, &bad); c != http.StatusBadRequest || bad.Field != "currency" {
		t.Errorf("walidacja waluty: %d %+v", c, bad)
	}

	var paused subscriptionDTO
	call(t, srv, "POST", "/api/subscriptions/"+created.ID+"/pause", tok, nil, &paused)
	if paused.Status != "PAUSED" {
		t.Errorf("pause: %+v", paused)
	}
	var resumed subscriptionDTO
	call(t, srv, "POST", "/api/subscriptions/"+created.ID+"/resume", tok, resumeRequest{NextPaymentDate: next}, &resumed)
	if resumed.Status != "ACTIVE" {
		t.Errorf("resume: %+v", resumed)
	}

	var sum summaryDTO
	call(t, srv, "GET", "/api/analytics/summary", tok, nil, &sum)
	if sum.MonthlyTotal.Amount != 49 || len(sum.ByCategory) != 1 {
		t.Errorf("summary: %+v", sum)
	}

	var rep workerReportDTO
	call(t, srv, "POST", "/api/dev/worker/run?date="+time.Now().Format(dateLayout), tok, nil, &rep)
	if rep.RemindersSent != 1 {
		t.Errorf("worker: %+v", rep)
	}
	var outbox []notify.Message
	call(t, srv, "GET", "/api/dev/outbox", tok, nil, &outbox)
	if len(outbox) != 1 {
		t.Errorf("outbox: %+v", outbox)
	}

	call(t, srv, "POST", "/api/dev/worker/run?date="+next, tok, nil, &rep) // dzień płatności
	var pays []paymentDTO
	call(t, srv, "GET", "/api/payments", tok, nil, &pays)
	if len(pays) != 1 {
		t.Fatalf("historia: %+v", pays)
	}
	if c := call(t, srv, "DELETE", "/api/subscriptions/"+created.ID, tok, nil, nil); c != http.StatusNoContent {
		t.Errorf("delete: %d", c)
	}
	var list []subscriptionDTO
	call(t, srv, "GET", "/api/subscriptions", tok, nil, &list)
	if len(list) != 0 {
		t.Errorf("lista po usunięciu: %+v", list)
	}
}
