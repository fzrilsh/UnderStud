package main

import (
	"fmt"
	"os"

	"backend/internal/cli"
	"backend/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "migrate":
		err = cli.Migrate(cfg)
	case "create-staff":
		err = cli.CreateStaff(cfg, args)
	case "deactivate-staff":
		err = cli.DeactivateStaff(cfg, args)
	case "check-location":
		err = cli.CheckLocation(cfg, args)
	case "seed-demo":
		err = cli.SeedDemo(cfg)
	default:
		usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `UnderStud CLI

  migrate                                                      jalankan migrasi database lalu keluar
  create-staff --email E --password P --name N --role BK|WALI_KELAS [--kelas XI-IPA-1]
  deactivate-staff --email E
  check-location --lat LAT --lng LNG [--radius 200]
  seed-demo`)
}
