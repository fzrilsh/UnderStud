package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"backend/internal/utils"
)

type RatelimitConfig struct {
	ChatsPerHour    int
	MessagesPerHour int
	// SessionsPerHour: batas pembuatan sesi anonim BARU per-IP (NV-2). Mencegah
	// spammer membuang cookie → sesi tak terbatas. Default 60.
	SessionsPerHour int
	// IPMultiplier: batas per-IP = batas per-sesi × ini (IP dipakai bersama
	// banyak siswa via NAT sekolah → jauh lebih longgar). Default 20.
	IPMultiplier int
}

type rateEntry struct {
	timestamps []time.Time
}

// RateLimiter menyimpan hitungan sliding-window per session_id (in-memory).
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
	cfg     RatelimitConfig
	calls   uint64
}

func NewRateLimiter(cfg RatelimitConfig) *RateLimiter {
	return &RateLimiter{entries: map[string]*rateEntry{}, cfg: cfg}
}

// Allow mengembalikan true bila aksi masih di bawah batas.
// kind = "chat" | "message".
func (rl *RateLimiter) Allow(sessionID, kind string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-time.Hour)
	e, ok := rl.entries[sessionID]
	if !ok {
		e = &rateEntry{}
		rl.entries[sessionID] = e
	}
	kept := e.timestamps[:0]
	for _, ts := range e.timestamps {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	e.timestamps = kept

	limit := rl.cfg.ChatsPerHour
	if kind == "message" {
		limit = rl.cfg.MessagesPerHour
	}
	if len(e.timestamps) >= limit {
		return false
	}
	e.timestamps = append(e.timestamps, time.Now())

	// L3: sesekali buang entry kosong supaya map tak tumbuh tanpa batas.
	rl.evictLocked(cutoff)
	return true
}

// evictLocked membuang entry yang tak punya timestamp aktif. Dipanggil
// periodik (tiap N panggilan) agar biaya tetap rendah.
func (rl *RateLimiter) evictLocked(cutoff time.Time) {
	rl.calls++
	if rl.calls%256 != 0 {
		return
	}
	for k, e := range rl.entries {
		alive := false
		for _, ts := range e.timestamps {
			if ts.After(cutoff) {
				alive = true
				break
			}
		}
		if !alive {
			delete(rl.entries, k)
		}
	}
}

// AllowAny mengecek beberapa kunci sekaligus (mis. sesi + IP). Kunci PERTAMA
// memakai batas normal; kunci berikutnya memakai batas × ipMultiplier (IP
// dibagi banyak siswa via NAT sekolah → batas harus jauh lebih longgar; ini
// backstop anti-abuse, bukan kuota per-siswa). Lolos hanya bila SEMUA di bawah.
func (rl *RateLimiter) AllowAny(keys []string, kind string, ipMultiplier int) bool {
	if ipMultiplier < 1 {
		ipMultiplier = 1
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-time.Hour)
	base := rl.cfg.ChatsPerHour
	if kind == "message" {
		base = rl.cfg.MessagesPerHour
	}
	// Cek semua dulu (jangan tambah timestamp kalau ada yang penuh).
	counts := make([]*rateEntry, 0, len(keys))
	for i, k := range keys {
		limit := base
		if i > 0 {
			limit = base * ipMultiplier
		}
		e, ok := rl.entries[k]
		if !ok {
			e = &rateEntry{}
			rl.entries[k] = e
		}
		kept := e.timestamps[:0]
		for _, ts := range e.timestamps {
			if ts.After(cutoff) {
				kept = append(kept, ts)
			}
		}
		e.timestamps = kept
		if len(e.timestamps) >= limit {
			return false
		}
		counts = append(counts, e)
	}
	for _, e := range counts {
		e.timestamps = append(e.timestamps, now)
	}
	rl.evictLocked(cutoff)
	return true
}

// CheckChat menolak bila kuota chat habis. Membatasi per-sesi DAN per-IP
// (audit M4): membuang cookie tidak menambah kuota karena IP tetap dihitung.
func (rl *RateLimiter) CheckChat(sessionID, ip string) error {
	if !rl.AllowAny([]string{"s:" + sessionID, "ip:" + ip}, "chat", rl.ipMultiplier()) {
		return utils.NewErrRateLimited("Terlalu banyak laporan. Coba lagi nanti.")
	}
	return nil
}

// CheckMessage menolak bila kuota pesan habis (per-sesi + per-IP).
func (rl *RateLimiter) CheckMessage(sessionID, ip string) error {
	if !rl.AllowAny([]string{"s:" + sessionID, "ip:" + ip}, "message", rl.ipMultiplier()) {
		return utils.NewErrRateLimited("Terlalu banyak permintaan. Coba lagi nanti.")
	}
	return nil
}

// AllowAnonSession menolak bila pembuatan sesi anonim baru per-IP melewati batas
// (NV-2). Dipanggil middleware ResolveAnon SEBELUM membuat sesi baru.
func (rl *RateLimiter) AllowAnonSession(ip string) bool {
	limit := rl.cfg.SessionsPerHour
	if limit < 1 {
		limit = 60
	}
	key := "newsess:" + ip
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-time.Hour)
	e, ok := rl.entries[key]
	if !ok {
		e = &rateEntry{}
		rl.entries[key] = e
	}
	kept := e.timestamps[:0]
	for _, ts := range e.timestamps {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	e.timestamps = kept
	if len(e.timestamps) >= limit {
		return false
	}
	e.timestamps = append(e.timestamps, now)
	rl.evictLocked(cutoff)
	return true
}

// ClientIP mengambil IP klien dari request. Bila trustProxyHeaders aktif
// (API di belakang proxy tepercaya), pakai X-Real-IP/X-Forwarded-For.
func ClientIP(r *http.Request, trustProxyHeaders bool) string {
	if trustProxyHeaders {
		if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
			return ip
		}
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if first := strings.TrimSpace(strings.Split(xff, ",")[0]); first != "" {
				return first
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (rl *RateLimiter) ipMultiplier() int {
	if rl.cfg.IPMultiplier < 1 {
		return 20
	}
	return rl.cfg.IPMultiplier
}
