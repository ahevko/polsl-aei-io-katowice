package api

import (
	"net/http"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/domain"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

// AuthHandler – kontroler UC konta użytkownika.
type AuthHandler struct {
	auth *service.AuthService
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	u, err := h.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toUserDTO(u))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	token, u, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{Token: token, User: toUserDTO(u)})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.auth.Logout(bearerToken(r))
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, toUserDTO(currentUser(r)))
}

func (h *AuthHandler) SetCurrency(w http.ResponseWriter, r *http.Request) {
	var req currencyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	c, err := domain.ParseCurrency(req.Currency)
	if err != nil {
		writeError(w, err)
		return
	}
	u, err := h.auth.SetBaseCurrency(r.Context(), currentUser(r).ID, c)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserDTO(u))
}
