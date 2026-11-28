package domain

import (
	"fmt"
	"time"
)

// Notification – zapis wysłanego przypomnienia (zapobiega duplikatom, WF-02).
type Notification struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId"`
	SubscriptionID string    `json:"subscriptionId"`
	PaymentDate    time.Time `json:"paymentDate"`
	DaysBefore     int       `json:"daysBefore"`
	Channel        string    `json:"channel"`
	SentAt         time.Time `json:"sentAt"`
}

// Key – klucz unikalności przypomnienia.
func (n Notification) Key() string {
	return NotificationKey(n.SubscriptionID, n.PaymentDate, n.DaysBefore)
}

// NotificationKey buduje klucz unikalności bez tworzenia encji.
func NotificationKey(subID string, paymentDate time.Time, daysBefore int) string {
	return fmt.Sprintf("%s|%s|%d", subID, paymentDate.Format("2006-01-02"), daysBefore)
}
