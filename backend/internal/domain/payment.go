package domain

import (
	"time"
)

// Wpis w historii płatności, tworzony przez Workera.
type Payment struct {
	ID               string    `json:"id"`
	UserID           string    `json:"userId"`
	SubscriptionID   string    `json:"subscriptionId"`
	SubscriptionName string    `json:"subscriptionName"`
	Date             time.Time `json:"date"`
	Amount           Money     `json:"amount"`
	CreatedAt        time.Time `json:"createdAt"`
}

// Rejestruje płatność za subskrypcję w dniu jej terminu.
func NewPayment(s *Subscription, now time.Time) *Payment {
	return &Payment{
		ID:               NewID(),
		UserID:           s.UserID,
		SubscriptionID:   s.ID,
		SubscriptionName: s.Name,
		Date:             s.NextPaymentDate,
		Amount:           s.Price,
		CreatedAt:        now,
	}
}
