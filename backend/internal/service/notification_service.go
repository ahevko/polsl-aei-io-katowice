package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/notify"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository"
)

// Wysyła przypomnienia o płatnościach.
type NotificationService struct {
	subs          repository.SubscriptionRepository
	users         repository.UserRepository
	notifications repository.NotificationRepository
	notifier      notify.Notifier
	reminderDays  []int
	clock         Clock
}

// Tworzy serwis powiadomień z listą dni przypomnień (np. 3 i 1).
func NewNotificationService(subs repository.SubscriptionRepository, users repository.UserRepository, notifications repository.NotificationRepository, notifier notify.Notifier, reminderDays []int, clock Clock) *NotificationService {
	if len(reminderDays) == 0 {
		reminderDays = []int{3, 1}
	}
	return &NotificationService{subs: subs, users: users, notifications: notifications, notifier: notifier, reminderDays: reminderDays, clock: clock}
}

// Wysyła przypomnienia dla płatności przypadających za N dni.
// Wstrzymane subskrypcje są pomijane; duplikaty blokuje rejestr powiadomień.
func (s *NotificationService) SendReminders(ctx context.Context, today time.Time) (int, error) {
	today = domain.DateOnly(today)
	subs, err := s.subs.ListActive(ctx)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, sub := range subs {
		for _, days := range s.reminderDays {
			if !sub.NextPaymentDate.Equal(today.AddDate(0, 0, days)) {
				continue
			}
			key := domain.NotificationKey(sub.ID, sub.NextPaymentDate, days)
			if done, err := s.notifications.Exists(ctx, key); err != nil || done {
				continue
			}
			u, err := s.users.GetByID(ctx, sub.UserID)
			if err != nil {
				continue
			}
			msg := notify.Message{
				To:      u.Email,
				Subject: fmt.Sprintf("Przypomnienie: %s - płatność za %d %s", sub.Name, days, dniLabel(days)),
				Body: fmt.Sprintf("Dnia %s zostanie pobrana opłata %s za subskrypcję „%s”.",
					sub.NextPaymentDate.Format("02.01.2006"), sub.Price, sub.Name),
			}
			if err := s.notifier.Send(ctx, msg); err != nil {
				return sent, err
			}
			n := &domain.Notification{ID: domain.NewID(), UserID: u.ID, SubscriptionID: sub.ID,
				PaymentDate: sub.NextPaymentDate, DaysBefore: days, Channel: "EMAIL", SentAt: s.clock.Now()}
			if err := s.notifications.Add(ctx, n); err != nil {
				return sent, err
			}
			sent++
		}
	}
	return sent, nil
}

// Zwraca etykietę liczby dni w języku polskim (1 dzień, 2 dni, 5 dni).
func dniLabel(n int) string {
	if n == 1 {
		return "dzień"
	}
	return "dni"
}
