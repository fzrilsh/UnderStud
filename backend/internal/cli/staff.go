package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/utils"
)

func connect(cfg config.Config) (*pgxpool.Pool, *db.Queries, error) {
	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL, cfg.MaxDBConns, cfg.MinDBConns)
	if err != nil {
		return nil, nil, err
	}
	return pool, db.New(pool), nil
}

func CreateStaff(cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("create-staff", flag.ExitOnError)
	email := fs.String("email", "", "email staf")
	password := fs.String("password", "", "kata sandi (min 8)")
	name := fs.String("name", "", "nama")
	role := fs.String("role", "", "BK | WALI_KELAS")
	kelas := fs.String("kelas", "", "kelas (wajib untuk WALI_KELAS)")
	_ = fs.Parse(args)

	if *email == "" || !strings.Contains(*email, "@") {
		return fmt.Errorf("--email wajib & harus format email")
	}
	if len(*password) < 8 {
		return fmt.Errorf("--password minimal 8 karakter")
	}
	if *name == "" || len(*name) > 100 {
		return fmt.Errorf("--name wajib (1-100 karakter)")
	}
	if *role != "BK" && *role != "WALI_KELAS" {
		return fmt.Errorf("--role harus BK atau WALI_KELAS")
	}
	if *role == "WALI_KELAS" && *kelas == "" {
		return fmt.Errorf("--kelas wajib untuk WALI_KELAS")
	}
	if *role == "BK" && *kelas != "" {
		return fmt.Errorf("--kelas hanya untuk WALI_KELAS")
	}

	pool, q, err := connect(cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	hash, err := utils.HashPassword(*password)
	if err != nil {
		return err
	}
	var kelasPtr *string
	if *kelas != "" {
		kelasPtr = kelas
	}
	u, err := q.CreateStaff(context.Background(), db.CreateStaffParams{
		Email:        strings.ToLower(*email),
		PasswordHash: hash,
		Name:         *name,
		Role:         *role,
		Kelas:        kelasPtr,
	})
	if err != nil {
		return err
	}
	fmt.Printf("OK — staf dibuat: %s (%s)\n", u.Email, u.Role)
	return nil
}

func DeactivateStaff(cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("deactivate-staff", flag.ExitOnError)
	email := fs.String("email", "", "email staf")
	_ = fs.Parse(args)
	if *email == "" {
		return fmt.Errorf("--email wajib")
	}
	pool, q, err := connect(cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := q.DeactivateStaff(context.Background(), strings.ToLower(*email)); err != nil {
		return err
	}
	fmt.Printf("OK — staf dinonaktifkan: %s\n", *email)
	return nil
}
