package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Mapuje błędy dziedzinowe na kody HTTP.
func writeError(w http.ResponseWriter, err error) {
	var ve *domain.ValidationError
	switch {
	case errors.As(err, &ve):
		writeJSON(w, http.StatusBadRequest, errorDTO{Error: ve.Message, Field: ve.Field})
	case errors.Is(err, domain.ErrValidation):
		writeJSON(w, http.StatusBadRequest, errorDTO{Error: err.Error()})
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorDTO{Error: "nie znaleziono"})
	case errors.Is(err, domain.ErrConflict):
		writeJSON(w, http.StatusConflict, errorDTO{Error: "konto z tym adresem e-mail już istnieje"})
	case errors.Is(err, domain.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, errorDTO{Error: "niepoprawny e-mail lub hasło albo sesja wygasła"})
	case errors.Is(err, domain.ErrInvalidState):
		writeJSON(w, http.StatusConflict, errorDTO{Error: "operacja niedozwolona w obecnym statusie subskrypcji"})
	default:
		writeJSON(w, http.StatusInternalServerError, errorDTO{Error: "błąd wewnętrzny serwera"})
	}
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return &domain.ValidationError{Field: "body", Message: "niepoprawny format JSON"}
	}
	return nil
}
