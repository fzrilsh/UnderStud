package config

import "fmt"

// DefaultJWTSecret = nilai default dev. JANGAN dipakai di produksi.
const DefaultJWTSecret = "dev-secret-change-me-in-prod"

// Validate enforces invariants that must hold before the server starts.
// Fail-fast: lebih baik menolak start daripada diam-diam salah perilaku.
func Validate(cfg Config) error {
	// C1: secret default = siapa pun bisa memalsukan token (RBAC dari klaim).
	// Tolak bila secret = default atau terlalu pendek (<32 byte).
	if cfg.JWTSecret == DefaultJWTSecret {
		return fmt.Errorf("JWT_SECRET masih nilai default (%q). Set JWT_SECRET yang kuat sebelum menjalankan.", DefaultJWTSecret)
	}
	if len(cfg.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET terlalu pendek (%d byte); minimal 32 byte.", len(cfg.JWTSecret))
	}

	if cfg.GeoFenceEnabled {
		if cfg.SchoolLat == 0 && cfg.SchoolLng == 0 {
			return fmt.Errorf("GEO_FENCE_ENABLED=true tetapi SCHOOL_LAT/SCHOOL_LNG belum di-set. " +
				"Set koordinat sekolah dulu, atau matikan geofence (GEO_FENCE_ENABLED=false).")
		}
		if cfg.SchoolLat < -90 || cfg.SchoolLat > 90 {
			return fmt.Errorf("SCHOOL_LAT tidak valid (%v); harus -90..90", cfg.SchoolLat)
		}
		if cfg.SchoolLng < -180 || cfg.SchoolLng > 180 {
			return fmt.Errorf("SCHOOL_LNG tidak valid (%v); harus -180..180", cfg.SchoolLng)
		}
		if cfg.SchoolRadiusM <= 0 {
			return fmt.Errorf("SCHOOL_RADIUS_M harus > 0 (dapat %d)", cfg.SchoolRadiusM)
		}
	}
	return nil
}
