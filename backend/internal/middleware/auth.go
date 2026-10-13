package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"backend/internal/tokens"
	"backend/internal/utils"
)

type ctxKey int

const userIDKey ctxKey = iota

// RequireAuth rejects requests without a valid auth token, lalu mengisi
// context dengan userID + role + kelas (untuk RBAC).
// Token sources (in order): `auth_token` httpOnly cookie, then
// `Authorization: Bearer` header (fallback exists so curl/Postman demos work).
func RequireAuth(tm *tokens.Manager) func(http.Handler) http.Handler {
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

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = SetAuthContext(ctx, claims.Role, claims.Kelas)
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
