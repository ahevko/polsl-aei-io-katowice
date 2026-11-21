package memory

import (
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository"
)

// Kontrola na etapie kompilacji, że implementacje spełniają kontrakty interfejsów repozytoriów.
var (
	_ repository.UserRepository         = (*UserRepository)(nil)
	_ repository.SubscriptionRepository = (*SubscriptionRepository)(nil)
	_ repository.CategoryRepository     = (*CategoryRepository)(nil)
	_ repository.PaymentRepository      = (*PaymentRepository)(nil)
	_ repository.NotificationRepository = (*NotificationRepository)(nil)
)
