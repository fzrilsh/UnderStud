package service

import (
	"context"
	"encoding/json"
	"strings"
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
	ID           string          `json:"id"`
	Version      int32           `json:"version"`
	Content      json.RawMessage `json:"content"` // objek JSON: headline/narrative/metrics/insights/data_used/red_flag
	Note         string          `json:"note"`
	Author       string          `json:"author"`
	Sources      any             `json:"sources"`
	SourcesUsed  int             `json:"sources_used"`
	SourcesTotal int             `json:"sources_total"`
	Model        string          `json:"model,omitempty"`
	Status       string          `json:"status"`
	Error        string          `json:"error,omitempty"`
	CreatedAt    string          `json:"created_at"`
}

// summaryContent: bentuk minimal untuk menghitung sources_used/total & fallback.
type summaryContent struct {
	DataUsed []struct {
		Status string `json:"status"`
	} `json:"data_used"`
}

const sourcesTotal = 8

// summaryFields = field bersama db.AiSummary & row types (dengan author_name).
type summaryFields struct {
	ID         uuid.UUID
	Content    string
	Sources    []byte
	Model      *string
	Status     string
	Error      *string
	CreatedAt  time.Time
	Version    int32
	Note       string
	AuthorName string
}

func summaryFieldsFromModel(s db.AiSummary) summaryFields {
	return summaryFields{
		ID: s.ID, Content: s.Content, Sources: s.Sources, Model: s.Model,
		Status: s.Status, Error: s.Error, CreatedAt: s.CreatedAt,
		Version: s.Version, Note: s.Note,
	}
}

func summaryFieldsFromLatest(s db.GetLatestSummaryRow) summaryFields {
	return summaryFields{
		ID: s.ID, Content: s.Content, Sources: s.Sources, Model: s.Model,
		Status: s.Status, Error: s.Error, CreatedAt: s.CreatedAt,
		Version: s.Version, Note: s.Note, AuthorName: s.AuthorName,
	}
}

func summaryFieldsFromList(s db.ListSummariesRow) summaryFields {
	return summaryFields{
		ID: s.ID, Content: s.Content, Sources: s.Sources, Model: s.Model,
		Status: s.Status, Error: s.Error, CreatedAt: s.CreatedAt,
		Version: s.Version, Note: s.Note, AuthorName: s.AuthorName,
	}
}

func summaryFieldsFromByID(s db.GetSummaryByIDRow) summaryFields {
	return summaryFields{
		ID: s.ID, Content: s.Content, Sources: s.Sources, Model: s.Model,
		Status: s.Status, Error: s.Error, CreatedAt: s.CreatedAt,
		Version: s.Version, Note: s.Note, AuthorName: s.AuthorName,
	}
}

// toSummaryView: content di-emit sebagai objek JSON (bila valid). Data lama
// (non-JSON) dibungkus agar tetap valid JSON ({"narrative": "..."}).
func toSummaryView(s summaryFields) SummaryView {
	raw := json.RawMessage(s.Content)
	if !json.Valid(raw) || len(strings.TrimSpace(s.Content)) == 0 {
		wrapped, _ := json.Marshal(map[string]string{"narrative": s.Content})
		raw = wrapped
	}
	v := SummaryView{
		ID: s.ID.String(), Version: s.Version, Content: raw, Note: s.Note,
		Author: s.AuthorName, Status: s.Status,
		CreatedAt: s.CreatedAt.Format(time.RFC3339), Sources: []any{},
		SourcesTotal: sourcesTotal,
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
	var c summaryContent
	if json.Unmarshal(raw, &c) == nil {
		for _, d := range c.DataUsed {
			if d.Status == "used" {
				v.SourcesUsed++
			}
		}
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

// CanGenerate: validasi scope + AI aktif sebelum trigger generate manual.
func (s *Summary) CanGenerate(ctx context.Context, actor Actor, studentID uuid.UUID) error {
	if err := s.ensureScope(ctx, actor, studentID); err != nil {
		return err
	}
	if !s.aiConfigured {
		return utils.NewErrAINotConfigured("Ringkasan AI belum dikonfigurasi.")
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
	return toSummaryView(summaryFieldsFromLatest(row)), nil
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
		out = append(out, toSummaryView(summaryFieldsFromList(r)))
	}
	// H5: total = jumlah seluruh versi (bukan ukuran halaman).
	total, err := s.queries.CountSummaries(ctx, studentID)
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat ringkasan")
	}
	return out, total, nil
}
