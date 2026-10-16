package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"backend/internal/db"
	"backend/internal/utils"
)

// Dashboard: agregat untuk halaman beranda BK & Wali Kelas.
type Dashboard struct {
	queries *db.Queries
}

func NewDashboard(q *db.Queries) *Dashboard { return &Dashboard{queries: q} }

// --- Bentuk response (mengikuti kebutuhan frontend) ---

type ChatStats struct {
	Waiting  int64 `json:"waiting"`
	Pinned   int64 `json:"pinned"`
	Drafted  int64 `json:"drafted"`
	Queued   int64 `json:"queued"`
	Unqueued int64 `json:"unqueued"`
}

type ChartPoint struct {
	Label   string `json:"label"` // "Mei", "Jun", ...
	Pinned  int64  `json:"pinned"`
	Drafted int64  `json:"drafted"`
	Queued  int64  `json:"queued"`
}

type ClassItem struct {
	Kelas    string `json:"kelas"`
	Grade    string `json:"grade"`
	Homeroom string `json:"homeroom"`
	Students int64  `json:"students"`
	RedFlag  int64  `json:"red_flag"`
	Status   string `json:"status"` // flagged | complete | not_filled
}

type RedTop struct {
	Student StudentItem `json:"student"`
	Tanggal string      `json:"tanggal"`
}

type LatestSummary struct {
	Student StudentItem `json:"student"`
	Summary SummaryView `json:"summary"`
}

type DashboardBK struct {
	Chats     ChatStats       `json:"chats"`
	RedFlags  int64           `json:"red_flags"`
	Summaries int64           `json:"summaries"`
	Latest    []LatestSummary `json:"latest"`
	RedTop    *RedTop         `json:"red_top"`
	Chart     []ChartPoint    `json:"chart"`
	Classes   []ClassItem     `json:"classes"`
}

type DashboardWK struct {
	Students       int64        `json:"students"`
	ERaporComplete int64        `json:"e_rapor_complete"`
	Observed       int64        `json:"observed"`
	RedFlags       int64        `json:"red_flags"`
	RedTop         *StudentItem `json:"red_top"`
	Chart          []ChartPoint `json:"chart"`
}

// monthLabels = nama bulan singkat Indonesia.
var monthLabels = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// gradeFromKelas: "XI IPA 1" → "XI"; "X TJKT 2" → "X"; "XII IPS 1" → "XII".
func gradeFromKelas(kelas string) string {
	up := strings.ToUpper(strings.TrimSpace(kelas))
	for _, g := range []string{"XII", "XI", "X"} {
		if strings.HasPrefix(up, g) {
			return g
		}
	}
	return ""
}

// buildChart: 6 titik bulan (5 bulan lalu s/d bulan ini) dari rows agregat.
func (d *Dashboard) buildChart(ctx context.Context) ([]ChartPoint, error) {
	rows, err := d.queries.CountChatsByMonth(ctx)
	if err != nil {
		return nil, utils.NewErrInternal("gagal memuat chart")
	}
	byMonth := map[string]db.CountChatsByMonthRow{}
	for _, r := range rows {
		byMonth[r.Bulan.Format("2006-01")] = r
	}
	now := time.Now().In(jakartaLoc())
	out := make([]ChartPoint, 0, 6)
	for i := 5; i >= 0; i-- {
		m := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		key := m.Format("2006-01")
		p := ChartPoint{Label: monthLabels[int(m.Month())-1]}
		if r, ok := byMonth[key]; ok {
			p.Pinned, p.Drafted, p.Queued = r.Pinned, r.Drafted, r.Queued
		}
		out = append(out, p)
	}
	return out, nil
}

// Classes: daftar kelas + statistik. BK = semua kelas; WK = kelasnya saja.
func (d *Dashboard) Classes(ctx context.Context, actor Actor) ([]ClassItem, error) {
	if err := requireScopeClass(actor); err != nil {
		return nil, err
	}
	stats, err := d.queries.ListClassStats(ctx)
	if err != nil {
		return nil, utils.NewErrInternal("gagal memuat data kelas")
	}
	homerooms, err := d.queries.ListHomerooms(ctx)
	if err != nil {
		return nil, utils.NewErrInternal("gagal memuat data kelas")
	}
	homeroomByKelas := map[string]string{}
	for _, h := range homerooms {
		if h.Kelas != nil {
			homeroomByKelas[*h.Kelas] = h.Name
		}
	}
	out := make([]ClassItem, 0, len(stats))
	for _, s := range stats {
		if !isBK(actor) && s.Kelas != actor.Kelas {
			continue
		}
		status := "complete"
		if s.RedFlag > 0 {
			status = "flagged"
		} else if s.Students == 0 {
			status = "not_filled"
		}
		out = append(out, ClassItem{
			Kelas: s.Kelas, Grade: gradeFromKelas(s.Kelas),
			Homeroom: homeroomByKelas[s.Kelas],
			Students: s.Students, RedFlag: s.RedFlag, Status: status,
		})
	}
	return out, nil
}

func (d *Dashboard) BK(ctx context.Context, actor Actor) (DashboardBK, error) {
	if !isBK(actor) {
		return DashboardBK{}, utils.NewErrForbidden("Hanya Guru BK yang bisa melihat dashboard ini.")
	}
	counts, err := d.queries.CountChatsByStatus(ctx)
	if err != nil {
		return DashboardBK{}, utils.NewErrInternal("gagal memuat dashboard")
	}
	active, err := d.queries.CountActiveStudents(ctx)
	if err != nil {
		return DashboardBK{}, utils.NewErrInternal("gagal memuat dashboard")
	}
	latest := []LatestSummary{}
	if rows, err := d.queries.ListLatestSummaries(ctx, 4); err == nil {
		for _, r := range rows {
			latest = append(latest, LatestSummary{
				Student: StudentItem{
					ID: r.StudentID.String(), Name: r.StudentName, Kelas: r.StudentKelas,
					HasRedFlag: r.StudentHasRedFlag,
				},
				Summary: toSummaryView(summaryFields{
					ID: r.ID, Content: r.Content, Sources: r.Sources,
					Model: r.Model, Status: r.Status, Error: r.Error, CreatedAt: r.CreatedAt,
					Version: r.Version, Note: r.Note,
				}),
			})
		}
	}
	var redTop *RedTop
	if row, err := d.queries.GetRedTopStudent(ctx); err == nil {
		item := StudentItem{ID: row.ID.String(), Name: row.Name, Kelas: row.Kelas, HasRedFlag: true}
		if row.Nisn != nil {
			item.NISN = *row.Nisn
		}
		redTop = &RedTop{Student: item, Tanggal: row.RedTanggal.Time.Format("2006-01-02")}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return DashboardBK{}, utils.NewErrInternal("gagal memuat dashboard")
	}
	chart, err := d.buildChart(ctx)
	if err != nil {
		return DashboardBK{}, err
	}
	classes, err := d.Classes(ctx, actor)
	if err != nil {
		return DashboardBK{}, err
	}
	return DashboardBK{
		Chats: ChatStats{
			Waiting: counts.Waiting, Pinned: counts.Pinned, Drafted: counts.Drafted,
			Queued: counts.Queued, Unqueued: counts.Unqueued,
		},
		RedFlags: active.RedFlags, Summaries: active.Summaries,
		Latest: latest, RedTop: redTop, Chart: chart, Classes: classes,
	}, nil
}

func (d *Dashboard) WK(ctx context.Context, actor Actor) (DashboardWK, error) {
	if err := requireScopeClass(actor); err != nil {
		return DashboardWK{}, err
	}
	counts, err := d.queries.CountStudentsByKelas(ctx, actor.Kelas)
	if err != nil {
		return DashboardWK{}, utils.NewErrInternal("gagal memuat dashboard")
	}
	var redTop *StudentItem
	if row, err := d.queries.GetRedTopStudentByKelas(ctx, actor.Kelas); err == nil {
		item := StudentItem{ID: row.ID.String(), Name: row.Name, Kelas: row.Kelas, HasRedFlag: true}
		if row.Nisn != nil {
			item.NISN = *row.Nisn
		}
		redTop = &item
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return DashboardWK{}, utils.NewErrInternal("gagal memuat dashboard")
	}
	chart, err := d.buildChart(ctx)
	if err != nil {
		return DashboardWK{}, err
	}
	return DashboardWK{
		Students: counts.Students, ERaporComplete: counts.ERaporComplete,
		Observed: counts.Observed, RedFlags: counts.RedFlags,
		RedTop: redTop, Chart: chart,
	}, nil
}
