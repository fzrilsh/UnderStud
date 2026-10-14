package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"
)

type ChatHandlers struct {
	chat *service.Chat
}

func NewChatHandlers(chat *service.Chat) *ChatHandlers { return &ChatHandlers{chat: chat} }

// parseMultipart membaca body + files[] dari multipart.
func parseMultipart(r *http.Request) (string, []service.UploadFile, error) {
	if err := r.ParseMultipartForm(24 << 20); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return "", nil, utils.NewErrPayloadTooLarge("Ukuran permintaan melebihi batas.")
		}
		return "", nil, utils.NewErrInvalidInput("Gagal membaca form.")
	}
	body := r.FormValue("body")
	var files []service.UploadFile
	if r.MultipartForm != nil {
		for _, fh := range r.MultipartForm.File["files"] {
			f, err := fh.Open()
			if err != nil {
				return "", nil, utils.NewErrInvalidInput("Gagal membaca file.")
			}
			data := make([]byte, 0, fh.Size)
			buf := make([]byte, 32*1024)
			for {
				n, err := f.Read(buf)
				if n > 0 {
					data = append(data, buf[:n]...)
				}
				if err != nil {
					break
				}
			}
			f.Close()
			files = append(files, service.UploadFile{Name: fh.Filename, Data: data})
		}
	}
	return body, files, nil
}

func atof(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// Create handles POST /api/v1/lapor/chats (multipart).
func (h *ChatHandlers) Create(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.AnonSessionFrom(r.Context())
	if !ok {
		Error(w, service.ErrAnonUnavailable)
		return
	}
	body, files, err := parseMultipart(r)
	if err != nil {
		Error(w, err)
		return
	}
	view, err := h.chat.Create(r.Context(), session, service.CreateChatInput{
		Body:         body,
		IdentityMode: r.FormValue("identity_mode"),
		Files:        files,
		Lat:          atof(r.FormValue("lat")),
		Lng:          atof(r.FormValue("lng")),
		Accuracy:     atof(r.FormValue("accuracy")),
	})
	if err != nil {
		Error(w, err)
		return
	}
	Created(w, view)
}

// List handles GET /api/v1/lapor/chats.
func (h *ChatHandlers) List(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.AnonSessionFrom(r.Context())
	if !ok {
		Error(w, service.ErrAnonUnavailable)
		return
	}
	limit, offset := pageParams(r)
	list, total, err := h.chat.ListBySession(r.Context(), session, limit, offset)
	if err != nil {
		Error(w, err)
		return
	}
	OKMeta(w, list, map[string]any{"limit": limit, "offset": offset, "total": total})
}

// Detail handles GET /api/v1/lapor/chats/{id}.
func (h *ChatHandlers) Detail(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.AnonSessionFrom(r.Context())
	if !ok {
		Error(w, service.ErrAnonUnavailable)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, utils.NewErrNotFound("Laporan tidak ditemukan."))
		return
	}
	detail, err := h.chat.Detail(r.Context(), session, id)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, detail)
}

// Reply handles POST /api/v1/lapor/chats/{id}/messages (multipart).
func (h *ChatHandlers) Reply(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.AnonSessionFrom(r.Context())
	if !ok {
		Error(w, service.ErrAnonUnavailable)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, utils.NewErrNotFound("Laporan tidak ditemukan."))
		return
	}
	body, files, err := parseMultipart(r)
	if err != nil {
		Error(w, err)
		return
	}
	msg, err := h.chat.Reply(r.Context(), session, id, service.ReplyInput{Body: body, Files: files})
	if err != nil {
		Error(w, err)
		return
	}
	Created(w, msg)
}

// Hide handles DELETE /api/v1/lapor/chats/{id}.
func (h *ChatHandlers) Hide(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.AnonSessionFrom(r.Context())
	if !ok {
		Error(w, service.ErrAnonUnavailable)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, utils.NewErrNotFound("Laporan tidak ditemukan."))
		return
	}
	if err := h.chat.HideByStudent(r.Context(), session, id); err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"ok": true})
}

// pageParams membaca ?limit=&offset= dengan default & clamp (spec §21.2).
func pageParams(r *http.Request) (int32, int32) {
	limit := int32(20)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = int32(n)
			if limit > 100 {
				limit = 100
			}
		}
	}
	offset := int32(0)
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	return limit, offset
}
