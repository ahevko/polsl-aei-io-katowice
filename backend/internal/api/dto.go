// Package api – warstwa prezentacji po stronie serwera: REST API (JSON)
// wywoływane przez aplikację SPA (React). Odpowiada za mapowanie HTTP ↔
// serwisy oraz za konwersję modelu domenowego na obiekty DTO.
package api

import (
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

const dateLayout = "2006-01-02"

// --- żądania (Data Transfer Objects) ---

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type currencyRequest struct {
	Currency string `json:"currency"`
}

type subscriptionRequest struct {
	Name            string   `json:"name"`
	Amount          float64  `json:"amount"`
	Currency        string   `json:"currency"`
	Cycle           string   `json:"cycle"`
	NextPaymentDate string   `json:"nextPaymentDate"`
	CategoryIDs     []string `json:"categoryIds"`
}

func (r subscriptionRequest) toInput() (service.SubscriptionInput, error) {
	d, err := parseDate("nextPaymentDate", r.NextPaymentDate)
	if err != nil {
		return service.SubscriptionInput{}, err
	}
	return service.SubscriptionInput{Name: r.Name, Amount: r.Amount, Currency: r.Currency,
		Cycle: r.Cycle, NextPaymentDate: d, CategoryIDs: r.CategoryIDs}, nil
}

type resumeRequest struct {
	NextPaymentDate string `json:"nextPaymentDate"`
}

func parseDate(field, s string) (time.Time, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, &domain.ValidationError{Field: field, Message: "oczekiwany format RRRR-MM-DD"}
	}
	return t, nil
}

// --- odpowiedzi ---

type moneyDTO struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func toMoneyDTO(m domain.Money) moneyDTO {
	return moneyDTO{Amount: m.Amount(), Currency: string(m.Currency)}
}

type userDTO struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	BaseCurrency string `json:"baseCurrency"`
	Role         string `json:"role"`
}

func toUserDTO(u *domain.User) userDTO {
	return userDTO{ID: u.ID, Email: u.Email, BaseCurrency: string(u.BaseCurrency), Role: string(u.Role)}
}

type loginResponse struct {
	Token string  `json:"token"`
	User  userDTO `json:"user"`
}

type subscriptionDTO struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Price           moneyDTO `json:"price"`
	MonthlyCost     moneyDTO `json:"monthlyCost"`
	Cycle           string   `json:"cycle"`
	NextPaymentDate string   `json:"nextPaymentDate"`
	CategoryIDs     []string `json:"categoryIds"`
	Status          string   `json:"status"`
}

func toSubscriptionDTO(s *domain.Subscription) subscriptionDTO {
	monthly, _ := s.MonthlyCost()
	return subscriptionDTO{
		ID: s.ID, Name: s.Name, Price: toMoneyDTO(s.Price), MonthlyCost: toMoneyDTO(monthly),
		Cycle: string(s.Cycle), NextPaymentDate: s.NextPaymentDate.Format(dateLayout),
		CategoryIDs: s.CategoryIDs, Status: string(s.Status),
	}
}

type categoryCostDTO struct {
	CategoryID   string   `json:"categoryId"`
	CategoryName string   `json:"categoryName"`
	Monthly      moneyDTO `json:"monthly"`
}

type summaryDTO struct {
	BaseCurrency        string            `json:"baseCurrency"`
	MonthlyTotal        moneyDTO          `json:"monthlyTotal"`
	YearlyTotal         moneyDTO          `json:"yearlyTotal"`
	ActiveSubscriptions int               `json:"activeSubscriptions"`
	PausedSubscriptions int               `json:"pausedSubscriptions"`
	ByCategory          []categoryCostDTO `json:"byCategory"`
}

func toSummaryDTO(s *service.CostSummary) summaryDTO {
	out := summaryDTO{BaseCurrency: string(s.BaseCurrency), MonthlyTotal: toMoneyDTO(s.MonthlyTotal),
		YearlyTotal: toMoneyDTO(s.YearlyTotal), ActiveSubscriptions: s.ActiveSubscriptions,
		PausedSubscriptions: s.PausedSubscriptions, ByCategory: []categoryCostDTO{}}
	for _, c := range s.ByCategory {
		out.ByCategory = append(out.ByCategory, categoryCostDTO{c.CategoryID, c.CategoryName, toMoneyDTO(c.Monthly)})
	}
	return out
}

type errorDTO struct {
	Error string `json:"error"`
	Field string `json:"field,omitempty"`
}
