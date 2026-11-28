package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository"
)

// PaymentEntry – wpis historii wraz z kwotą przeliczoną na walutę bazową.
type PaymentEntry struct {
	Payment   *domain.Payment
	Converted domain.Money
}

// PaymentHistoryService – UC „Przeglądaj historię płatności”, „Eksportuj dane do CSV” (WF-06).
type PaymentHistoryService struct {
	payments  repository.PaymentRepository
	users     repository.UserRepository
	converter *CurrencyConverter
}

func NewPaymentHistoryService(payments repository.PaymentRepository, users repository.UserRepository, converter *CurrencyConverter) *PaymentHistoryService {
	return &PaymentHistoryService{payments: payments, users: users, converter: converter}
}

// History zwraca historię płatności (najnowsze na początku).
func (s *PaymentHistoryService) History(ctx context.Context, userID string) ([]PaymentEntry, domain.Currency, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	ps, err := s.payments.ListByUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	out := make([]PaymentEntry, 0, len(ps))
	for _, p := range ps {
		conv, err := s.converter.Convert(ctx, p.Amount, u.BaseCurrency)
		if err != nil {
			return nil, "", err
		}
		out = append(out, PaymentEntry{Payment: p, Converted: conv})
	}
	return out, u.BaseCurrency, nil
}

// ExportCSV zapisuje historię w formacie CSV (separator ';' i BOM UTF-8,
// aby plik poprawnie otwierał się w polskim Excelu).
func (s *PaymentHistoryService) ExportCSV(ctx context.Context, userID string, w io.Writer) error {
	entries, base, err := s.History(ctx, userID)
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	_ = cw.Write([]string{"data", "subskrypcja", "kwota", "waluta", "kwota_w_walucie_bazowej", "waluta_bazowa"})
	for _, e := range entries {
		_ = cw.Write([]string{
			e.Payment.Date.Format("2006-01-02"),
			e.Payment.SubscriptionName,
			fmt.Sprintf("%.2f", e.Payment.Amount.Amount()),
			string(e.Payment.Amount.Currency),
			fmt.Sprintf("%.2f", e.Converted.Amount()),
			string(base),
		})
	}
	cw.Flush()
	return cw.Error()
}
