package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx menjalankan fn dalam satu transaksi. Bila fn mengembalikan error,
// transaksi di-rollback (tidak ada perubahan parsial). Audit F6/F8: rangkaian
// tulis multi-langkah (impor, chat, observasi) harus atomik.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(q *Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(New(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
