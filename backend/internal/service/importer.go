package service

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/ai"
	"backend/internal/db"
	"backend/internal/leger"
	"backend/internal/middleware"
	"backend/internal/utils"
)

const maxImportBytes = 2 * 1024 * 1024 // 2 MB (spec §10.5)
const previewTTL = 10 * time.Minute     // spec §10.3.1

type Importer struct {
	queries *db.Queries
	queue   *ai.Queue
}

func NewImporter(q *db.Queries, queue *ai.Queue) *Importer {
	return &Importer{queries: q, queue: queue}
}

type ImportRow struct {
	NISN  string      `json:"nisn"`
	Nama  string      `json:"nama"`
	Kelas string      `json:"kelas"`
	Nilai []EraporRow `json:"nilai_erapor"`
	Sakit int         `json:"sakit"`
	Izin  int         `json:"izin"`
	Alpa  int         `json:"alpa"`
}

type ImportError struct {
	Baris int    `json:"baris"`
	NISN  string `json:"nisn,omitempty"`
	Pesan string `json:"pesan"`
}

type ImportPreview struct {
	PreviewToken string        `json:"preview_token"`
	Baru         []ImportRow   `json:"baru"`
	Update       []ImportRow   `json:"update"`
	Errors       []ImportError `json:"errors"`
	Ringkasan    ImportSummary `json:"ringkasan"`
}

type ImportSummary struct {
	Total  int `json:"total"`
	Baru   int `json:"baru"`
	Update int `json:"update"`
	Error  int `json:"error"`
}

type ImportResult struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Errors  int `json:"errors"`
}

// detectFormat: dari ekstensi, fallback magic bytes (zip = xlsx).
func detectFormat(filename string, head []byte) string {
	lower := strings.ToLower(strings.TrimSpace(filename))
	switch {
	case strings.HasSuffix(lower, ".xlsx"):
		return "xlsx"
	case strings.HasSuffix(lower, ".csv"):
		return "csv"
	}
	if len(head) >= 4 && head[0] == 'P' && head[1] == 'K' && head[2] == 0x03 && head[3] == 0x04 {
		return "xlsx"
	}
	return ""
}

// toGrid: baca bytes jadi leger.Grid sesuai format.
func toGrid(format string, data []byte) (leger.Grid, error) {
	switch format {
	case "xlsx":
		return leger.GridFromXLSX(bytes.NewReader(data))
	case "csv":
		return leger.GridFromCSV(bytes.NewReader(data))
	}
	return leger.Grid{}, utils.NewErrUnsupportedMedia("Hanya file .xlsx atau .csv.")
}

func (i *Importer) Preview(ctx context.Context, actor Actor, filename string, data []byte) (ImportPreview, error) {
	if actor.Role != middleware.RoleWaliKelas {
		return ImportPreview{}, utils.NewErrForbidden("Hanya Wali Kelas yang bisa mengimpor data siswa.")
	}
	if len(data) == 0 {
		return ImportPreview{}, utils.NewErrInvalidInput("File kosong.")
	}
	if len(data) > maxImportBytes {
		return ImportPreview{}, utils.NewErrPayloadTooLarge("File maksimal 2 MB.")
	}
	format := detectFormat(filename, data)
	if format == "" {
		return ImportPreview{}, utils.NewErrUnsupportedMedia("Hanya file .xlsx atau .csv.")
	}
	grid, err := toGrid(format, data)
	if err != nil {
		return ImportPreview{}, utils.NewErrInvalidInput("File tidak dapat dibaca: " + err.Error())
	}
	lg, err := leger.Parse(grid)
	if err != nil {
		return ImportPreview{}, utils.NewErrInvalidInput(err.Error())
	}

	// Kelas metadata file harus cocok kelas Wali Kelas (spec §10.3).
	if lg.Kelas == "" {
		return ImportPreview{}, utils.NewErrInvalidInput("Kelas tidak terbaca dari file leger.")
	}
	if lg.Kelas != actor.Kelas {
		return ImportPreview{}, utils.NewErrInvalidInput("File untuk kelas " + lg.Kelas + ", kamu Wali Kelas " + actor.Kelas + ".")
	}

	var baru, upd []ImportRow
	var errs []ImportError
	for idx, r := range lg.Rows {
		line := idx + 1
		if err := validateNISN(r.NISN); err != nil {
			errs = append(errs, ImportError{Baris: line, NISN: r.NISN, Pesan: "NISN tidak valid (4–20 digit)."})
			continue
		}
		if strings.TrimSpace(r.Nama) == "" {
			errs = append(errs, ImportError{Baris: line, NISN: r.NISN, Pesan: "Nama wajib diisi."})
			continue
		}
		row := ImportRow{
			NISN: r.NISN, Nama: r.Nama, Kelas: actor.Kelas,
			Nilai: make([]EraporRow, 0, len(r.Nilai)),
			Sakit: r.Sakit, Izin: r.Izin, Alpa: r.Alpa,
		}
		for _, n := range r.Nilai {
			row.Nilai = append(row.Nilai, EraporRow{Mapel: n.Mapel, Nilai: n.Nilai, Periode: lg.Periode})
		}
		nisn := r.NISN
		existing, err := i.queries.GetStudentByNISN(ctx, &nisn)
		if err != nil {
			baru = append(baru, row)
			continue
		}
		if existing.ArchivedAt != nil {
			errs = append(errs, ImportError{Baris: line, NISN: r.NISN, Pesan: "NISN " + r.NISN + " ada di data arsip. Pulihkan siswa dulu."})
			continue
		}
		if existing.Kelas != actor.Kelas {
			errs = append(errs, ImportError{Baris: line, NISN: r.NISN, Pesan: "NISN " + r.NISN + " terdaftar di kelas " + existing.Kelas + ", file menyebut " + actor.Kelas + "."})
			continue
		}
		upd = append(upd, row)
	}

	token := uuid.NewString()
	payload, _ := json.Marshal(map[string]any{"baru": baru, "update": upd, "errors": errs})
	if _, err := i.queries.CreateImportPreview(ctx, db.CreateImportPreviewParams{
		Token: token, CreatedBy: actor.UserID, Payload: payload,
		ExpiresAt: time.Now().Add(previewTTL),
	}); err != nil {
		return ImportPreview{}, utils.NewErrInternal("gagal menyimpan preview")
	}

	if baru == nil {
		baru = []ImportRow{}
	}
	if upd == nil {
		upd = []ImportRow{}
	}
	if errs == nil {
		errs = []ImportError{}
	}
	return ImportPreview{
		PreviewToken: token, Baru: baru, Update: upd, Errors: errs,
		Ringkasan: ImportSummary{Total: len(baru) + len(upd) + len(errs), Baru: len(baru), Update: len(upd), Error: len(errs)},
	}, nil
}

func (i *Importer) Commit(ctx context.Context, actor Actor, token string) (ImportResult, error) {
	if actor.Role != middleware.RoleWaliKelas {
		return ImportResult{}, utils.NewErrForbidden("Hanya Wali Kelas yang bisa mengimpor data siswa.")
	}
	prev, err := i.queries.GetImportPreview(ctx, token)
	if err != nil {
		return ImportResult{}, utils.NewErrPreviewExpired("Preview kedaluwarsa. Upload ulang file.")
	}
	if prev.CreatedBy != actor.UserID {
		return ImportResult{}, utils.NewErrForbidden("Preview ini bukan milikmu.")
	}
	var payload struct {
		Baru   []ImportRow   `json:"baru"`
		Update []ImportRow   `json:"update"`
		Errors []ImportError `json:"errors"`
	}
	if err := json.Unmarshal(prev.Payload, &payload); err != nil {
		return ImportResult{}, utils.NewErrInternal("payload preview rusak")
	}

	res := ImportResult{Errors: len(payload.Errors)}
	for _, row := range payload.Baru {
		nisn := row.NISN
		st, err := i.queries.CreateStudent(ctx, db.CreateStudentParams{Nisn: &nisn, Name: row.Nama, Kelas: row.Kelas})
		if err != nil {
			continue
		}
		_ = i.queries.CreateStudentProfileEmpty(ctx, st.ID)
		if err := i.writeProfile(ctx, actor, st.ID, row); err != nil {
			return ImportResult{}, err
		}
		res.Created++
		i.queue.Enqueue(st.ID, actor.UserID, "profile_updated")
	}
	for _, row := range payload.Update {
		nisn := row.NISN
		st, err := i.queries.GetStudentByNISN(ctx, &nisn)
		if err != nil {
			continue
		}
		if err := i.writeProfile(ctx, actor, st.ID, row); err != nil {
			return ImportResult{}, err
		}
		res.Updated++
		i.queue.Enqueue(st.ID, actor.UserID, "profile_updated")
	}
	return res, nil
}

// writeProfile: gabung nilai baru ke nilai_erapor (tak hapus mapel lain) +
// timpa absensi. Spec §10.3.
func (i *Importer) writeProfile(ctx context.Context, actor Actor, studentID uuid.UUID, row ImportRow) error {
	prof, err := i.queries.GetStudentProfile(ctx, studentID)
	if err != nil {
		return utils.NewErrInternal("gagal memuat profil")
	}
	merged := map[string]EraporRow{}
	var order []string
	if len(prof.NilaiErapor) > 0 {
		var existing []EraporRow
		if json.Unmarshal(prof.NilaiErapor, &existing) == nil {
			for _, e := range existing {
				key := e.Mapel + "|" + e.Periode
				merged[key] = e
				order = append(order, key)
			}
		}
	}
	for _, n := range row.Nilai {
		key := n.Mapel + "|" + n.Periode
		if _, ok := merged[key]; !ok {
			order = append(order, key)
		}
		merged[key] = n
	}
	out := make([]EraporRow, 0, len(order))
	for _, k := range order {
		out = append(out, merged[k])
	}
	nilaiJSON, _ := json.Marshal(out)
	_, err = i.queries.UpsertStudentProfile(ctx, db.UpsertStudentProfileParams{
		StudentID:            studentID,
		NilaiErapor:          nilaiJSON,
		Sakit:                int32(row.Sakit),
		Izin:                 int32(row.Izin),
		Alpa:                 int32(row.Alpa),
		MinatBakat:           prof.MinatBakat,
		TanggalLahir:         prof.TanggalLahir,
		CatatanPsikologis:    prof.CatatanPsikologis,
		CatatanKasusPrestasi: prof.CatatanKasusPrestasi,
		MasukanBk:            prof.MasukanBk,
		PendapatWaliKelas:    prof.PendapatWaliKelas,
		UpdatedBy:            &actor.UserID,
	})
	if err != nil {
		return utils.NewErrInternal("gagal menyimpan profil")
	}
	return nil
}
