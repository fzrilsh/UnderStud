package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

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
// NV-2: pembuatan sesi BARU dibatasi per-IP (cegah spammer membuang cookie →
// sesi tak terbatas). rl/trustProxyHeaders boleh nil/false (tanpa batas).
func ResolveAnon(queries *db.Queries, ttl time.Duration, secure bool, rl *RateLimiter, trustProxyHeaders bool) func(http.Handler) http.Handler {
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
				// Hanya batasi saat MEMBUAT sesi baru (bukan saat resolve cookie valid).
				if rl != nil && !rl.AllowAnonSession(ClientIP(r, trustProxyHeaders)) {
					utils.WriteAppError(w, utils.NewErrRateLimited("Terlalu banyak permintaan. Coba lagi nanti."))
					return
				}
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
// F10: cek terhadap uuid.Nil, BUKAN string kosong — uuid.Nil.String() =
// "00000000-0000-0000-0000-000000000000" (non-kosong) sehingga sesi zero-value
// yang gagal bootstrap sebelumnya lolos sebagai ok=true.
func AnonSessionFrom(ctx context.Context) (db.AnonSession, bool) {
	s, ok := ctx.Value(anonSessionKey).(db.AnonSession)
	return s, ok && s.ID != uuid.Nil
}
