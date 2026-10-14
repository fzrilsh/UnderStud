package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"backend/internal/db"
	"backend/internal/tokens"
	"backend/internal/utils"
)

type ctxKey int

const userIDKey ctxKey = iota

// RequireAuth rejects requests without a valid auth token, lalu mengisi
// context dengan userID + role + kelas (untuk RBAC).
// Token sources (in order): `auth_token` httpOnly cookie, then
// `Authorization: Bearer` header (fallback exists so curl/Postman demos work).
func RequireAuth(tm *tokens.Manager, q *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := ""
			if c, err := r.Cookie("auth_token"); err == nil && c.Value != "" {
				tokenStr = c.Value
			} else if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				tokenStr = strings.TrimPrefix(h, "Bearer ")
			}

			if tokenStr == "" {
				utils.WriteAppError(w, utils.NewErrUnauthorized("Sesi tidak valid. Silakan masuk kembali."))
				return
			}
			claims, err := tm.ParseClaims(tokenStr)
			if errors.Is(err, tokens.ErrExpired) {
				utils.WriteAppError(w, utils.NewErrTokenExpired("Sesi berakhir. Silakan masuk kembali."))
				return
			}
			if err != nil {
				utils.WriteAppError(w, utils.NewErrUnauthorized("Sesi tidak valid. Silakan masuk kembali."))
				return
			}

			// M2 + RBAC#4: cek ulang ke DB — akun nonaktif langsung ditolak, dan
			// role/kelas selalu fresh (perubahan kelas/nonaktif tak tunggu token expired).
			uid, perr := uuid.Parse(claims.UserID)
			if perr != nil {
				utils.WriteAppError(w, utils.NewErrUnauthorized("Sesi tidak valid. Silakan masuk kembali."))
				return
			}
			u, uerr := q.GetUserByID(r.Context(), uid)
			if uerr != nil {
				utils.WriteAppError(w, utils.NewErrUnauthorized("Sesi tidak valid. Silakan masuk kembali."))
				return
			}
			if !u.IsActive {
				utils.WriteAppError(w, utils.NewErrAccountInactive("Akun belum aktif. Hubungi administrator sekolah."))
				return
			}
			kelas := ""
			if u.Kelas != nil {
				kelas = *u.Kelas
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = SetAuthContext(ctx, u.Role, kelas)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFrom extracts the authenticated user ID from request context.
func UserIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}
