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
