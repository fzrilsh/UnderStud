package middleware

import (
	"context"
	"net/http"
	"time"

	"backend/internal/anontoken"
	"backend/internal/db"
	"backend/internal/tokens"
	"backend/internal/utils"
)

// RejectStaff menolak request ke domain anonim bila membawa token staf yang
// VALID (audit M3, spec §6.1: staf ❌ kirim laporan anonim). Token invalid/expired
// diabaikan — supaya cookie basi tak mengunci akses /lapor.
func RejectStaff(tm *tokens.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := ""
			if c, err := r.Cookie("auth_token"); err == nil && c.Value != "" {
				tokenStr = c.Value
			} else if h := r.Header.Get("Authorization"); len(h) > 7 && h[:7] == "Bearer " {
				tokenStr = h[7:]
			}
			if tokenStr != "" {
				if _, err := tm.ParseClaims(tokenStr); err == nil {
					utils.WriteAppError(w, utils.NewErrForbidden("Kamu sedang masuk sebagai staf. Keluar dulu untuk memakai kanal anonim."))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

type anonCtxKey int

const anonSessionKey anonCtxKey = iota

// AnonCookieName adalah nama cookie identitas siswa anonim.
const AnonCookieName = "anon_token"

// ResolveAnon membaca cookie anon_token → session. Bila belum ada / invalid,
// membuat session baru + set cookie (bootstrap, spec §5 keputusan #38).
func ResolveAnon(queries *db.Queries, ttl time.Duration, secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var session db.AnonSession
			ok := false

			if c, err := r.Cookie(AnonCookieName); err == nil && c.Value != "" {
				if s, err := queries.GetAnonSessionByTokenHash(r.Context(), anontoken.Hash(c.Value)); err == nil {
					session, ok = s, true
				}
			}

			if !ok {
				raw, hash, err := anontoken.New()
				if err == nil {
					if s, err := queries.CreateAnonSession(r.Context(), hash); err == nil {
						session, ok = s, true
						http.SetCookie(w, &http.Cookie{
							Name:     AnonCookieName,
							Value:    raw,
							Path:     "/",
							MaxAge:   int(ttl.Seconds()),
							HttpOnly: true,
							SameSite: http.SameSiteLaxMode,
							Secure:   secure,
						})
					}
				}
			}

			ctx := context.WithValue(r.Context(), anonSessionKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AnonSessionFrom mengambil session anonim dari context (ok=false bila gagal bootstrap).
func AnonSessionFrom(ctx context.Context) (db.AnonSession, bool) {
	s, ok := ctx.Value(anonSessionKey).(db.AnonSession)
	return s, ok && s.ID.String() != ""
}
