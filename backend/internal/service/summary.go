package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/utils"
)

type Summary struct {
	queries      *db.Queries
	aiConfigured bool
}

func NewSummary(q *db.Queries, aiConfigured bool) *Summary {
	return &Summary{queries: q, aiConfigured: aiConfigured}
}

type SummaryView struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Sources   any    `json:"sources"`
	Model     string `json:"model,omitempty"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
}

func toSummaryView(s db.AiSummary) SummaryView {
	v := SummaryView{
		ID: s.ID.String(), Content: s.Content, Status: s.Status,
		CreatedAt: s.CreatedAt.Format(time.RFC3339), Sources: []any{},
	}
	if len(s.Sources) > 0 {
		v.Sources = json.RawMessage(s.Sources)
	}
	if s.Model != nil {
		v.Model = *s.Model
	}
	if s.Error != nil {
		v.Error = *s.Error
	}
	return v
}

func (s *Summary) ensureScope(ctx context.Context, actor Actor, studentID uuid.UUID) error {
	st, err := s.queries.GetStudentByID(ctx, studentID)
	if err != nil {
		return utils.NewErrNotFound("Siswa tidak ditemukan.")
	}
	if actor.Role != middleware.RoleBK && st.Kelas != actor.Kelas {
		return utils.NewErrOutOfScope("Siswa ini di luar cakupanmu.")
	}
	return nil
}

func (s *Summary) Latest(ctx context.Context, actor Actor, studentID uuid.UUID) (SummaryView, error) {
	if err := s.ensureScope(ctx, actor, studentID); err != nil {
		return SummaryView{}, err
	}
	row, err := s.queries.GetLatestSummary(ctx, studentID)
	if err != nil {
		if !s.aiConfigured {
			return SummaryView{}, utils.NewErrAINotConfigured("Ringkasan AI belum dikonfigurasi.")
		}
		return SummaryView{}, utils.NewErrNotFound("Belum ada ringkasan untuk siswa ini.")
	}
	return toSummaryView(row), nil
}

func (s *Summary) List(ctx context.Context, actor Actor, studentID uuid.UUID, limit, offset int32) ([]SummaryView, int64, error) {
	if err := s.ensureScope(ctx, actor, studentID); err != nil {
		return nil, 0, err
	}
	rows, err := s.queries.ListSummaries(ctx, db.ListSummariesParams{
		StudentID: studentID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat ringkasan")
	}
	out := make([]SummaryView, 0, len(rows))
	for _, r := range rows {
		out = append(out, toSummaryView(r))
	}
	// H5: total = jumlah seluruh versi (bukan ukuran halaman).
	total, err := s.queries.CountSummaries(ctx, studentID)
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat ringkasan")
	}
	return out, total, nil
}
