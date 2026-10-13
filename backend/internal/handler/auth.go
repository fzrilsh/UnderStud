package handler

import (
	"net/http"
	"time"

	"backend/internal/middleware"
	"backend/internal/service"
)

const cookieName = "auth_token"

// AuthHandlers exposes auth endpoints. Cookie written on register/login; cleared on logout.
type AuthHandlers struct {
	auth     *service.Auth
	tokenTTL time.Duration
}

func NewAuthHandlers(auth *service.Auth, tokenTTL time.Duration) *AuthHandlers {
	return &AuthHandlers{auth: auth, tokenTTL: tokenTTL}
}

// Login handles POST /api/v1/auth/login.
// Akun staf dibuat lewat CLI (bukan endpoint publik) — lihat spec §2 #8.
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var in service.LoginInput
	if err := Bind(r, &in); err != nil {
		Error(w, err)
		return
	}
	res, err := h.auth.Login(r.Context(), in)
	if err != nil {
		Error(w, err)
		return
	}
	h.setAuthCookie(w, res.Token)
	OK(w, map[string]any{"user": res.User})
}

// Logout handles POST /api/v1/auth/logout — clears the auth cookie.
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	OK(w, map[string]any{"ok": true})
}

// Me handles GET /api/v1/auth/me (protected).
func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	user, err := h.auth.Me(r.Context(), userID)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, map[string]any{"user": user})
}

func (h *AuthHandlers) setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.tokenTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
