package middleware

import (
	"context"
	"net/http"

	"backend/internal/utils"
)

const (
	RoleBK        = "BK"
	RoleWaliKelas = "WALI_KELAS"
)

type roleCtxKey int

const (
	roleKey roleCtxKey = iota
	kelasKey
)

// SetAuthContext menyimpan role & kelas ke context (dipanggil RequireAuth).
func SetAuthContext(ctx context.Context, role, kelas string) context.Context {
	ctx = context.WithValue(ctx, roleKey, role)
	return context.WithValue(ctx, kelasKey, kelas)
}

func RoleFrom(ctx context.Context) string {
	if v, ok := ctx.Value(roleKey).(string); ok {
		return v
	}
	return ""
}

func KelasFrom(ctx context.Context) string {
	if v, ok := ctx.Value(kelasKey).(string); ok {
		return v
	}
	return ""
}

// RequireRole menolak request bila role di context tidak termasuk daftar.
// Wajib dipasang setelah RequireAuth.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := RoleFrom(r.Context())
			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			utils.WriteAppError(w, utils.NewErrForbidden("Kamu tidak punya akses ke aksi ini."))
		})
	}
}
