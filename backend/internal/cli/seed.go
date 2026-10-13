package cli

import (
	"context"
	"fmt"

	"backend/internal/config"
	"backend/internal/db"
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

	siswa := []struct{ nis, name string }{
		{"2026001", "Andi Pratama"},
		{"2026002", "Bunga Lestari"},
		{"2026003", "Citra Dewi"},
		{"2026004", "Dimas Saputra"},
	}
	for _, s := range siswa {
		nis := s.nis
		st, err := q.CreateStudent(ctx, db.CreateStudentParams{Nis: &nis, Name: s.name, Kelas: kelas})
		if err != nil {
			st, err = q.GetStudentByNIS(ctx, &nis)
			if err != nil {
				return err
			}
		}
		if err := q.CreateStudentProfileEmpty(ctx, st.ID); err != nil {
			return err
		}
	}

	fmt.Println("OK — seed demo selesai (akun: bk@demo.local / wali@demo.local, password: password123)")
	return nil
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
