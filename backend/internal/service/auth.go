package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/db"
	"backend/internal/tokens"
	"backend/internal/utils"
)

// Auth contains auth business logic.
type Auth struct {
	queries *db.Queries
	tm      *tokens.Manager
}

func NewAuth(queries *db.Queries, tm *tokens.Manager) *Auth {
	return &Auth{queries: queries, tm: tm}
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User is the public shape of a user (never includes password_hash).
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Kelas     string `json:"kelas,omitempty"`
	CreatedAt string `json:"created_at"`
}

// LoginResult membawa user publik + token + info RBAC.
type LoginResult struct {
	User  User
	Token string
	Role  string
	Kelas string
}

func (s *Auth) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	u, err := s.queries.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return LoginResult{}, utils.NewErrUnauthorized("Email atau kata sandi tidak cocok.")
	}
	if !utils.CheckPassword(u.PasswordHash, in.Password) {
		return LoginResult{}, utils.NewErrUnauthorized("Email atau kata sandi tidak cocok.")
	}
	if !u.IsActive {
		return LoginResult{}, utils.NewErrAccountInactive("Akun belum aktif. Hubungi administrator sekolah.")
	}
	kelas := ""
	if u.Kelas != nil {
		kelas = *u.Kelas
	}
	token, err := s.tm.MakeClaims(u.ID.String(), u.Role, kelas)
	if err != nil {
		return LoginResult{}, utils.NewErrInternal("Gagal membuat sesi. Coba lagi.")
	}
	return LoginResult{User: toPublicUser(u), Token: token, Role: u.Role, Kelas: kelas}, nil
}

func (s *Auth) Me(ctx context.Context, userID string) (User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return User{}, utils.NewErrUnauthorized("Sesi tidak valid. Silakan masuk kembali.")
	}
	u, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return User{}, utils.NewErrUnauthorized("Sesi tidak valid. Silakan masuk kembali.")
	}
	return toPublicUser(u), nil
}

func toPublicUser(u db.User) User {
	kelas := ""
	if u.Kelas != nil {
		kelas = *u.Kelas
	}
	return User{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		Kelas:     kelas,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}
