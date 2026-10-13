package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"
)

type ReportHandlers struct {
	report *service.Report
}

func NewReportHandlers(report *service.Report) *ReportHandlers {
	return &ReportHandlers{report: report}
}

func chatIDParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, utils.NewErrNotFound("Laporan tidak ditemukan.")
	}
	return id, nil
}

// Queue handles GET /api/v1/reports.
func (h *ReportHandlers) Queue(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, total, err := h.report.Queue(r.Context(), limit, offset)
	if err != nil {
		Error(w, err)
		return
	}
	OKMeta(w, items, map[string]any{"limit": limit, "offset": offset, "total": total})
}

// History handles GET /api/v1/reports/history (riwayat: dilewati + di-hide BK).
func (h *ReportHandlers) History(w http.ResponseWriter, r *http.Request) {
	limit, offset := pageParams(r)
	items, total, err := h.report.History(r.Context(), limit, offset)
	if err != nil {
		Error(w, err)
		return
	}
	OKMeta(w, items, map[string]any{"limit": limit, "offset": offset, "total": total})
}

// Detail handles GET /api/v1/reports/{id}.
func (h *ReportHandlers) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		Error(w, err)
		return
	}
	d, err := h.report.Detail(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, d)
}

func (h *ReportHandlers) Pin(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		Error(w, err)
		return
	}
	if err := h.report.Pin(r.Context(), id); err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"ok": true})
}

func (h *ReportHandlers) Unpin(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		Error(w, err)
		return
	}
	if err := h.report.Unpin(r.Context(), id); err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"ok": true})
}

func (h *ReportHandlers) Draft(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		Error(w, err)
		return
	}
	if err := h.report.Draft(r.Context(), id); err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"ok": true})
}

func (h *ReportHandlers) Next(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		Error(w, err)
		return
	}
	if err := h.report.Next(r.Context(), id); err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"ok": true})
}

func (h *ReportHandlers) Hide(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		Error(w, err)
		return
	}
	if err := h.report.Hide(r.Context(), id); err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"ok": true})
}

// Reply handles POST /api/v1/reports/{id}/messages (JSON body {body}).
func (h *ReportHandlers) Reply(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		Error(w, err)
		return
	}
	staffID, err := uuid.Parse(middleware.UserIDFrom(r.Context()))
	if err != nil {
		Error(w, utils.NewErrUnauthorized("Sesi tidak valid. Silakan masuk kembali."))
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if err := Bind(r, &in); err != nil {
		Error(w, err)
		return
	}
	msg, err := h.report.Reply(r.Context(), staffID, id, in.Body)
	if err != nil {
		Error(w, err)
		return
	}
	Created(w, msg)
}
