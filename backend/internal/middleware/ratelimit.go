package middleware

import (
	"sync"
	"time"

	"backend/internal/utils"
)

type RatelimitConfig struct {
	ChatsPerHour    int
	MessagesPerHour int
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

// AllowAny mengecek beberapa kunci sekaligus (mis. sesi + IP): lolos hanya bila
// SEMUA kunci di bawah batas (audit M4 — cegah bypass dengan membuang cookie).
func (rl *RateLimiter) AllowAny(keys []string, kind string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-time.Hour)
	limit := rl.cfg.ChatsPerHour
	if kind == "message" {
		limit = rl.cfg.MessagesPerHour
	}
	// Cek semua dulu (jangan tambah timestamp kalau ada yang penuh).
	counts := make([]*rateEntry, 0, len(keys))
	for _, k := range keys {
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
	if !rl.AllowAny([]string{"s:" + sessionID, "ip:" + ip}, "chat") {
		return utils.NewErrRateLimited("Terlalu banyak laporan. Coba lagi nanti.")
	}
	return nil
}

// CheckMessage menolak bila kuota pesan habis (per-sesi + per-IP).
func (rl *RateLimiter) CheckMessage(sessionID, ip string) error {
	if !rl.AllowAny([]string{"s:" + sessionID, "ip:" + ip}, "message") {
		return utils.NewErrRateLimited("Terlalu banyak permintaan. Coba lagi nanti.")
	}
	return nil
}
