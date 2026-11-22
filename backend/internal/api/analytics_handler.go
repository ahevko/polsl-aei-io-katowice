package api

import (
	"net/http"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

// Udostępnia słownik kategorii.
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

// AnalyticsHandler – kontroler UC analizy wydatków.
type AnalyticsHandler struct {
	analytics *service.AnalyticsService
}

func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	s, err := h.analytics.Summary(r.Context(), currentUser(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSummaryDTO(s))
}
