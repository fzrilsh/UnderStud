package cli

import (
	"flag"
	"fmt"

	"backend/internal/config"
	"backend/internal/geo"
)

func CheckLocation(cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("check-location", flag.ExitOnError)
	lat := fs.Float64("lat", 0, "latitude sekolah")
	lng := fs.Float64("lng", 0, "longitude sekolah")
	radius := fs.Int("radius", 200, "radius (meter)")
	_ = fs.Parse(args)

	if !geo.ValidCoord(*lat, *lng) {
		return fmt.Errorf("❌ Koordinat tidak valid (lat -90..90, lng -180..180).")
	}
	fmt.Printf("OK — koordinat valid: %.6f, %.6f (radius %d m)\n", *lat, *lng, *radius)
	if *radius < 100 {
		fmt.Println("⚠️  Radius terlalu kecil (<100 m). Akurasi GPS indoor bisa 50–100 m → rawan salah tolak.")
	}
	if *radius > 2000 {
		fmt.Println("⚠️  Radius sangat besar (>2 km). Geofence jadi tidak efektif.")
	}
	return nil
}
