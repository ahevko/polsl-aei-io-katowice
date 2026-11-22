package currency

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

// Tabela kursów: ile PLN kosztuje 1 jednostka danej waluty.
type RateTable struct {
	Rates     map[domain.Currency]float64 `json:"rates"`
	FetchedAt time.Time                   `json:"fetchedAt"`
	Source    string                      `json:"source"`
}

// Zwraca kurs waluty względem PLN.
func (t RateTable) PLNPer(c domain.Currency) (float64, error) {
	r, ok := t.Rates[c]
	if !ok || r <= 0 {
		return 0, fmt.Errorf("brak kursu dla %s", c)
	}
	return r, nil
}

// Kontrakt dostawcy kursów walut.
type RateProvider interface {
	CurrentRates(ctx context.Context) (RateTable, error)
}

// SYMULATOR zewnętrznego API kursów walut.
// Symuluje opóźnienie sieci i zlicza wywołania, aby było widać działanie cache.
type ExchangeAPISimulator struct {
	reference map[domain.Currency]float64
	latency   time.Duration
	calls     atomic.Int64
}

// Tworzy symulator z kursami referencyjnymi.
func NewExchangeAPISimulator(latency time.Duration) *ExchangeAPISimulator {
	return &ExchangeAPISimulator{
		reference: map[domain.Currency]float64{domain.PLN: 1.0, domain.EUR: 4.25, domain.USD: 3.70},
		latency:   latency,
	}
}

func (s *ExchangeAPISimulator) CurrentRates(ctx context.Context) (RateTable, error) {
	n := s.calls.Add(1)
	select {
	case <-time.After(s.latency):
	case <-ctx.Done():
		return RateTable{}, ctx.Err()
	}
	log.Printf("[kursy] wywołanie symulatora zewnętrznego API (#%d)", n)
	rates := make(map[domain.Currency]float64, len(s.reference))
	for k, v := range s.reference {
		rates[k] = v
	}
	return RateTable{Rates: rates, FetchedAt: time.Now(), Source: "symulator-api"}, nil
}

// Liczba wywołań "zewnętrznego" API (do demonstracji i testów).
func (s *ExchangeAPISimulator) Calls() int64 { return s.calls.Load() }

// Wzorzec PEŁNOMOCNIK (Proxy) / DEKORATOR:
// ten sam interfejs co RateProvider, ale wynik jest buforowany przez ttl,
// co ogranicza liczbę zapytań do zewnętrznego API.
type CachedRateProvider struct {
	inner   RateProvider
	ttl     time.Duration
	now     func() time.Time
	mu      sync.Mutex
	cached  *RateTable
	expires time.Time
}

// Opakowuje dowolnego dostawcę kursów.
func NewCachedRateProvider(inner RateProvider, ttl time.Duration) *CachedRateProvider {
	return &CachedRateProvider{inner: inner, ttl: ttl, now: time.Now}
}

func (p *CachedRateProvider) CurrentRates(ctx context.Context) (RateTable, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cached != nil && p.now().Before(p.expires) {
		return *p.cached, nil // trafienie w cache
	}
	t, err := p.inner.CurrentRates(ctx)
	if err != nil {
		if p.cached != nil { // degradacja: lepiej stare kursy niż brak
			return *p.cached, nil
		}
		return RateTable{}, err
	}
	p.cached = &t
	p.expires = p.now().Add(p.ttl)
	return t, nil
}

// Czyści bufor (np. po aktualizacji kursów przez administratora).
func (p *CachedRateProvider) Invalidate() {
	p.mu.Lock()
	p.cached = nil
	p.mu.Unlock()
}
