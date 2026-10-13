package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/ai"
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/utils"
)

const maxImportBytes = 2 * 1024 * 1024 // 2 MB (spec §10.1)
const previewTTL = 10 * time.Minute     // spec §10.3.1

type Importer struct {
	queries *db.Queries
	queue   *ai.Queue
}

func NewImporter(q *db.Queries, queue *ai.Queue) *Importer {
	return &Importer{queries: q, queue: queue}
}

type ImportRow struct {
	NIS   string      `json:"nis"`
	Nama  string      `json:"nama"`
	Kelas string      `json:"kelas"`
	Nilai []EraporRow `json:"nilai_erapor"`
}

type ImportError struct {
	Baris int    `json:"baris"`
	NIS   string `json:"nis,omitempty"`
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

// parseCSV: parse + validasi + grup per NIS. scopeKelas = kelas Wali Kelas.
func parseCSV(data []byte, scopeKelas string) ([]ImportRow, []ImportError) {
	r := csv.NewReader(strings.NewReader(string(data)))
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, []ImportError{{Baris: 1, Pesan: "File CSV kosong atau tidak valid."}}
	}
	idx := map[string]int{}
	for i, h := range header {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, need := range []string{"nis", "nama", "kelas"} {
		if _, ok := idx[need]; !ok {
			return nil, []ImportError{{Baris: 1, Pesan: "Kolom wajib tidak ditemukan: " + need}}
		}
	}

	type acc struct {
		row  ImportRow
		line int
	}
	byNIS := map[string]*acc{}
	var order []string
	var errs []ImportError

	line := 1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			errs = append(errs, ImportError{Baris: line, Pesan: "Baris tidak dapat dibaca."})
			continue
		}
		get := func(key string) string {
			if i, ok := idx[key]; ok && i < len(rec) {
				return strings.TrimSpace(rec[i])
			}
			return ""
		}
		nis := get("nis")
		nama := get("nama")
		kelas := get("kelas")
		if err := validateNIS(nis); err != nil {
			errs = append(errs, ImportError{Baris: line, NIS: nis, Pesan: "NIS tidak valid (1–30 karakter alfanumerik)."})
			continue
		}
		if nama == "" {
			errs = append(errs, ImportError{Baris: line, NIS: nis, Pesan: "Nama wajib diisi."})
			continue
		}
		if kelas != scopeKelas {
			errs = append(errs, ImportError{Baris: line, NIS: nis, Pesan: "Kelas " + kelas + " di luar cakupanmu (" + scopeKelas + ")."})
			continue
		}
		a, ok := byNIS[nis]
		if !ok {
			a = &acc{row: ImportRow{NIS: nis, Nama: nama, Kelas: kelas, Nilai: []EraporRow{}}, line: line}
			byNIS[nis] = a
			order = append(order, nis)
		}
		mapel := get("mapel")
		if mapel != "" {
			nilaiStr := get("nilai")
			nilai, err := strconv.ParseFloat(nilaiStr, 64)
			if err != nil || nilai < 0 || nilai > 100 {
				errs = append(errs, ImportError{Baris: line, NIS: nis, Pesan: "Nilai tidak valid (0–100)."})
				continue
			}
			a.row.Nilai = append(a.row.Nilai, EraporRow{Mapel: mapel, Nilai: nilai, Periode: get("periode")})
		}
	}

	rows := make([]ImportRow, 0, len(order))
	for _, nis := range order {
		rows = append(rows, byNIS[nis].row)
	}
	return rows, errs
}

func (i *Importer) Preview(ctx context.Context, actor Actor, data []byte) (ImportPreview, error) {
	if actor.Role != middleware.RoleWaliKelas {
		return ImportPreview{}, utils.NewErrForbidden("Hanya Wali Kelas yang bisa mengimpor data siswa.")
	}
	if len(data) == 0 {
		return ImportPreview{}, utils.NewErrInvalidInput("File CSV kosong.")
	}
	if len(data) > maxImportBytes {
		return ImportPreview{}, utils.NewErrPayloadTooLarge("File CSV maksimal 2 MB.")
	}
	rows, errs := parseCSV(data, actor.Kelas)

	var baru, upd []ImportRow
	for _, row := range rows {
		nis := row.NIS
		existing, err := i.queries.GetStudentByNIS(ctx, &nis)
		if err != nil {
			baru = append(baru, row)
			continue
		}
		if existing.ArchivedAt != nil {
			errs = append(errs, ImportError{NIS: nis, Pesan: "NIS " + nis + " ada di data arsip. Pulihkan siswa dulu."})
			continue
		}
		if existing.Kelas != actor.Kelas {
			errs = append(errs, ImportError{NIS: nis, Pesan: "NIS " + nis + " terdaftar di kelas " + existing.Kelas + ", file menyebut " + actor.Kelas + "."})
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
		nis := row.NIS
		st, err := i.queries.CreateStudent(ctx, db.CreateStudentParams{Nis: &nis, Name: row.Nama, Kelas: row.Kelas})
		if err != nil {
			continue
		}
		_ = i.queries.CreateStudentProfileEmpty(ctx, st.ID)
		if err := i.writeErapor(ctx, actor, st.ID, row.Nilai); err != nil {
			return ImportResult{}, err
		}
		res.Created++
		i.queue.Enqueue(st.ID, actor.UserID, "profile_updated")
	}
	for _, row := range payload.Update {
		nis := row.NIS
		st, err := i.queries.GetStudentByNIS(ctx, &nis)
		if err != nil {
			continue
		}
		if err := i.writeErapor(ctx, actor, st.ID, row.Nilai); err != nil {
			return ImportResult{}, err
		}
		res.Updated++
		i.queue.Enqueue(st.ID, actor.UserID, "profile_updated")
	}
	return res, nil
}

// writeErapor: gabung nilai baru ke profil (tanpa menghapus mapel lain).
func (i *Importer) writeErapor(ctx context.Context, actor Actor, studentID uuid.UUID, nilai []EraporRow) error {
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
	for _, n := range nilai {
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
		MinatBakat:           prof.MinatBakat,
		TanggalLahir:         prof.TanggalLahir,
		CatatanPsikologis:    prof.CatatanPsikologis,
		CatatanKasusPrestasi: prof.CatatanKasusPrestasi,
		MasukanBk:            prof.MasukanBk,
		PendapatWaliKelas:    prof.PendapatWaliKelas,
		UpdatedBy:            &actor.UserID,
	})
	if err != nil {
		return utils.NewErrInternal("gagal menyimpan nilai e-rapor")
	}
	return nil
}
