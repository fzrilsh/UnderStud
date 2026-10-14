package service

import (
	"context"
	"strings"

	"backend/internal/db"
	"backend/internal/utils"
)

// ErrAnonUnavailable: sesi anonim gagal dibuat (DB error).
// AppErr supaya pesan/kode konsisten (audit M9) — bukan 500 generik.
var ErrAnonUnavailable = utils.NewErrInternal("Sesi anonim tidak tersedia. Coba lagi.")

type Anon struct {
	queries *db.Queries
}

func NewAnon(queries *db.Queries) *Anon { return &Anon{queries: queries} }

type PelaporIdentity struct {
	Name  string `json:"name"`
	Kelas string `json:"kelas"`
}

func toIdentity(s db.AnonSession) PelaporIdentity {
	name, kelas := "", ""
	if s.DisplayName != nil {
		name = *s.DisplayName
	}
	if s.DisplayKelas != nil {
		kelas = *s.DisplayKelas
	}
	return PelaporIdentity{Name: name, Kelas: kelas}
}

func (s *Anon) GetProfile(session db.AnonSession) PelaporIdentity {
	return toIdentity(session)
}

type UpdateProfileInput struct {
	Name  string `json:"name"`
	Kelas string `json:"kelas"`
}

func (s *Anon) UpdateProfile(ctx context.Context, session db.AnonSession, in UpdateProfileInput) (PelaporIdentity, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Kelas = strings.TrimSpace(in.Kelas)
	if len([]rune(in.Name)) > 80 {
		return PelaporIdentity{}, utils.NewErrInvalidInput("Nama maksimal 80 karakter.")
	}
	if len([]rune(in.Kelas)) > 40 {
		return PelaporIdentity{}, utils.NewErrInvalidInput("Kelas maksimal 40 karakter.")
	}
	var namePtr, kelasPtr *string
	if in.Name != "" {
		namePtr = &in.Name
	}
	if in.Kelas != "" {
		kelasPtr = &in.Kelas
	}
	updated, err := s.queries.UpdateAnonProfile(ctx, db.UpdateAnonProfileParams{
		ID:           session.ID,
		DisplayName:  namePtr,
		DisplayKelas: kelasPtr,
	})
	if err != nil {
		return PelaporIdentity{}, utils.NewErrInternal("gagal menyimpan identitas")
	}
	return toIdentity(updated), nil
}
