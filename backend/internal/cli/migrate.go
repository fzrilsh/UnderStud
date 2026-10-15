package cli

import (
	"context"
	"fmt"

	"backend/internal/config"
	"backend/internal/db"
)

// Migrate menjalankan migrasi database secara manual lalu keluar.
//
// Kenapa ada: di server yang memakai connection pooler (pgbouncer) dalam mode
// transaction, migrasi yang dijalankan saat startup aplikasi bisa gagal (goose
// butuh sesi yang stabil, sedangkan pooler memindahkan koneksi antar transaksi).
// Karena itu di produksi migrasi dijalankan terpisah, langsung ke Postgres.
//
// Contoh:
//
//	cli migrate
func Migrate(cfg config.Config) error {
	ctx := context.Background()

	// Pool kecil: migrasi hanya butuh satu koneksi.
	pool, err := db.NewPool(ctx, cfg.DatabaseURL, 1, 0)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := db.Migrate(pool); err != nil {
		return err
	}
	fmt.Println("OK — migrasi database selesai.")
	return nil
}
