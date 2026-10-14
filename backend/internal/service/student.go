package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/utils"
)

// Actor = identitas pemanggil (dari JWT) untuk scoping RBAC.
type Actor struct {
	UserID uuid.UUID
	Role   string
	Kelas  string
}

type Student struct {
	queries *db.Queries
}

func NewStudent(q *db.Queries) *Student { return &Student{queries: q} }

type StudentItem struct {
	ID         string  `json:"id"`
	NISN       string  `json:"nisn"`
	Name       string  `json:"name"`
	Kelas      string  `json:"kelas"`
	ArchivedAt *string `json:"archived_at"` // M11: selalu ada (null bila aktif)
	HasRedFlag bool    `json:"has_red_flag"`
}

type CreateStudentInput struct {
	NISN  string `json:"nisn"`
	Name  string `json:"name"`
	Kelas string `json:"kelas"`
}

type UpdateStudentInput struct {
	NISN  string `json:"nisn"`
	Name  string `json:"name"`
	Kelas string `json:"kelas"`
}

type ProfileInput struct {
	NilaiErapor          []EraporRow `json:"nilai_erapor"`
	Sakit                int         `json:"sakit"`
	Izin                 int         `json:"izin"`
	Alpa                 int         `json:"alpa"`
	MinatBakat           string      `json:"minat_bakat"`
	TanggalLahir         string      `json:"tanggal_lahir"` // YYYY-MM-DD
	CatatanPsikologis    string      `json:"catatan_psikologis"`
	CatatanKasusPrestasi string      `json:"catatan_kasus_prestasi"`
	MasukanBK            string      `json:"masukan_bk"`
	PendapatWaliKelas    string      `json:"pendapat_wali_kelas"`
}

type EraporRow struct {
	Mapel   string  `json:"mapel"`
	Nilai   float64 `json:"nilai"`
	Periode string  `json:"periode"`
}

type ProfileView struct {
	NilaiErapor          []EraporRow `json:"nilai_erapor"`
	Sakit                int         `json:"sakit"`
	Izin                 int         `json:"izin"`
	Alpa                 int         `json:"alpa"`
	MinatBakat           string      `json:"minat_bakat,omitempty"`
	TanggalLahir         string      `json:"tanggal_lahir,omitempty"`
	CatatanPsikologis    string      `json:"catatan_psikologis,omitempty"`
	CatatanKasusPrestasi string      `json:"catatan_kasus_prestasi,omitempty"`
	MasukanBK            string      `json:"masukan_bk,omitempty"`
	PendapatWaliKelas    string      `json:"pendapat_wali_kelas,omitempty"`
	UpdatedAt            string      `json:"updated_at"`
}

// ObservationAuthor: bentuk objek sesuai spec §21.4 (`author: {name}`).
type ObservationAuthor struct {
	Name string `json:"name"`
}

type ObservationView struct {
	ID           string            `json:"id"`
	StudentID    string            `json:"student_id"`
	Items        []string          `json:"items"`
	Categories   []string          `json:"categories"`
	RedFlag      bool              `json:"red_flag"`
	HandledAt    *string           `json:"handled_at"` // M11: selalu ada (null bila belum)
	Catatan      string            `json:"catatan,omitempty"`
	Tanggal      string            `json:"tanggal"`
	TindakanAwal string            `json:"tindakan_awal,omitempty"`
	Author       ObservationAuthor `json:"author"`
}

// StudentDetail: `latest_summary` diisi oleh handler (compose dengan summary
// service, Task 7) — service student tidak tahu soal ringkasan.
type StudentDetail struct {
	Student      StudentItem       `json:"student"`
	Profile      ProfileView       `json:"profile"`
	Observations []ObservationView `json:"observations"`
}

// isBK = akses semua kelas.
func isBK(actor Actor) bool { return actor.Role == middleware.RoleBK }

// scopeKelas: Wali Kelas → kelasnya; BK → "" (tanpa filter).
// M1: non-BK dengan kelas kosong ditolak (fail-closed) — bukan berarti "semua kelas".
func scopeKelas(actor Actor) string {
	if isBK(actor) {
		return ""
	}
	return actor.Kelas
}

// requireScopeClass: fail-closed untuk non-BK yang kelasnya kosong (audit M1).
// Mencegah Wali Kelas tanpa kelas melihat SELURUH sekolah (query `kelas=''` = tanpa filter).
func requireScopeClass(actor Actor) error {
	if !isBK(actor) && strings.TrimSpace(actor.Kelas) == "" {
		return utils.NewErrForbidden("Akunmu belum punya kelas. Hubungi administrator.")
	}
	return nil
}

// ensureWritable: hanya Wali Kelas yang boleh menulis Data Siswa.
func ensureWritable(actor Actor) error {
	if actor.Role != middleware.RoleWaliKelas {
		return utils.NewErrForbidden("Hanya Wali Kelas yang bisa mengubah data siswa.")
	}
	return nil
}

// ensureScope: Wali Kelas hanya boleh menyentuh siswa di kelasnya.
func (s *Student) ensureScope(ctx context.Context, actor Actor, id uuid.UUID) (db.Student, error) {
	st, err := s.queries.GetStudentByID(ctx, id)
	if err != nil {
		return db.Student{}, utils.NewErrNotFound("Siswa tidak ditemukan.")
	}
	if !isBK(actor) && st.Kelas != actor.Kelas {
		return db.Student{}, utils.NewErrOutOfScope("Siswa ini di luar cakupanmu.")
	}
	return st, nil
}

// validateNISN: wajib digit, panjang 4–20 (spec §10.3).
func validateNISN(nisn string) error {
	if len(nisn) < 4 || len(nisn) > 20 {
		return utils.NewErrInvalidInput("NISN harus 4–20 digit.")
	}
	for _, r := range nisn {
		if r < '0' || r > '9' {
			return utils.NewErrInvalidInput("NISN hanya boleh angka.")
		}
	}
	return nil
}

func validateNameKelas(name, kelas string) error {
	if strings.TrimSpace(name) == "" || len(name) > 80 {
		return utils.NewErrInvalidInput("Nama siswa wajib diisi, maksimal 80 karakter.")
	}
	if strings.TrimSpace(kelas) == "" || len(kelas) > 40 {
		return utils.NewErrInvalidInput("Kelas wajib diisi, maksimal 40 karakter.")
	}
	return nil
}

func toStudentItem(st db.Student, hasRedFlag bool) StudentItem {
	item := StudentItem{ID: st.ID.String(), Name: st.Name, Kelas: st.Kelas, HasRedFlag: hasRedFlag}
	if st.Nisn != nil {
		item.NISN = *st.Nisn
	}
	if st.ArchivedAt != nil {
		s := st.ArchivedAt.Format(time.RFC3339)
		item.ArchivedAt = &s
	}
	return item
}

func (s *Student) List(ctx context.Context, actor Actor, q string, includeArchived bool, limit, offset int32) ([]StudentItem, int64, error) {
	if err := requireScopeClass(actor); err != nil {
		return nil, 0, err
	}
	kelas := scopeKelas(actor)
	rows, err := s.queries.ListStudents(ctx, db.ListStudentsParams{
		Kelas: kelas, IncludeArchived: includeArchived, Search: strings.TrimSpace(q), Lim: limit, Off: offset,
	})
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat daftar siswa")
	}
	total, err := s.queries.CountStudents(ctx, db.CountStudentsParams{
		Kelas: kelas, IncludeArchived: includeArchived, Search: strings.TrimSpace(q),
	})
	if err != nil {
		return nil, 0, utils.NewErrInternal("gagal memuat daftar siswa")
	}
	out := make([]StudentItem, 0, len(rows))
	for _, st := range rows {
		red, err := s.hasRedFlag(ctx, st.ID)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, toStudentItem(st, red))
	}
	return out, total, nil
}

// hasRedFlag: siswa punya ≥1 observasi red_flag belum ditangani.
func (s *Student) hasRedFlag(ctx context.Context, studentID uuid.UUID) (bool, error) {
	obs, err := s.queries.ListObservationsByStudent(ctx, db.ListObservationsByStudentParams{
		StudentID: studentID, Limit: 200, Offset: 0,
	})
	if err != nil {
		return false, utils.NewErrInternal("gagal memeriksa red flag")
	}
	for _, o := range obs {
		if o.RedFlag && o.HandledAt == nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *Student) Create(ctx context.Context, actor Actor, in CreateStudentInput) (StudentItem, error) {
	if err := ensureWritable(actor); err != nil {
		return StudentItem{}, err
	}
	if err := validateNameKelas(in.Name, in.Kelas); err != nil {
		return StudentItem{}, err
	}
	if in.Kelas != actor.Kelas {
		return StudentItem{}, utils.NewErrOutOfScope("Kamu hanya bisa menambah siswa di kelasmu.")
	}
	var nisn *string
	if strings.TrimSpace(in.NISN) != "" {
		if err := validateNISN(in.NISN); err != nil {
			return StudentItem{}, err
		}
		nisn = &in.NISN
		if _, err := s.queries.GetStudentByNISN(ctx, nisn); err == nil {
			return StudentItem{}, utils.NewErrConflict("NISN sudah terdaftar.")
		}
	}
	st, err := s.queries.CreateStudent(ctx, db.CreateStudentParams{Nisn: nisn, Name: in.Name, Kelas: in.Kelas})
	if err != nil {
		// L9: hanya pelanggaran unik yang jadi CONFLICT; error lain → INTERNAL.
		if isUniqueViolation(err) {
			return StudentItem{}, utils.NewErrConflict("NISN sudah terdaftar.")
		}
		return StudentItem{}, utils.NewErrInternal("Gagal membuat siswa. Coba lagi.")
	}
	if err := s.queries.CreateStudentProfileEmpty(ctx, st.ID); err != nil {
		return StudentItem{}, utils.NewErrInternal("gagal membuat profil siswa")
	}
	return toStudentItem(st, false), nil
}

func (s *Student) Update(ctx context.Context, actor Actor, id uuid.UUID, in UpdateStudentInput) (StudentItem, error) {
	if err := ensureWritable(actor); err != nil {
		return StudentItem{}, err
	}
	if _, err := s.ensureScope(ctx, actor, id); err != nil {
		return StudentItem{}, err
	}
	if err := validateNameKelas(in.Name, in.Kelas); err != nil {
		return StudentItem{}, err
	}
	if in.Kelas != actor.Kelas {
		return StudentItem{}, utils.NewErrOutOfScope("Kamu hanya bisa memindahkan siswa dalam kelasmu.")
	}
	var nisn *string
	if strings.TrimSpace(in.NISN) != "" {
		if err := validateNISN(in.NISN); err != nil {
			return StudentItem{}, err
		}
		nisn = &in.NISN
	}
	st, err := s.queries.UpdateStudent(ctx, db.UpdateStudentParams{ID: id, Nisn: nisn, Name: in.Name, Kelas: in.Kelas})
	if err != nil {
		if isUniqueViolation(err) {
			return StudentItem{}, utils.NewErrConflict("NISN sudah dipakai siswa lain.")
		}
		return StudentItem{}, utils.NewErrInternal("Gagal memperbarui siswa. Coba lagi.")
	}
	red, err := s.hasRedFlag(ctx, id)
	if err != nil {
		return StudentItem{}, err
	}
	return toStudentItem(st, red), nil
}

func (s *Student) Archive(ctx context.Context, actor Actor, id uuid.UUID) error {
	if err := ensureWritable(actor); err != nil {
		return err
	}
	if _, err := s.ensureScope(ctx, actor, id); err != nil {
		return err
	}
	if _, err := s.queries.ArchiveStudent(ctx, id); err != nil {
		return utils.NewErrInternal("gagal mengarsipkan siswa")
	}
	return nil
}

func (s *Student) Restore(ctx context.Context, actor Actor, id uuid.UUID) error {
	if err := ensureWritable(actor); err != nil {
		return err
	}
	if _, err := s.ensureScope(ctx, actor, id); err != nil {
		return err
	}
	if _, err := s.queries.RestoreStudent(ctx, id); err != nil {
		return utils.NewErrInternal("gagal memulihkan siswa")
	}
	return nil
}

func (s *Student) GetProfile(ctx context.Context, actor Actor, id uuid.UUID) (ProfileView, error) {
	if _, err := s.ensureScope(ctx, actor, id); err != nil {
		return ProfileView{}, err
	}
	p, err := s.queries.GetStudentProfile(ctx, id)
	if err != nil {
		return ProfileView{}, utils.NewErrInternal("gagal memuat profil")
	}
	return toProfileView(p), nil
}

func (s *Student) SaveProfile(ctx context.Context, actor Actor, id uuid.UUID, in ProfileInput) (ProfileView, error) {
	if err := ensureWritable(actor); err != nil {
		return ProfileView{}, err
	}
	if _, err := s.ensureScope(ctx, actor, id); err != nil {
		return ProfileView{}, err
	}
	if err := validateProfile(in); err != nil {
		return ProfileView{}, err
	}
	nilaiJSON, err := marshalErapor(in.NilaiErapor)
	if err != nil {
		return ProfileView{}, utils.NewErrInvalidInput("Format nilai e-rapor tidak valid.")
	}
	var tgl *time.Time
	if in.TanggalLahir != "" {
		t, err := time.Parse("2006-01-02", in.TanggalLahir)
		if err != nil {
			return ProfileView{}, utils.NewErrInvalidInput("Format tanggal lahir harus YYYY-MM-DD.")
		}
		tgl = &t
	}
	p, err := s.queries.UpsertStudentProfile(ctx, db.UpsertStudentProfileParams{
		StudentID:            id,
		NilaiErapor:          nilaiJSON,
		Sakit:                int32(in.Sakit),
		Izin:                 int32(in.Izin),
		Alpa:                 int32(in.Alpa),
		MinatBakat:           optStr(in.MinatBakat),
		TanggalLahir:         toPgDate(tgl),
		CatatanPsikologis:    optStr(in.CatatanPsikologis),
		CatatanKasusPrestasi: optStr(in.CatatanKasusPrestasi),
		MasukanBk:            optStr(in.MasukanBK),
		PendapatWaliKelas:    optStr(in.PendapatWaliKelas),
		UpdatedBy:            &actor.UserID,
	})
	if err != nil {
		return ProfileView{}, utils.NewErrInternal("gagal menyimpan profil")
	}
	return toProfileView(p), nil
}

func (s *Student) Detail(ctx context.Context, actor Actor, id uuid.UUID) (StudentDetail, error) {
	st, err := s.ensureScope(ctx, actor, id)
	if err != nil {
		return StudentDetail{}, err
	}
	red, err := s.hasRedFlag(ctx, id)
	if err != nil {
		return StudentDetail{}, err
	}
	prof, err := s.queries.GetStudentProfile(ctx, id)
	if err != nil {
		return StudentDetail{}, utils.NewErrInternal("gagal memuat profil")
	}
	obsRows, err := s.queries.ListObservationsByStudent(ctx, db.ListObservationsByStudentParams{
		StudentID: id, Limit: 200, Offset: 0,
	})
	if err != nil {
		return StudentDetail{}, utils.NewErrInternal("gagal memuat observasi")
	}
	obs := make([]ObservationView, 0, len(obsRows))
	for _, o := range obsRows {
		obs = append(obs, toObservationView(obsFieldsFromRow(o), o.AuthorName))
	}
	return StudentDetail{
		Student:      toStudentItem(st, red),
		Profile:      toProfileView(prof),
		Observations: obs,
	}, nil
}

// --- helper konversi ---

// jakartaLoc: lokasi WIB untuk tampilan tanggal/waktu (spec §5 timezone).
func jakartaLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.UTC
	}
	return loc
}

// isUniqueViolation: error Postgres kode 23505 (unique_violation).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func optStr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func marshalErapor(rows []EraporRow) ([]byte, error) {
	if len(rows) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(rows)
}

func toPgDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func toProfileView(p db.StudentProfile) ProfileView {
	v := ProfileView{
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
		NilaiErapor: []EraporRow{},
		Sakit:       int(p.Sakit),
		Izin:        int(p.Izin),
		Alpa:        int(p.Alpa),
	}
	if len(p.NilaiErapor) > 0 {
		_ = json.Unmarshal(p.NilaiErapor, &v.NilaiErapor)
	}
	if p.MinatBakat != nil {
		v.MinatBakat = *p.MinatBakat
	}
	if p.TanggalLahir.Valid {
		v.TanggalLahir = p.TanggalLahir.Time.Format("2006-01-02")
	}
	if p.CatatanPsikologis != nil {
		v.CatatanPsikologis = *p.CatatanPsikologis
	}
	if p.CatatanKasusPrestasi != nil {
		v.CatatanKasusPrestasi = *p.CatatanKasusPrestasi
	}
	if p.MasukanBk != nil {
		v.MasukanBK = *p.MasukanBk
	}
	if p.PendapatWaliKelas != nil {
		v.PendapatWaliKelas = *p.PendapatWaliKelas
	}
	return v
}

// obsFields = field bersama db.Observation & ListObservationsByStudentRow.
type obsFields struct {
	ID           uuid.UUID
	StudentID    uuid.UUID
	Items        []byte
	Categories   []byte
	RedFlag      bool
	HandledAt    *time.Time
	Catatan      *string
	Tanggal      pgtype.Date
	TindakanAwal *string
}

func toObservationView(o obsFields, authorName string) ObservationView {
	v := ObservationView{
		ID: o.ID.String(), StudentID: o.StudentID.String(), RedFlag: o.RedFlag,
		Items: []string{}, Categories: []string{},
		Tanggal: o.Tanggal.Time.Format("2006-01-02"),
		Author:  ObservationAuthor{Name: authorName},
	}
	_ = json.Unmarshal(o.Items, &v.Items)
	_ = json.Unmarshal(o.Categories, &v.Categories)
	if o.Catatan != nil {
		v.Catatan = *o.Catatan
	}
	if o.TindakanAwal != nil {
		v.TindakanAwal = *o.TindakanAwal
	}
	if o.HandledAt != nil {
		s := o.HandledAt.Format(time.RFC3339)
		v.HandledAt = &s
	}
	return v
}

// obsFieldsFromObservation = adapter untuk baris dari Create/Update (tanpa join).
func obsFieldsFromObservation(o db.Observation) obsFields {
	return obsFields{
		ID: o.ID, StudentID: o.StudentID, Items: o.Items, Categories: o.Categories,
		RedFlag: o.RedFlag, HandledAt: o.HandledAt, Catatan: o.Catatan,
		Tanggal: o.Tanggal, TindakanAwal: o.TindakanAwal,
	}
}

// obsFieldsFromRow = adapter untuk baris dari List (dengan author_name).
func obsFieldsFromRow(o db.ListObservationsByStudentRow) obsFields {
	return obsFields{
		ID: o.ID, StudentID: o.StudentID, Items: o.Items, Categories: o.Categories,
		RedFlag: o.RedFlag, HandledAt: o.HandledAt, Catatan: o.Catatan,
		Tanggal: o.Tanggal, TindakanAwal: o.TindakanAwal,
	}
}

func validateProfile(in ProfileInput) error {
	for _, n := range []int{in.Sakit, in.Izin, in.Alpa} {
		if n < 0 {
			return utils.NewErrInvalidInput("Absensi tidak boleh negatif.")
		}
	}
	if len(in.MinatBakat) > 1000 {
		return utils.NewErrInvalidInput("Minat & bakat maksimal 1.000 karakter.")
	}
	for _, f := range []struct {
		v   string
		max int
		msg string
	}{
		{in.CatatanPsikologis, 2000, "Catatan psikologis maksimal 2.000 karakter."},
		{in.CatatanKasusPrestasi, 2000, "Catatan kasus/prestasi maksimal 2.000 karakter."},
		{in.MasukanBK, 2000, "Masukan BK maksimal 2.000 karakter."},
		{in.PendapatWaliKelas, 2000, "Pendapat Wali Kelas maksimal 2.000 karakter."},
	} {
		if len([]rune(f.v)) > f.max {
			return utils.NewErrInvalidInput(f.msg)
		}
	}
	for _, r := range in.NilaiErapor {
		if strings.TrimSpace(r.Mapel) == "" {
			return utils.NewErrInvalidInput("Setiap baris nilai harus punya nama mapel.")
		}
		if r.Nilai < 0 || r.Nilai > 100 {
			return utils.NewErrInvalidInput("Nilai harus antara 0 dan 100.")
		}
	}
	return nil
}
