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
	enqueue func(uuid.UUID, uuid.UUID, string)
}

func NewSummaryHandlers(s *service.Summary) *SummaryHandlers {
	return &SummaryHandlers{summary: s}
}

// WithEnqueue menyuntikkan hook queue AI (dipanggil server setelah queue dibuat).
func (h *SummaryHandlers) WithEnqueue(fn func(uuid.UUID, uuid.UUID, string)) *SummaryHandlers {
	h.enqueue = fn
	return h
}

// Generate handles POST /api/v1/students/{id}/summaries — trigger manual.
// Generate berjalan asinkron lewat queue AI; balas 202 lalu FE poll /latest.
func (h *SummaryHandlers) Generate(w http.ResponseWriter, r *http.Request) {
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
	if err := h.summary.CanGenerate(r.Context(), actor, id); err != nil {
		Error(w, err)
		return
	}
	if h.enqueue == nil {
		Error(w, utils.NewErrAINotConfigured("Ringkasan AI belum dikonfigurasi."))
		return
	}
	h.enqueue(id, actor.UserID, "manual")
	w.WriteHeader(http.StatusAccepted)
	OK(w, map[string]any{"status": "queued"})
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
	attachmentHeader(w, file.Filename)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(file.Data)
}
