package api

import (
	"net/http"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

// CategoryHandler – udostępnia słownik kategorii.
type CategoryHandler struct {
	categories *service.CategoryService
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cs, err := h.categories.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cs)
}

// AnalyticsHandler – kontroler UC analizy wydatków i historii płatności.
type AnalyticsHandler struct {
	analytics *service.AnalyticsService
	history   *service.PaymentHistoryService
}

func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	s, err := h.analytics.Summary(r.Context(), currentUser(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummaryDTO(s))
}

func (h *AnalyticsHandler) Payments(w http.ResponseWriter, r *http.Request) {
	entries, _, err := h.history.History(r.Context(), currentUser(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]paymentDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, toPaymentDTO(e))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *AnalyticsHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="historia-platnosci.csv"`)
	if err := h.history.ExportCSV(r.Context(), currentUser(r).ID, w); err != nil {
		writeError(w, err)
	}
}
