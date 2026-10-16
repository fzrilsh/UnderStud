package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"backend/internal/db"
)

type job struct {
	studentID   uuid.UUID
	generatedBy uuid.UUID
	reason      string
	deadline    time.Time
}

// Queue: debounce per-siswa. Job baru untuk siswa yang sama membatalkan yang lama.
type Queue struct {
	queries  *db.Queries
	client   *Client
	debounce time.Duration
	logger   *slog.Logger

	mu      sync.Mutex
	pending map[uuid.UUID]*job

	// F2: batasi jumlah generate paralel (tiap panggilan AI bisa sampai 60s).
	// Worker tunggal → satu panggilan lambat memblokir SEMUA siswa (head-of-line).
	sem chan struct{}
}

func NewQueue(q *db.Queries, client *Client, debounce time.Duration, logger *slog.Logger) *Queue {
	return &Queue{
		queries: q, client: client, debounce: debounce, logger: logger,
		pending: map[uuid.UUID]*job{}, sem: make(chan struct{}, 4),
	}
}

// Enqueue menjadwalkan generate untuk siswa; aman dipanggil berkali-kali.
func (q *Queue) Enqueue(studentID, generatedBy uuid.UUID, reason string) {
	if !q.client.Configured() {
		return // AI nonaktif → no-op, save tetap sukses (spec §9.1)
	}
	q.mu.Lock()
	q.pending[studentID] = &job{
		studentID: studentID, generatedBy: generatedBy, reason: reason,
		deadline: time.Now().Add(q.debounce),
	}
	q.mu.Unlock()
}

// Start menjalankan loop worker sampai ctx dibatalkan.
func (q *Queue) Start(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.drain(ctx)
		}
	}
}

func (q *Queue) drain(ctx context.Context) {
	now := time.Now()
	var ready []*job
	q.mu.Lock()
	for id, j := range q.pending {
		if now.After(j.deadline) {
			ready = append(ready, j)
			delete(q.pending, id)
		}
	}
	q.mu.Unlock()
	// F2: jalankan job siap secara paralel (dibatasi semaphore). Tiap job
	// punya recover sendiri (F1) — panic di satu job tak mematikan worker.
	for _, j := range ready {
		q.sem <- struct{}{}
		go func(j *job) {
			defer func() { <-q.sem }()
			q.generateSafe(ctx, j)
		}(j)
	}
}

// generateSafe memanggil generate dengan pemulihan panic (F1).
func (q *Queue) generateSafe(ctx context.Context, j *job) {
	defer func() {
		if rec := recover(); rec != nil {
			q.logger.Error("ai worker panic recovered", "student", j.studentID, "panic", rec)
		}
	}()
	q.generate(ctx, j)
}

func (q *Queue) generate(ctx context.Context, j *job) {
	in, ci, ok, err := q.gather(ctx, j.studentID)
	if err != nil {
		q.logger.Warn("ai gather failed", "student", j.studentID, "error", err)
		return
	}
	if !ok {
		return // tidak ada data bermakna → tidak generate (spec §9.1)
	}
	narrative, err := q.client.Summarize(ctx, in)
	status, errMsg := "ready", (*string)(nil)
	if err != nil {
		status = "failed"
		m := err.Error()
		errMsg = &m
		q.logger.Warn("ai summarize failed", "student", j.studentID, "error", err)
	}
	// Bangun konten terstruktur DETERMINISTIK dari data siswa; AI hanya
	// menyumbang headline+narrative. Bila AI gagal, konten tetap tersimpan
	// (narrative kosong) sehingga struktur tidak hilang.
	if u, err := q.queries.GetUserByID(ctx, j.generatedBy); err == nil {
		ci.AuthorName = u.Name
	}
	content := BuildContent(ci, narrative.Headline, narrative.Narrative)
	contentJSON, _ := json.Marshal(content)
	sources, _ := json.Marshal(in.sources())
	model := q.client.model
	// Nomor versi: lanjut dari ringkasan sebelumnya (bila ada).
	version, err := q.queries.NextSummaryVersion(ctx, j.studentID)
	if err != nil || version < 1 {
		version = 1
	}
	note := "Menambahkan data terbaru"
	if version == 1 {
		note = "Ringkasan pertama"
	}
	if _, err := q.queries.CreateSummary(ctx, db.CreateSummaryParams{
		StudentID: j.studentID, GeneratedBy: j.generatedBy, TriggerReason: j.reason,
		Content: string(contentJSON), Sources: sources, Model: &model, Status: status, Error: errMsg,
		Version: version, Note: note,
	}); err != nil {
		q.logger.Error("ai save summary failed", "student", j.studentID, "error", err)
	}
}

// gather mengumpulkan data siswa → SummarizeInput (untuk AI) + ContentInput
// (untuk membangun struktur konten). ok=false bila tak ada data bermakna.
func (q *Queue) gather(ctx context.Context, studentID uuid.UUID) (SummarizeInput, ContentInput, bool, error) {
	st, err := q.queries.GetStudentByID(ctx, studentID)
	if err != nil {
		return SummarizeInput{}, ContentInput{}, false, err
	}
	prof, err := q.queries.GetStudentProfile(ctx, studentID)
	if err != nil {
		return SummarizeInput{}, ContentInput{}, false, err
	}
	obsRows, err := q.queries.ListObservationsByStudent(ctx, db.ListObservationsByStudentParams{
		StudentID: studentID, Limit: 200, Offset: 0,
	})
	if err != nil {
		return SummarizeInput{}, ContentInput{}, false, err
	}
	in := SummarizeInput{Kelas: st.Kelas}
	var eraporRows []EraporRow
	if len(prof.NilaiErapor) > 0 {
		var rows []map[string]any
		if json.Unmarshal(prof.NilaiErapor, &rows) == nil && len(rows) > 0 {
			b, _ := json.Marshal(rows)
			in.NilaiErapor = string(b)
		}
		_ = json.Unmarshal(prof.NilaiErapor, &eraporRows)
	}
	in.MinatBakat = deref(prof.MinatBakat)
	in.CatatanPsikologis = deref(prof.CatatanPsikologis)
	in.CatatanKasusPrestasi = deref(prof.CatatanKasusPrestasi)
	in.MasukanBK = deref(prof.MasukanBk)
	in.PendapatWaliKelas = deref(prof.PendapatWaliKelas)

	if prof.Sakit > 0 || prof.Izin > 0 || prof.Alpa > 0 {
		in.Kehadiran = fmt.Sprintf("sakit %d, izin %d, alpa %d", prof.Sakit, prof.Izin, prof.Alpa)
	}

	meaningful := in.NilaiErapor != "" || in.MinatBakat != "" || in.CatatanPsikologis != "" ||
		in.CatatanKasusPrestasi != "" || in.MasukanBK != "" || in.PendapatWaliKelas != "" ||
		in.Kehadiran != "" || len(obsRows) > 0
	if !meaningful {
		return SummarizeInput{}, ContentInput{}, false, nil
	}

	// Kumpulkan kategori & red flag & tanggal observasi terbaru.
	catSet := map[string]struct{}{}
	redCount := 0
	var lastObs time.Time
	for _, o := range obsRows {
		var items []string
		_ = json.Unmarshal(o.Items, &items)
		line := o.Tanggal.Time.Format("2006-01-02") + ": " + joinItems(items)
		if o.Catatan != nil && *o.Catatan != "" {
			line += " — " + *o.Catatan
		}
		in.Observasi = append(in.Observasi, line)
		var cats []string
		_ = json.Unmarshal(o.Categories, &cats)
		for _, c := range cats {
			catSet[c] = struct{}{}
		}
		if o.RedFlag {
			redCount++
		}
		if o.Tanggal.Valid && o.Tanggal.Time.After(lastObs) {
			lastObs = o.Tanggal.Time
		}
	}
	catLabels := make([]string, 0, len(catSet))
	for c := range catSet {
		catLabels = append(catLabels, kategoriLabel(c))
	}
	sort.Strings(catLabels)

	// Judul red flag: dari observasi red flag terbaru (bila ada).
	redTitle := ""
	if redCount > 0 {
		for _, o := range obsRows { // obsRows urut tanggal DESC
			if o.RedFlag {
				who := o.AuthorName
				if strings.TrimSpace(who) == "" {
					who = "Wali Kelas"
				}
				redTitle = fmt.Sprintf("Red flag dari %s · %s · BK sudah diberi tahu",
					who, fmtShortDate(o.Tanggal.Time))
				break
			}
		}
	}

	ci := ContentInput{
		StudentName:          st.Name,
		Nilai:                eraporRows,
		ObsCount:             len(obsRows),
		RedFlagCount:         redCount,
		Categories:           catLabels,
		LastObsDate:          lastObs,
		HasLastObsDate:       !lastObs.IsZero(),
		ProfileUpdated:       prof.UpdatedAt,
		MinatBakat:           in.MinatBakat,
		PendapatWaliKelas:    in.PendapatWaliKelas,
		MasukanBK:            in.MasukanBK,
		CatatanKasusPrestasi: in.CatatanKasusPrestasi,
		CatatanPsikologis:    in.CatatanPsikologis,
		RedFlagTitle:         redTitle,
	}
	return in, ci, true, nil
}

// kategoriLabel: kode kategori → label manusiawi.
func kategoriLabel(k string) string {
	if l, ok := kategoriLabels[k]; ok {
		return l
	}
	return k
}

var kategoriLabels = map[string]string{
	"emosi":     "Emosi",
	"sosial":    "Sosial",
	"akademik":  "Akademik & motivasi",
	"kehadiran": "Kehadiran & disiplin",
	"fisik":     "Fisik & kebiasaan",
	"keluarga":  "Situasi keluarga",
	"red_flag":  "Red flag",
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func joinItems(items []string) string {
	out := ""
	for i, it := range items {
		if i > 0 {
			out += ", "
		}
		out += it
	}
	return out
}

// sources mengembalikan field yang benar-benar terisi (spec §22).
func (in SummarizeInput) sources() []map[string]string {
	var out []map[string]string
	add := func(kelompok, sumber, tanggal string) {
		out = append(out, map[string]string{"kelompok": kelompok, "sumber": sumber, "tanggal": tanggal})
	}
	today := time.Now().Format("2006-01-02")
	if in.NilaiErapor != "" {
		add("Akademik", "nilai_erapor", today)
	}
	if in.Kehadiran != "" {
		add("Kehadiran", "absensi", today)
	}
	if in.MinatBakat != "" {
		add("Minat", "minat_bakat", today)
	}
	if in.CatatanPsikologis != "" {
		add("Kesejahteraan", "catatan_psikologis", today)
	}
	if in.CatatanKasusPrestasi != "" {
		add("Kasus/Prestasi", "catatan_kasus_prestasi", today)
	}
	if in.MasukanBK != "" {
		add("Masukan", "masukan_bk", today)
	}
	if in.PendapatWaliKelas != "" {
		add("Pendapat", "pendapat_wali_kelas", today)
	}
	if len(in.Observasi) > 0 {
		add("Observasi", "observations", today)
	}
	return out
}
