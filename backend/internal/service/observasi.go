package service

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/utils"
)

// ChecklistItem = satu item observasi (referensi tetap, spec §4.4).
type ChecklistItem struct {
	Key      string `json:"key"`      // "<kategori>:<item>"
	Kategori string `json:"kategori"`
	Label    string `json:"label"`
}

// checklistCatalog = 28 item + tidak_ada, hardcoded (spec §4.4).
var checklistCatalog = []ChecklistItem{
	// 1 — Emosi
	{"emosi:mudah_marah", "emosi", "Mudah marah atau emosi meledak"},
	{"emosi:mudah_menangis", "emosi", "Mudah menangis tanpa sebab jelas"},
	{"emosi:murung", "emosi", "Sering murung atau terlihat sedih"},
	{"emosi:mood_berubah", "emosi", "Mood berubah drastis dalam sehari"},
	{"emosi:cemas_berlebihan", "emosi", "Tampak cemas berlebihan (gelisah, takut ditanya, takut ujian)"},
	// 2 — Sosial
	{"sosial:menyendiri", "sosial", "Menyendiri atau menarik diri dari teman"},
	{"sosial:konflik", "sosial", "Sering konflik atau berkelahi dengan teman"},
	{"sosial:korban_ejekan", "sosial", "Jadi korban ejekan atau dijauhi"},
	{"sosial:mendominasi", "sosial", "Mengejek atau mendominasi teman lain"},
	{"sosial:ganti_pertemanan", "sosial", "Tiba-tiba ganti lingkaran pertemanan"},
	// 3 — Akademik dan motivasi
	{"akademik:nilai_turun", "akademik", "Nilai atau kualitas tugas turun mendadak"},
	{"akademik:sulit_fokus", "akademik", "Sulit fokus atau sering melamun di kelas"},
	{"akademik:hilang_minat", "akademik", "Kehilangan minat pada pelajaran/kegiatan yang dulu disukai"},
	{"akademik:tidak_kerjakan_tugas", "akademik", "Sering tidak mengerjakan tugas"},
	{"akademik:perfeksionis", "akademik", "Terlalu perfeksionis atau panik kalau salah"},
	// 4 — Kehadiran dan disiplin
	{"kehadiran:terlambat_bolos", "kehadiran", "Sering terlambat atau bolos"},
	{"kehadiran:izin_sakit", "kehadiran", "Sering izin sakit tanpa kejelasan"},
	{"kehadiran:melanggar_aturan", "kehadiran", "Melanggar aturan berulang"},
	// 5 — Fisik dan kebiasaan
	{"fisik:mengantuk_lelah", "fisik", "Sering mengantuk atau kelelahan"},
	{"fisik:nafsu_makan", "fisik", "Perubahan nafsu makan atau berat badan yang terlihat"},
	{"fisik:sakit_fisik", "fisik", "Sering mengeluh sakit kepala atau perut"},
	{"fisik:kebersihan_menurun", "fisik", "Penampilan atau kebersihan diri menurun drastis"},
	// 6 — Situasi keluarga
	{"keluarga:perubahan_kondisi", "keluarga", "Perubahan kondisi di rumah (perceraian, kehilangan, masalah ekonomi)"},
	{"keluarga:kurang_diperhatikan", "keluarga", "Tidak ada yang menjemput atau tampak kurang diperhatikan di rumah"},
	// 7 — Red flag
	{"red_flag:menyakiti_diri", "red_flag", "Mengucapkan ingin menyakiti diri atau tidak ingin hidup"},
	{"red_flag:luka_tidak_wajar", "red_flag", "Luka atau memar yang tidak wajar"},
	{"red_flag:mengancam", "red_flag", "Mengancam orang lain"},
	{"red_flag:takut_orang", "red_flag", "Tampak ketakutan terhadap seseorang tertentu"},
}

var checklistSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(checklistCatalog))
	for _, it := range checklistCatalog {
		m[it.Key] = struct{}{}
	}
	return m
}()

// ChecklistCatalog mengembalikan salinan catalog (untuk endpoint FE).
func ChecklistCatalog() []ChecklistItem {
	out := make([]ChecklistItem, len(checklistCatalog))
	copy(out, checklistCatalog)
	return out
}

// ValidateItems memvalidasi item mentah, mengembalikan items ter-dedup,
// categories (turunan, urut), dan redFlag (spec §4.4).
func ValidateItems(raw []string) ([]string, []string, bool, error) {
	if len(raw) == 0 {
		return nil, nil, false, utils.NewErrInvalidInput("Pilih minimal satu item observasi.")
	}
	seen := map[string]struct{}{}
	var items []string
	hasTidakAda := false
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if r == "tidak_ada" {
			hasTidakAda = true
			continue
		}
		if _, ok := checklistSet[r]; !ok {
			return nil, nil, false, utils.NewErrInvalidInput("Item observasi tidak dikenal: " + r)
		}
		if _, dup := seen[r]; dup {
			continue
		}
		seen[r] = struct{}{}
		items = append(items, r)
	}
	if hasTidakAda && len(items) > 0 {
		return nil, nil, false, utils.NewErrInvalidInput("'Tidak ada indikator' tidak bisa digabung dengan item lain.")
	}
	if hasTidakAda {
		return []string{"tidak_ada"}, nil, false, nil
	}
	if len(items) == 0 {
		return nil, nil, false, utils.NewErrInvalidInput("Pilih minimal satu item observasi.")
	}
	catSet := map[string]struct{}{}
	redFlag := false
	for _, it := range items {
		cat := it[:strings.IndexByte(it, ':')]
		catSet[cat] = struct{}{}
		if cat == "red_flag" {
			redFlag = true
		}
	}
	cats := make([]string, 0, len(catSet))
	for c := range catSet {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return items, cats, redFlag, nil
}

// --- Service observasi ---

type Observasi struct {
	queries *db.Queries
}

func NewObservasi(q *db.Queries) *Observasi { return &Observasi{queries: q} }

type ObservationInput struct {
	Items        []string `json:"items"`
	Catatan      string   `json:"catatan"`
	Tanggal      string   `json:"tanggal"`       // YYYY-MM-DD
	TindakanAwal string   `json:"tindakan_awal"` // pantau|konsultasi_bk|eskalasi_segera
}

type RevisionView struct {
	ID        string `json:"id"`
	EditedBy  string `json:"edited_by"`
	Snapshot  any    `json:"snapshot"`
	CreatedAt string `json:"created_at"`
}

type RedFlagItem struct {
	StudentID     string   `json:"student_id"`
	StudentName   string   `json:"student_name"`
	Kelas         string   `json:"kelas"`
	ObservationID string   `json:"observation_id"`
	Tanggal       string   `json:"tanggal"`
	RedFlagItems  []string `json:"red_flag_items"`
}

func (o *Observasi) List(ctx context.Context, actor Actor, studentID uuid.UUID, limit, offset int32) ([]ObservationView, int64, error) {
	if _, err := o.ensureStudentScope(ctx, actor, studentID); err != nil {
		return nil, 0, err
	}
	rows, err := o.queries.ListObservationsByStudent(ctx, db.ListObservationsByStudentParams{
		StudentID: studentID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat observasi")
	}
	total, err := o.queries.CountObservationsByStudent(ctx, studentID)
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat observasi")
	}
	out := make([]ObservationView, 0, len(rows))
	for _, r := range rows {
		out = append(out, toObservationView(r, ""))
	}
	return out, total, nil
}

func (o *Observasi) ensureStudentScope(ctx context.Context, actor Actor, studentID uuid.UUID) (db.Student, error) {
	st, err := o.queries.GetStudentByID(ctx, studentID)
	if err != nil {
		return db.Student{}, utils.NewErrNotFound("Siswa tidak ditemukan.")
	}
	if actor.Role != middleware.RoleBK && st.Kelas != actor.Kelas {
		return db.Student{}, utils.NewErrOutOfScope("Siswa ini di luar cakupanmu.")
	}
	return st, nil
}

// studentOf: observasi → student_id, lalu cek scope.
func (o *Observasi) studentOf(ctx context.Context, actor Actor, obsID uuid.UUID) (db.Observation, db.Student, error) {
	obs, err := o.queries.GetObservationByID(ctx, obsID)
	if err != nil {
		return db.Observation{}, db.Student{}, utils.NewErrNotFound("Observasi tidak ditemukan.")
	}
	st, err := o.ensureStudentScope(ctx, actor, obs.StudentID)
	if err != nil {
		return db.Observation{}, db.Student{}, err
	}
	return obs, st, nil
}

func validateObservation(in ObservationInput) error {
	if len([]rune(in.Catatan)) > 1000 {
		return utils.NewErrInvalidInput("Catatan observasi maksimal 1.000 karakter.")
	}
	switch in.TindakanAwal {
	case "", "pantau", "konsultasi_bk", "eskalasi_segera":
	default:
		return utils.NewErrInvalidInput("Tindakan awal tidak valid.")
	}
	return nil
}

// tanggalObservasi: parse YYYY-MM-DD, tolak > hari ini (WIB).
func tanggalObservasi(s string) (pgtype.Date, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	t, err := time.ParseInLocation("2006-01-02", s, loc)
	if err != nil {
		return pgtype.Date{}, utils.NewErrInvalidInput("Tanggal observasi harus format YYYY-MM-DD.")
	}
	if t.After(time.Now().In(loc).AddDate(0, 0, 1)) {
		return pgtype.Date{}, utils.NewErrInvalidInput("Tanggal observasi tidak boleh di masa depan.")
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func (o *Observasi) Create(ctx context.Context, actor Actor, studentID uuid.UUID, in ObservationInput) (ObservationView, error) {
	if actor.Role != middleware.RoleWaliKelas {
		return ObservationView{}, utils.NewErrForbidden("Hanya Wali Kelas yang bisa mencatat observasi.")
	}
	if _, err := o.ensureStudentScope(ctx, actor, studentID); err != nil {
		return ObservationView{}, err
	}
	if err := validateObservation(in); err != nil {
		return ObservationView{}, err
	}
	items, cats, redFlag, err := ValidateItems(in.Items)
	if err != nil {
		return ObservationView{}, err
	}
	tgl, err := tanggalObservasi(in.Tanggal)
	if err != nil {
		return ObservationView{}, err
	}
	itemsJSON, _ := json.Marshal(items)
	catsJSON, _ := json.Marshal(cats)
	obs, err := o.queries.CreateObservation(ctx, db.CreateObservationParams{
		StudentID: studentID, AuthorID: actor.UserID,
		Items: itemsJSON, Categories: catsJSON, RedFlag: redFlag,
		Catatan: optStr(in.Catatan), Tanggal: tgl, TindakanAwal: optStr(in.TindakanAwal),
	})
	if err != nil {
		return ObservationView{}, utils.NewErrInternal("gagal menyimpan observasi")
	}
	return toObservationView(obs, ""), nil
}

func (o *Observasi) Update(ctx context.Context, actor Actor, obsID uuid.UUID, in ObservationInput) (ObservationView, error) {
	if actor.Role != middleware.RoleWaliKelas {
		return ObservationView{}, utils.NewErrForbidden("Hanya Wali Kelas yang bisa mengubah observasi.")
	}
	old, _, err := o.studentOf(ctx, actor, obsID)
	if err != nil {
		return ObservationView{}, err
	}
	if err := validateObservation(in); err != nil {
		return ObservationView{}, err
	}
	items, cats, redFlag, err := ValidateItems(in.Items)
	if err != nil {
		return ObservationView{}, err
	}
	tgl, err := tanggalObservasi(in.Tanggal)
	if err != nil {
		return ObservationView{}, err
	}
	// simpan snapshot SEBELUM edit (spec §4.2: revisi = nilai lama)
	snapshot := map[string]any{
		"items": json.RawMessage(old.Items), "categories": json.RawMessage(old.Categories),
		"red_flag": old.RedFlag, "catatan": old.Catatan, "tanggal": old.Tanggal.Time.Format("2006-01-02"),
		"tindakan_awal": old.TindakanAwal,
	}
	snapJSON, _ := json.Marshal(snapshot)
	if _, err := o.queries.CreateObservationRevision(ctx, db.CreateObservationRevisionParams{
		ObservationID: obsID, EditedBy: actor.UserID, Snapshot: snapJSON,
	}); err != nil {
		return ObservationView{}, utils.NewErrInternal("gagal menyimpan revisi")
	}
	itemsJSON, _ := json.Marshal(items)
	catsJSON, _ := json.Marshal(cats)
	obs, err := o.queries.UpdateObservation(ctx, db.UpdateObservationParams{
		ID: obsID, Items: itemsJSON, Categories: catsJSON, RedFlag: redFlag,
		Catatan: optStr(in.Catatan), Tanggal: tgl, TindakanAwal: optStr(in.TindakanAwal),
	})
	if err != nil {
		return ObservationView{}, utils.NewErrInternal("gagal memperbarui observasi")
	}
	return toObservationView(obs, ""), nil
}

func (o *Observasi) Revisions(ctx context.Context, actor Actor, obsID uuid.UUID) ([]RevisionView, error) {
	if _, _, err := o.studentOf(ctx, actor, obsID); err != nil {
		return nil, err
	}
	rows, err := o.queries.ListObservationRevisions(ctx, obsID)
	if err != nil {
		return nil, utils.NewErrInternal("gagal memuat riwayat")
	}
	out := make([]RevisionView, 0, len(rows))
	for _, r := range rows {
		out = append(out, RevisionView{
			ID: r.ID.String(), EditedBy: r.EditedBy.String(),
			Snapshot: json.RawMessage(r.Snapshot), CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (o *Observasi) Handle(ctx context.Context, actor Actor, obsID uuid.UUID) error {
	if actor.Role != middleware.RoleBK {
		return utils.NewErrForbidden("Hanya Guru BK yang bisa menandai red flag ditangani.")
	}
	obs, err := o.queries.GetObservationByID(ctx, obsID)
	if err != nil {
		return utils.NewErrNotFound("Observasi tidak ditemukan.")
	}
	if !obs.RedFlag {
		return utils.NewErrInvalidInput("Observasi ini bukan red flag.")
	}
	if obs.HandledAt != nil {
		return utils.NewErrInvalidInput("Red flag ini sudah ditandai ditangani.")
	}
	if _, err := o.queries.HandleRedFlag(ctx, db.HandleRedFlagParams{ID: obsID, HandledBy: &actor.UserID}); err != nil {
		return utils.NewErrInternal("gagal menandai red flag")
	}
	return nil
}

func (o *Observasi) RedFlags(ctx context.Context, actor Actor, limit, offset int32) ([]RedFlagItem, int64, error) {
	kelas := ""
	if actor.Role != middleware.RoleBK {
		kelas = actor.Kelas
	}
	rows, err := o.queries.ListActiveRedFlagsScoped(ctx, db.ListActiveRedFlagsScopedParams{
		Kelas: kelas, Lim: limit, Off: offset,
	})
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat red flag")
	}
	total, err := o.queries.CountActiveRedFlagsScoped(ctx, kelas)
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat red flag")
	}
	out := make([]RedFlagItem, 0, len(rows))
	for _, r := range rows {
		var items []string
		_ = json.Unmarshal(r.Items, &items)
		var redOnly []string
		for _, it := range items {
			if strings.HasPrefix(it, "red_flag:") {
				redOnly = append(redOnly, it)
			}
		}
		out = append(out, RedFlagItem{
			StudentID: r.StudentID.String(), StudentName: r.StudentName, Kelas: r.Kelas,
			ObservationID: r.ID.String(), Tanggal: r.Tanggal.Time.Format("2006-01-02"),
			RedFlagItems: redOnly,
		})
	}
	return out, total, nil
}
