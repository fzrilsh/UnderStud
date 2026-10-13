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

// Export handles GET /api/v1/students/{id}/summaries/{sid}/export?format=pdf|txt.
func (h *SummaryHandlers) Export(w http.ResponseWriter, r *http.Request) {
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
	sid, err := uuid.Parse(chi.URLParam(r, "sid"))
	if err != nil {
		Error(w, utils.NewErrNotFound("Ringkasan tidak ditemukan."))
		return
	}
	format := r.URL.Query().Get("format")
	file, err := h.summary.Export(r.Context(), actor, id, sid, format)
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+file.Filename+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(file.Data)
}
