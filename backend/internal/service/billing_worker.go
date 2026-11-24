package service

import (
	"context"
	"log"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository"
)

// Wynik jednego przebiegu Workera.
type WorkerReport struct {
	Day              time.Time
	PaymentsRecorded int
	RemindersSent    int
}

// Worker rozlicza zaległe płatności i wysyła przypomnienia.
type BillingWorker struct {
	subs          repository.SubscriptionRepository
	payments      repository.PaymentRepository
	notifications *NotificationService
	clock         Clock
	logger        *log.Logger
}

func NewBillingWorker(subs repository.SubscriptionRepository, payments repository.PaymentRepository, notifications *NotificationService, clock Clock, logger *log.Logger) *BillingWorker {
	return &BillingWorker{subs: subs, payments: payments, notifications: notifications, clock: clock, logger: logger}
}

// Rozlicza wszystkie zaległe płatności do dnia day włącznie
// (również kilka cykli wstecz, jeśli Worker nie działał) i wysyła przypomnienia.
func (w *BillingWorker) RunOnce(ctx context.Context, day time.Time) (WorkerReport, error) {
	day = domain.DateOnly(day)
	rep := WorkerReport{Day: day}
	due, err := w.subs.ListDue(ctx, day)
	if err != nil {
		return rep, err
	}
	for _, sub := range due {
		for sub.IsDue(day) {
			if err := w.payments.Add(ctx, domain.NewPayment(sub, w.clock.Now())); err != nil {
				return rep, err
			}
			rep.PaymentsRecorded++
			if err := sub.AdvanceBillingDate(); err != nil {
				return rep, err
			}
		}
		if err := w.subs.Update(ctx, sub); err != nil {
			return rep, err
		}
	}
	if rep.RemindersSent, err = w.notifications.SendReminders(ctx, day); err != nil {
		return rep, err
	}
	w.logger.Printf("[worker] %s: zapisano płatności=%d, wysłano przypomnień=%d",
		day.Format("2006-01-02"), rep.PaymentsRecorded, rep.RemindersSent)
	return rep, nil
}

// Uruchamia Workera cyklicznie (harmonogram) do czasu anulowania ctx.
func (w *BillingWorker) Start(ctx context.Context, interval time.Duration) {
	go func() {
		run := func() {
			if _, err := w.RunOnce(ctx, w.clock.Now()); err != nil {
				w.logger.Printf("[worker] błąd: %v", err)
			}
		}
		run()
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				run()
			}
		}
	}()
}
