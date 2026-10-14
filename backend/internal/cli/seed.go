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
	kelas := "X TJKT 1"
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
		nisn              string
		name              string
		nilai             []map[string]any
		minat             string
		psik              string
		sakit, izin, alpa int
		obs               []seedObs
	}{
		{"2060691501", "Andi Pratama", erapor5(), "Basket", "", 0, 0, 0, []seedObs{
			{[]string{"emosi:murung", "emosi:cemas_berlebihan"}, "Terlihat murung 2 minggu terakhir."},
		}},
		{"2060691502", "Bunga Lestari", erapor5(), "Menulis", "", 0, 0, 0, []seedObs{
			{[]string{"red_flag:menyakiti_diri"}, "Menulis di buku: 'aku capek hidup'."},
		}},
		{"2060691503", "Citra Dewi", erapor5(), "", "", 3, 1, 0, []seedObs{
			{[]string{"akademik:nilai_turun"}, "Nilai matematika turun drastis."},
			{[]string{"kehadiran:terlambat_bolos"}, "Bolos 3x minggu ini."},
		}},
		{"2060691504", "Dimas Saputra", erapor5(), "", "Sedang menghadapi perceraian orang tua.", 0, 0, 0, nil},
	}

	for _, s := range siswa {
		nisn := s.nisn
		st, err := q.GetStudentByNISN(ctx, &nisn)
		if err != nil {
			st, err = q.CreateStudent(ctx, db.CreateStudentParams{Nisn: &nisn, Name: s.name, Kelas: kelas})
			if err != nil {
				return err
			}
			if err := q.CreateStudentProfileEmpty(ctx, st.ID); err != nil {
				return err
			}
		}
		if err := seedProfile(ctx, q, st.ID, wali.ID, s.nilai, s.minat, s.psik, s.sakit, s.izin, s.alpa); err != nil {
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

	if err := seedAnonChat(ctx, q); err != nil {
		return err
	}

	fmt.Println("OK — seed demo selesai (akun: bk@demo.local / wali@demo.local, password: password123)")
	return nil
}

// seedAnonChat membuat 1 chat anonim contoh (2 pesan, status baru) sesuai spec §23.
// Idempotent: ditandai lewat token_hash tetap; dilewati bila sudah ada.
func seedAnonChat(ctx context.Context, q *db.Queries) error {
	const seedTokenHash = "seed-demo-anon-token-hash-fixed-000000000000000000000000000000"
	sess, err := q.GetAnonSessionByTokenHash(ctx, seedTokenHash)
	if err != nil {
		sess, err = q.CreateAnonSession(ctx, seedTokenHash)
		if err != nil {
			return err
		}
	}
	n, err := q.CountChatsBySession(ctx, sess.ID)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	chat, err := q.CreateChat(ctx, db.CreateChatParams{
		SessionID: sess.ID, IdentityMode: "anonim", SnapshotName: nil, SnapshotKelas: nil,
	})
	if err != nil {
		return err
	}
	if _, err := q.CreateMessage(ctx, db.CreateMessageParams{
		ChatID: chat.ID, Sender: "SISWA", SenderUserID: nil,
		Body: "Saya sering merasa cemas saat ujian. Boleh minta saran?",
	}); err != nil {
		return err
	}
	if _, err := q.CreateMessage(ctx, db.CreateMessageParams{
		ChatID: chat.ID, Sender: "BK", SenderUserID: nil,
		Body: "Terima kasih sudah bercerita. Kapan kamu merasa paling cemas?",
	}); err != nil {
		return err
	}
	return nil
}

func erapor5() []map[string]any {
	return []map[string]any{
		{"mapel": "Matematika (Umum)", "nilai": 85, "periode": "2025/2026 Genap"},
		{"mapel": "Fisika", "nilai": 78, "periode": "2025/2026 Genap"},
		{"mapel": "Kimia", "nilai": 82, "periode": "2025/2026 Genap"},
		{"mapel": "Biologi", "nilai": 88, "periode": "2025/2026 Genap"},
		{"mapel": "Bahasa Indonesia", "nilai": 90, "periode": "2025/2026 Genap"},
	}
}

func seedProfile(ctx context.Context, q *db.Queries, studentID, waliID uuid.UUID, nilai []map[string]any, minat, psik string, sakit, izin, alpa int) error {
	nilaiJSON, _ := json.Marshal(nilai)
	params := db.UpsertStudentProfileParams{
		StudentID: studentID, NilaiErapor: nilaiJSON, UpdatedBy: &waliID,
		Sakit: int32(sakit), Izin: int32(izin), Alpa: int32(alpa),
	}
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
	existing, err := q.GetUserByEmail(ctx, email)
	if err == nil {
		// Sudah ada → sinkronkan role/kelas (mis. kelas seed berubah antar versi).
		_, err = q.UpdateStaff(ctx, db.UpdateStaffParams{
			Email: existing.Email, Name: name, Role: role, Kelas: kelas,
		})
		return err
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
