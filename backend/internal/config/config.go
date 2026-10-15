package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration. Every field has a default so that
// zero env vars are required — `docker-compose up` must always be enough.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	CORSOrigins []string
	MaxDBConns  int32
	MinDBConns  int32
	LogLevel    string
	Timezone    string

	// Auth
	StaffTokenTTL time.Duration // staf: 1 minggu
	AnonTokenTTL  time.Duration // siswa anonim: 10 tahun
	CookieSecure  bool
	// TrustProxyHeaders: percayai X-Real-IP/X-Forwarded-For untuk rate-limit per-IP.
	// true hanya bila API DI BELAKANG reverse proxy tepercaya (nginx compose).
	TrustProxyHeaders bool

	// Geofence
	GeoFenceEnabled bool
	SchoolLat       float64
	SchoolLng       float64
	SchoolRadiusM   int

	// AI (OpenAI-compatible)
	AIBaseURL         string
	AIAPIKey          string
	AIModel           string
	AIDebounceSeconds int

	// Migrasi otomatis saat startup. Matikan (false) bila database diakses lewat
	// connection pooler mode transaction (pgbouncer) — migrasi dijalankan
	// terpisah via `cli migrate` langsung ke Postgres.
	MigrateOnStart bool

	// Upload
	// Catatan: batas per-file adalah storage.MaxFileBytes (10 MB, dipakai untuk
	// validasi magic-bytes). Tidak ada env terpisah agar tak ada config "mati".
	UploadDir        string
	MaxUploadTotalMB int

	// Rate limit
	RateLimitChatsPerHour    int
	RateLimitMessagesPerHour int
	RateLimitSessionsPerHour int // NV-2: batas pembuatan sesi anonim baru per-IP
	RateLimitIPMultiplier    int
}

// Load reads configuration from environment variables with sane defaults.
// Sebelum membaca, ia memuat file .env (bila ada) — sehingga cukup menaruh
// AI_BASE_URL/AI_API_KEY/AI_MODEL di .env tanpa export manual. Env asli
// (docker/VPS) menang atas .env karena godotenv tidak menimpa yang sudah ada.
func Load() (Config, error) {
	loadDotEnv()
	cfg := Config{
		Port: getEnv("PORT", "8080"),
		// 127.0.0.1 + port 5433: hindari Postgres native di host yang biasanya
		// memakai 5432 (dan bisa nabrak resolusi localhost→IPv6 ::1).
		// Port 5433 = port host default container db (lihat project/docker-compose.yml).
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5433/hackathon?sslmode=disable"),
		JWTSecret:   jwtSecret(),
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000"), ","),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Timezone:    getEnv("TZ", "Asia/Jakarta"),

		CookieSecure:      getEnvBool("COOKIE_SECURE", false),
		TrustProxyHeaders: getEnvBool("TRUST_PROXY_HEADERS", false),

		GeoFenceEnabled: getEnvBool("GEO_FENCE_ENABLED", false),
		SchoolRadiusM:   getEnvInt("SCHOOL_RADIUS_M", 200),

		AIBaseURL:         normalizeAIBaseURL(getEnv("AI_BASE_URL", "http://localhost:11434/v1")),
		AIAPIKey:          getEnv("AI_API_KEY", ""),
		AIModel:           getEnv("AI_MODEL", "gpt-4o-mini"),
		AIDebounceSeconds: getEnvInt("AI_DEBOUNCE_SECONDS", 5),

		MigrateOnStart: getEnvBool("MIGRATE_ON_START", true),

		UploadDir:        getEnv("UPLOAD_DIR", "./data/uploads"),
		MaxUploadTotalMB: getEnvInt("MAX_UPLOAD_TOTAL_MB", 20),

		RateLimitChatsPerHour:    getEnvInt("RATE_LIMIT_CHATS_PER_HOUR", 10),
		RateLimitMessagesPerHour: getEnvInt("RATE_LIMIT_MESSAGES_PER_HOUR", 30),
		RateLimitSessionsPerHour: getEnvInt("RATE_LIMIT_SESSIONS_PER_HOUR", 60),
		RateLimitIPMultiplier:    getEnvInt("RATE_LIMIT_IP_MULTIPLIER", 20),
	}

	var err error
	if cfg.StaffTokenTTL, err = parseDuration("STAFF_TOKEN_TTL", "168h"); err != nil {
		return Config{}, err
	}
	if cfg.AnonTokenTTL, err = parseDuration("ANON_TOKEN_TTL", "87600h"); err != nil {
		return Config{}, err
	}
	if cfg.MaxDBConns, err = parseInt32Env("DB_MAX_CONNS", 10); err != nil {
		return Config{}, err
	}
	if cfg.MinDBConns, err = parseInt32Env("DB_MIN_CONNS", 2); err != nil {
		return Config{}, err
	}
	if cfg.SchoolLat, err = getEnvFloat("SCHOOL_LAT", 0); err != nil {
		return Config{}, err
	}
	if cfg.SchoolLng, err = getEnvFloat("SCHOOL_LNG", 0); err != nil {
		return Config{}, err
	}

	for i := range cfg.CORSOrigins {
		cfg.CORSOrigins[i] = strings.TrimSpace(cfg.CORSOrigins[i])
	}
	return cfg, nil
}

// jwtSecret: pakai JWT_SECRET bila di-set; kalau tidak, generate acak 32 byte.
// Menghindari secret default yang diketahui publik (audit C1) tanpa melanggar
// prinsip "zero required env". Token jadi invalid saat restart (dev/demo OK).
func jwtSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		// Sangat jarang; fallback aman: nilai unik per-proses.
		return fmt.Sprintf("ephemeral-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// normalizeAIBaseURL membuat `.env` yang sama jalan di dua lingkungan tanpa
// perlu diedit: provider AI yang dijalankan di HOST dijangkau lewat `localhost`
// saat app jalan di host, tapi harus `host.docker.internal` saat app jalan di
// dalam container (di sana `localhost` = container itu sendiri).
//
// Aturan: bila app DI DALAM container dan host URL = localhost/127.0.0.1/::1,
// tulis ulang jadi host.docker.internal. Bila app di HOST dan host URL =
// host.docker.internal (tak bisa di-resolve di host), tulis ulang jadi 127.0.0.1.
func normalizeAIBaseURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return raw
	}
	host := u.Hostname()
	if host == "" {
		return raw
	}
	port := u.Port()
	inContainer := runningInContainer()

	loopback := host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "[::1]"
	dockerHost := host == "host.docker.internal"

	if inContainer && loopback {
		host = "host.docker.internal"
	} else if !inContainer && dockerHost {
		// Di host, host.docker.internal tidak resolve → pakai loopback IPv4
		// (hindari `localhost` yang bisa nyangkut ke IPv6 ::1 saat provider
		// hanya listen di IPv4).
		host = "127.0.0.1"
	} else {
		return raw
	}
	if port != "" {
		u.Host = net.JoinHostPort(host, port)
	} else {
		u.Host = host
	}
	return u.String()
}

// runningInContainer: heuristik murah — ada file /.dockerenv (Docker) atau
// env container (kubernetes). Tanpa dependensi tambahan.
func runningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if _, err := os.Stat("/run/.containerenv"); err == nil { // podman
		return true
	}
	return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}

// loadDotEnv memuat `.env` dari CWD atau direktori induk (best-effort). Tidak
// error bila file tak ada (default env tetap jalan). Nilai yang sudah ada di
// environment TIDAK ditimpa (godotenv.Load semantics) — jadi env produksi/docker
// selalu menang atas file .env lokal.
func loadDotEnv() {
	candidates := []string{".env", "backend/.env", "../.env", "../backend/.env"}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			return
		}
	}
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvFloat(key string, fallback float64) (float64, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return f, nil
}

func parseDuration(key, fallback string) (time.Duration, error) {
	d, err := time.ParseDuration(getEnv(key, fallback))
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return d, nil
}

func parseInt32Env(key string, fallback int32) (int32, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return int32(n), nil
}
