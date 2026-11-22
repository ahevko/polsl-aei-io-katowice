package service

import (
	"context"
	"sort"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/repository"
)

// CategoryCost – koszt miesięczny przypadający na kategorię.
type CategoryCost struct {
	CategoryID   string
	CategoryName string
	Monthly      domain.Money
}

// CostSummary – wynik UC „Przeglądaj podsumowanie kosztów”.
type CostSummary struct {
	BaseCurrency        domain.Currency
	MonthlyTotal        domain.Money
	YearlyTotal         domain.Money
	ActiveSubscriptions int
	PausedSubscriptions int
	ByCategory          []CategoryCost
}

// AnalyticsService – UC z diagramu „Analiza wydatków” (WF-03).
type AnalyticsService struct {
	subs       repository.SubscriptionRepository
	users      repository.UserRepository
	categories repository.CategoryRepository
	converter  *CurrencyConverter
}

func NewAnalyticsService(subs repository.SubscriptionRepository, users repository.UserRepository, categories repository.CategoryRepository, converter *CurrencyConverter) *AnalyticsService {
	return &AnalyticsService{subs: subs, users: users, categories: categories, converter: converter}
}

// Summary – UC „Wylicz koszty miesięczne i roczne” + „Grupuj wydatki
// według kategorii”. Subskrypcja w kilku kategoriach dzieli koszt po równo,
// żeby suma po kategoriach równała się sumie całkowitej.
func (s *AnalyticsService) Summary(ctx context.Context, userID string) (*CostSummary, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	subs, err := s.subs.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	cats, err := s.categories.List(ctx)
	if err != nil {
		return nil, err
	}
	names := map[string]string{}
	for _, c := range cats {
		names[c.ID] = c.Name
	}
	base := u.BaseCurrency
	res := &CostSummary{
		BaseCurrency: base,
		MonthlyTotal: domain.Money{Currency: base},
		YearlyTotal:  domain.Money{Currency: base},
	}
	perCat := map[string]int64{}
	for _, sub := range subs {
		if !sub.IsActive() {
			res.PausedSubscriptions++
			continue
		}
		res.ActiveSubscriptions++
		monthly, err := sub.MonthlyCost()
		if err != nil {
			return nil, err
		}
		yearly, err := sub.YearlyCost()
		if err != nil {
			return nil, err
		}
		if monthly, err = s.converter.Convert(ctx, monthly, base); err != nil {
			return nil, err
		}
		if yearly, err = s.converter.Convert(ctx, yearly, base); err != nil {
			return nil, err
		}
		res.MonthlyTotal.Minor += monthly.Minor
		res.YearlyTotal.Minor += yearly.Minor
		n := int64(len(sub.CategoryIDs))
		for _, cid := range sub.CategoryIDs {
			perCat[cid] += monthly.Minor / n
		}
	}
	for cid, minor := range perCat {
		name := names[cid]
		if name == "" {
			name = cid
		}
		res.ByCategory = append(res.ByCategory, CategoryCost{CategoryID: cid, CategoryName: name, Monthly: domain.Money{Minor: minor, Currency: base}})
	}
	sort.Slice(res.ByCategory, func(i, j int) bool { return res.ByCategory[i].Monthly.Minor > res.ByCategory[j].Monthly.Minor })
	return res, nil
}
