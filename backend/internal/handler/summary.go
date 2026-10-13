package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/service"
	"backend/internal/utils"
)

type SummaryHandlers struct {
	summary *service.Summary
}

func NewSummaryHandlers(s *service.Summary) *SummaryHandlers {
	return &SummaryHandlers{summary: s}
}

// Latest handles GET /api/v1/students/{id}/summaries/latest.
func (h *SummaryHandlers) Latest(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	id, err := studentIDParam(r)
	if err != nil {
		Error(w, err)
		return
	}
	view, err := h.summary.Latest(r.Context(), actor, id)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, view)
}

// List handles GET /api/v1/students/{id}/summaries.
func (h *SummaryHandlers) List(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	id, err := studentIDParam(r)
	if err != nil {
		Error(w, err)
		return
	}
	limit, offset := pageParams(r)
	items, total, err := h.summary.List(r.Context(), actor, id, limit, offset)
	if err != nil {
		Error(w, err)
		return
	}
	OKMeta(w, items, map[string]any{"limit": limit, "offset": offset, "total": total})
}

// Export handles GET /api/v1/students/{id}/summaries/{sid}/export — T3 (501 dulu).
func (h *SummaryHandlers) Export(w http.ResponseWriter, r *http.Request) {
	if _, err := uuid.Parse(chi.URLParam(r, "sid")); err != nil {
		Error(w, utils.NewErrNotFound("Ringkasan tidak ditemukan."))
		return
	}
	Error(w, &utils.AppErr{Status: http.StatusNotImplemented, Code: "NOT_IMPLEMENTED", Message: "Ekspor tersedia di rilis berikutnya."})
}
