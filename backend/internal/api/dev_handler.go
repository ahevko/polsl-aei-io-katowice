package api

import (
	"net/http"
	"time"

	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/notify"
	"github.com/ahevko/polsl-aei-io-katowice/backend/internal/service"
)

// Narzędzia demonstracyjne prototypu: ręczne uruchomienie Workera dla
// wybranej daty i podgląd skrzynki symulatora e-mail. Wyłączane flagą -dev=false.
type DevHandler struct {
	worker *service.BillingWorker
	email  *notify.EmailSimulator
}

func (h *DevHandler) RunWorker(w http.ResponseWriter, r *http.Request) {
	day := time.Now().UTC()
	if s := r.URL.Query().Get("date"); s != "" {
		d, err := parseDate("date", s)
		if err != nil {
			writeError(w, err)
			return
		}
		day = d
	}
	rep, err := h.worker.RunOnce(r.Context(), day)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workerReportDTO{Day: rep.Day.Format(dateLayout),
		PaymentsRecorded: rep.PaymentsRecorded, RemindersSent: rep.RemindersSent})
}

func (h *DevHandler) Outbox(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out := []notify.Message{}
	for _, m := range h.email.Outbox() {
		if m.To == u.Email {
			out = append(out, m)
		}
	}
	writeJSON(w, http.StatusOK, out)
}
