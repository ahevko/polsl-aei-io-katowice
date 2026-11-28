package api

import (
	"net/http"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

// SubscriptionHandler – kontroler UC zarządzania subskrypcjami.
type SubscriptionHandler struct {
	subs *service.SubscriptionService
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	subs, err := h.subs.List(r.Context(), currentUser(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]subscriptionDTO, 0, len(subs))
	for _, s := range subs {
		out = append(out, toSubscriptionDTO(s))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	s, err := h.subs.Get(r.Context(), currentUser(r).ID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSubscriptionDTO(s))
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	in, ok := readSubscriptionInput(w, r)
	if !ok {
		return
	}
	s, err := h.subs.Create(r.Context(), currentUser(r).ID, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSubscriptionDTO(s))
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	in, ok := readSubscriptionInput(w, r)
	if !ok {
		return
	}
	s, err := h.subs.Update(r.Context(), currentUser(r).ID, r.PathValue("id"), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSubscriptionDTO(s))
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.subs.Delete(r.Context(), currentUser(r).ID, r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionHandler) Pause(w http.ResponseWriter, r *http.Request) {
	s, err := h.subs.Pause(r.Context(), currentUser(r).ID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSubscriptionDTO(s))
}

func (h *SubscriptionHandler) Resume(w http.ResponseWriter, r *http.Request) {
	var req resumeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	d, err := parseDate("nextPaymentDate", req.NextPaymentDate)
	if err != nil {
		writeError(w, err)
		return
	}
	s, err := h.subs.Resume(r.Context(), currentUser(r).ID, r.PathValue("id"), d)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toSubscriptionDTO(s))
}

func readSubscriptionInput(w http.ResponseWriter, r *http.Request) (service.SubscriptionInput, bool) {
	var req subscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return service.SubscriptionInput{}, false
	}
	in, err := req.toInput()
	if err != nil {
		writeError(w, err)
		return service.SubscriptionInput{}, false
	}
	return in, true
}
