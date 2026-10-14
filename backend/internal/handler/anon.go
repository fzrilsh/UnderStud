package handler

import (
	"net/http"

	"backend/internal/middleware"
	"backend/internal/service"
)

type AnonHandlers struct {
	anon *service.Anon
	chat *service.Chat
}

func NewAnonHandlers(anon *service.Anon, chat *service.Chat) *AnonHandlers {
	return &AnonHandlers{anon: anon, chat: chat}
}

// Bootstrap handles GET /api/v1/lapor — memastikan sesi ada, kembalikan identitas + daftar chat.
// (Middleware ResolveAnon sudah membuat sesi bila belum ada.)
// F23: kirim daftar chat nyata (bukan hardcode []) sesuai spec §21.4.
func (h *AnonHandlers) Bootstrap(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.AnonSessionFrom(r.Context())
	if !ok {
		Error(w, service.ErrAnonUnavailable)
		return
	}
	limit, offset := pageParams(r)
	chats, _, err := h.chat.ListBySession(r.Context(), session, limit, offset)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"profile": h.anon.GetProfile(session), "chats": chats})
}

// GetProfile handles GET /api/v1/lapor/profile.
func (h *AnonHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.AnonSessionFrom(r.Context())
	if !ok {
		Error(w, service.ErrAnonUnavailable)
		return
	}
	OK(w, map[string]any{"profile": h.anon.GetProfile(session)})
}

// UpdateProfile handles PUT /api/v1/lapor/profile.
func (h *AnonHandlers) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.AnonSessionFrom(r.Context())
	if !ok {
		Error(w, service.ErrAnonUnavailable)
		return
	}
	var in service.UpdateProfileInput
	if err := Bind(r, &in); err != nil {
		Error(w, err)
		return
	}
	ident, err := h.anon.UpdateProfile(r.Context(), session, in)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"profile": ident})
}
