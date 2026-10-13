package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"
)

type AttachmentHandlers struct {
	chat *service.Chat
}

func NewAttachmentHandlers(chat *service.Chat) *AttachmentHandlers {
	return &AttachmentHandlers{chat: chat}
}

// Serve handles GET /api/v1/attachments/{id} (staf) dan
// GET /api/v1/lapor/attachments/{id} (siswa anonim pemilik).
// Akses: BK (semua) ATAU siswa pemilik chat (session cocok).
func (h *AttachmentHandlers) Serve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, utils.NewErrNotFound("Lampiran tidak ditemukan."))
		return
	}
	meta, err := h.chat.AttachmentMeta(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}

	allowed := false
	if middleware.RoleFrom(r.Context()) == middleware.RoleBK {
		allowed = true
	} else if session, ok := middleware.AnonSessionFrom(r.Context()); ok {
		if _, err := h.chat.Detail(r.Context(), session, meta.ChatID); err == nil {
			allowed = true
		}
	}
	if !allowed {
		Error(w, utils.NewErrNotFound("Lampiran tidak ditemukan."))
		return
	}

	f, err := h.chat.Store().Open(meta.StoredPath)
	if err != nil {
		Error(w, utils.NewErrNotFound("Lampiran tidak ditemukan."))
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", meta.Mime)
	w.Header().Set("Content-Disposition", `attachment; filename="`+meta.OriginalName+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, meta.OriginalName, time.Time{}, f)
}
