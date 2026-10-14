package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Health handles GET /healthz — reports API + database status.
// L12: dibungkus envelope {data} agar konsisten dengan §21.1.
func Health(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"data": map[string]string{"status": "unhealthy", "db": "down"},
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]string{"status": "ok", "db": "up"},
		})
	}
}
