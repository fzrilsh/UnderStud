package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"
)

type StudentHandlers struct {
	student *service.Student
	obs     *service.Observasi
	summary *service.Summary
	imports *service.Importer
	enqueue func(uuid.UUID, uuid.UUID, string)
}

func NewStudentHandlers(s *service.Student, o *service.Observasi, sum *service.Summary, imp *service.Importer) *StudentHandlers {
	return &StudentHandlers{student: s, obs: o, summary: sum, imports: imp}
}

// WithEnqueue menyuntikkan hook queue AI (dipanggil server setelah queue dibuat).
func (h *StudentHandlers) WithEnqueue(fn func(uuid.UUID, uuid.UUID, string)) *StudentHandlers {
	h.enqueue = fn
	return h
}

func (h *StudentHandlers) enqueueAI(studentID, by uuid.UUID, reason string) {
	if h.enqueue != nil {
		h.enqueue(studentID, by, reason)
	}
}

// actorFrom membangun service.Actor dari context (RequireAuth sudah jalan).
func actorFrom(r *http.Request) (service.Actor, error) {
	id, err := uuid.Parse(middleware.UserIDFrom(r.Context()))
	if err != nil {
		return service.Actor{}, utils.NewErrUnauthorized("Sesi tidak valid. Silakan masuk kembali.")
	}
	return service.Actor{UserID: id, Role: middleware.RoleFrom(r.Context()), Kelas: middleware.KelasFrom(r.Context())}, nil
}

func studentIDParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, utils.NewErrNotFound("Siswa tidak ditemukan.")
	}
	return id, nil
}

// List handles GET /api/v1/students.
func (h *StudentHandlers) List(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	limit, offset := pageParams(r)
	includeArchived := r.URL.Query().Get("include_archived") == "true"
	filter := r.URL.Query().Get("filter")
	items, total, err := h.student.List(r.Context(), actor, r.URL.Query().Get("q"), includeArchived, filter, limit, offset)
	if err != nil {
		Error(w, err)
		return
	}
	OKMeta(w, items, map[string]any{"limit": limit, "offset": offset, "total": total})
}

// Create handles POST /api/v1/students.
func (h *StudentHandlers) Create(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	var in service.CreateStudentInput
	if err := Bind(r, &in); err != nil {
		Error(w, err)
		return
	}
	item, err := h.student.Create(r.Context(), actor, in)
	if err != nil {
		Error(w, err)
		return
	}
	Created(w, item)
}

// Detail handles GET /api/v1/students/{id} — gabungan + latest_summary.
func (h *StudentHandlers) Detail(w http.ResponseWriter, r *http.Request) {
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
	detail, err := h.student.Detail(r.Context(), actor, id)
	if err != nil {
		Error(w, err)
		return
	}
	// latest_summary best-effort: error (belum ada / AI off) → null, tidak gagal.
	payload := map[string]any{
		"student": detail.Student, "profile": detail.Profile, "observations": detail.Observations,
		"summaries": detail.Summaries, "latest_summary": nil,
	}
	if sum, err := h.summary.Latest(r.Context(), actor, id); err == nil {
		payload["latest_summary"] = sum
	}
	OK(w, payload)
}

// Update handles PATCH /api/v1/students/{id}.
func (h *StudentHandlers) Update(w http.ResponseWriter, r *http.Request) {
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
	var in service.UpdateStudentInput
	if err := Bind(r, &in); err != nil {
		Error(w, err)
		return
	}
	item, err := h.student.Update(r.Context(), actor, id, in)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, item)
}

// Archive handles POST /api/v1/students/{id}/archive.
func (h *StudentHandlers) Archive(w http.ResponseWriter, r *http.Request) {
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
	if err := h.student.Archive(r.Context(), actor, id); err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"ok": true})
}

// Restore handles POST /api/v1/students/{id}/restore.
func (h *StudentHandlers) Restore(w http.ResponseWriter, r *http.Request) {
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
	if err := h.student.Restore(r.Context(), actor, id); err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"ok": true})
}

// SaveProfile handles PUT /api/v1/students/{id}/profile — trigger AI.
func (h *StudentHandlers) SaveProfile(w http.ResponseWriter, r *http.Request) {
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
	var in service.ProfileInput
	if err := Bind(r, &in); err != nil {
		Error(w, err)
		return
	}
	view, err := h.student.SaveProfile(r.Context(), actor, id, in)
	if err != nil {
		Error(w, err)
		return
	}
	h.enqueueAI(id, actor.UserID, "profile_updated")
	OK(w, view)
}

// CreateObservation handles POST /api/v1/students/{id}/observations — trigger AI.
func (h *StudentHandlers) CreateObservation(w http.ResponseWriter, r *http.Request) {
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
	var in service.ObservationInput
	if err := Bind(r, &in); err != nil {
		Error(w, err)
		return
	}
	view, err := h.obs.Create(r.Context(), actor, id, in)
	if err != nil {
		Error(w, err)
		return
	}
	h.enqueueAI(id, actor.UserID, "observation_added")
	Created(w, view)
}

// ListObservations handles GET /api/v1/students/{id}/observations.
func (h *StudentHandlers) ListObservations(w http.ResponseWriter, r *http.Request) {
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
	items, total, err := h.obs.List(r.Context(), actor, id, limit, offset)
	if err != nil {
		Error(w, err)
		return
	}
	OKMeta(w, items, map[string]any{"limit": limit, "offset": offset, "total": total})
}

// UpdateObservation handles PUT /api/v1/observations/{id} — trigger AI.
func (h *StudentHandlers) UpdateObservation(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, utils.NewErrNotFound("Observasi tidak ditemukan."))
		return
	}
	var in service.ObservationInput
	if err := Bind(r, &in); err != nil {
		Error(w, err)
		return
	}
	view, err := h.obs.Update(r.Context(), actor, id, in)
	if err != nil {
		Error(w, err)
		return
	}
	if sid, err := uuid.Parse(view.StudentID); err == nil {
		h.enqueueAI(sid, actor.UserID, "observation_edited")
	}
	OK(w, view)
}

// Revisions handles GET /api/v1/observations/{id}/revisions.
func (h *StudentHandlers) Revisions(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, utils.NewErrNotFound("Observasi tidak ditemukan."))
		return
	}
	revs, err := h.obs.Revisions(r.Context(), actor, id)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, revs)
}

// HandleRedFlag handles POST /api/v1/observations/{id}/handle (BK).
func (h *StudentHandlers) HandleRedFlag(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, utils.NewErrNotFound("Observasi tidak ditemukan."))
		return
	}
	if err := h.obs.Handle(r.Context(), actor, id); err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"ok": true})
}

// RedFlags handles GET /api/v1/red-flags.
func (h *StudentHandlers) RedFlags(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	limit, offset := pageParams(r)
	items, total, err := h.obs.RedFlags(r.Context(), actor, limit, offset)
	if err != nil {
		Error(w, err)
		return
	}
	OKMeta(w, items, map[string]any{"limit": limit, "offset": offset, "total": total})
}

// ImportPreview handles POST /api/v1/students/import/preview (multipart, field "file").
func (h *StudentHandlers) ImportPreview(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	if err := r.ParseMultipartForm(4 << 20); err != nil {
		Error(w, utils.NewErrInvalidInput("Gagal membaca file."))
		return
	}
	f, fh, err := r.FormFile("file")
	if err != nil {
		Error(w, utils.NewErrInvalidInput("File leger wajib diunggah."))
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 3<<20))
	if err != nil {
		Error(w, utils.NewErrInvalidInput("Gagal membaca file."))
		return
	}
	prev, err := h.imports.Preview(r.Context(), actor, fh.Filename, data)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, prev)
}

// ImportTemplate handles GET /api/v1/students/import/template?format=xlsx|csv.
func (h *StudentHandlers) ImportTemplate(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	if actor.Role != middleware.RoleWaliKelas {
		Error(w, utils.NewErrForbidden("Hanya Wali Kelas yang bisa mengunduh template."))
		return
	}
	data, ct, fn, err := service.BuildTemplate(r.URL.Query().Get("format"), actor.Kelas)
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-Type", ct)
	attachmentHeader(w, fn)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ImportCommit handles POST /api/v1/students/import/commit (JSON {preview_token}).
func (h *StudentHandlers) ImportCommit(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	var in struct {
		PreviewToken string `json:"preview_token"`
	}
	if err := Bind(r, &in); err != nil {
		Error(w, err)
		return
	}
	res, err := h.imports.Commit(r.Context(), actor, in.PreviewToken)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, res)
}
