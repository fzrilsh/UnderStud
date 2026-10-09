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

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
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
	CreatedAt string `json:"created_at"`
}

func (s *Auth) Register(ctx context.Context, in RegisterInput) (User, string, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Name = strings.TrimSpace(in.Name)
	if in.Email == "" || !strings.Contains(in.Email, "@") {
		return User{}, "", utils.NewErrInvalidInput("valid email is required")
	}
	if len(in.Password) < 6 {
		return User{}, "", utils.NewErrInvalidInput("password must be at least 6 characters")
	}
	if in.Name == "" {
		return User{}, "", utils.NewErrInvalidInput("name is required")
	}

	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return User{}, "", utils.NewErrInternal("failed to hash password")
	}

	u, err := s.queries.CreateUser(ctx, db.CreateUserParams{Email: in.Email, PasswordHash: hash, Name: in.Name})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return User{}, "", utils.NewErrConflict("email already registered")
		}
		return User{}, "", utils.NewErrInternal("failed to create user")
	}

	token, err := s.tm.Make(u.ID.String())
	if err != nil {
		return User{}, "", utils.NewErrInternal("failed to issue token")
	}
	return toPublicUser(u), token, nil
}

func (s *Auth) Login(ctx context.Context, in LoginInput) (User, string, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	u, err := s.queries.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return User{}, "", utils.NewErrUnauthorized("invalid email or password")
	}
	if !utils.CheckPassword(u.PasswordHash, in.Password) {
		return User{}, "", utils.NewErrUnauthorized("invalid email or password")
	}
	token, err := s.tm.Make(u.ID.String())
	if err != nil {
		return User{}, "", utils.NewErrInternal("failed to issue token")
	}
	return toPublicUser(u), token, nil
}

func (s *Auth) Me(ctx context.Context, userID string) (User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return User{}, utils.NewErrUnauthorized("invalid user id in token")
	}
	u, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return User{}, utils.NewErrNotFound("user not found")
	}
	return toPublicUser(u), nil
}

func toPublicUser(u db.User) User {
	return User{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}
