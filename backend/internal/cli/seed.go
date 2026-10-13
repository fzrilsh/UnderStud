package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/service"
	"backend/internal/utils"
)

// SeedDemo mengisi data demo (idempotent — aman dijalankan berulang).
func SeedDemo(cfg config.Config) error {
	pool, q, err := connect(cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	ctx := context.Background()

	if err := upsertStaff(ctx, q, "bk@demo.local", "password123", "Rina Pratiwi", "BK", nil); err != nil {
		return err
	}
	kelas := "XI-IPA-1"
	if err := upsertStaff(ctx, q, "wali@demo.local", "password123", "Budi Santoso", "WALI_KELAS", &kelas); err != nil {
		return err
	}
	wali, err := q.GetUserByEmail(ctx, "wali@demo.local")
	if err != nil {
		return err
	}

	type seedObs struct {
		items   []string
		catatan string
	}
	siswa := []struct {
		nis   string
		name  string
		nilai []map[string]any
		minat string
		psik  string
		obs   []seedObs
	}{
		{"2026001", "Andi Pratama", erapor5(), "Basket", "", []seedObs{
			{[]string{"emosi:murung", "emosi:cemas_berlebihan"}, "Terlihat murung 2 minggu terakhir."},
		}},
		{"2026002", "Bunga Lestari", erapor5(), "Menulis", "", []seedObs{
			{[]string{"red_flag:menyakiti_diri"}, "Menulis di buku: 'aku capek hidup'."},
		}},
		{"2026003", "Citra Dewi", erapor5(), "", "", []seedObs{
			{[]string{"akademik:nilai_turun"}, "Nilai matematika turun drastis."},
			{[]string{"kehadiran:terlambat_bolos"}, "Bolos 3x minggu ini."},
		}},
		{"2026004", "Dimas Saputra", erapor5(), "", "Sedang menghadapi perceraian orang tua.", nil},
	}

	for _, s := range siswa {
		nis := s.nis
		st, err := q.GetStudentByNIS(ctx, &nis)
		if err != nil {
			st, err = q.CreateStudent(ctx, db.CreateStudentParams{Nis: &nis, Name: s.name, Kelas: kelas})
			if err != nil {
				return err
			}
			if err := q.CreateStudentProfileEmpty(ctx, st.ID); err != nil {
				return err
			}
		}
		if err := seedProfile(ctx, q, st.ID, wali.ID, s.nilai, s.minat, s.psik); err != nil {
			return err
		}
		existing, err := q.CountObservationsByStudent(ctx, st.ID)
		if err != nil {
			return err
		}
		if existing > 0 {
			continue
		}
		for _, o := range s.obs {
			items, cats, redFlag, err := service.ValidateItems(o.items)
			if err != nil {
				return err
			}
			itemsJSON, _ := json.Marshal(items)
			catsJSON, _ := json.Marshal(cats)
			var catatan *string
			if o.catatan != "" {
				catatan = &o.catatan
			}
			if _, err := q.CreateObservation(ctx, db.CreateObservationParams{
				StudentID: st.ID, AuthorID: wali.ID,
				Items: itemsJSON, Categories: catsJSON, RedFlag: redFlag,
				Catatan: catatan, Tanggal: pgtype.Date{Time: time.Now(), Valid: true},
			}); err != nil {
				return err
			}
		}
	}

	fmt.Println("OK — seed demo selesai (akun: bk@demo.local / wali@demo.local, password: password123)")
	return nil
}

func erapor5() []map[string]any {
	return []map[string]any{
		{"mapel": "Matematika", "nilai": 85, "periode": "2026-Ganjil"},
		{"mapel": "Fisika", "nilai": 78, "periode": "2026-Ganjil"},
		{"mapel": "Kimia", "nilai": 82, "periode": "2026-Ganjil"},
		{"mapel": "Biologi", "nilai": 88, "periode": "2026-Ganjil"},
		{"mapel": "Bahasa Indonesia", "nilai": 90, "periode": "2026-Ganjil"},
	}
}

func seedProfile(ctx context.Context, q *db.Queries, studentID, waliID uuid.UUID, nilai []map[string]any, minat, psik string) error {
	nilaiJSON, _ := json.Marshal(nilai)
	params := db.UpsertStudentProfileParams{StudentID: studentID, NilaiErapor: nilaiJSON, UpdatedBy: &waliID}
	if minat != "" {
		params.MinatBakat = &minat
	}
	if psik != "" {
		params.CatatanPsikologis = &psik
	}
	_, err := q.UpsertStudentProfile(ctx, params)
	return err
}

func upsertStaff(ctx context.Context, q *db.Queries, email, pass, name, role string, kelas *string) error {
	_, err := q.GetUserByEmail(ctx, email)
	if err == nil {
		return nil // sudah ada
	}
	hash, err := utils.HashPassword(pass)
	if err != nil {
		return err
	}
	_, err = q.CreateStaff(ctx, db.CreateStaffParams{
		Email: email, PasswordHash: hash, Name: name, Role: role, Kelas: kelas,
	})
	return err
}
