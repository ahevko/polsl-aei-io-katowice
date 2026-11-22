package domain

import (
	"fmt"
	"time"
)

// Zapis wysłanego przypomnienia (zapobiega duplikatom).
type Notification struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId"`
	SubscriptionID string    `json:"subscriptionId"`
	PaymentDate    time.Time `json:"paymentDate"`
	DaysBefore     int       `json:"daysBefore"`
	Channel        string    `json:"channel"`
	SentAt         time.Time `json:"sentAt"`
}

// Klucz unikalności przypomnienia.
func (n Notification) Key() string {
	return NotificationKey(n.SubscriptionID, n.PaymentDate, n.DaysBefore)
}

// Buduje klucz unikalności bez tworzenia encji.
func NotificationKey(subID string, paymentDate time.Time, daysBefore int) string {
	return fmt.Sprintf("%s|%s|%d", subID, paymentDate.Format("2006-01-02"), daysBefore)
}
