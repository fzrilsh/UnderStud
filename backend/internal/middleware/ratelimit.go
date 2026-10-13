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
	return true
}

// CheckChat menolak bila kuota chat per sesi sudah habis.
func (rl *RateLimiter) CheckChat(sessionID string) error {
	if !rl.Allow(sessionID, "chat") {
		return utils.NewErrRateLimited("Terlalu banyak laporan. Coba lagi nanti.")
	}
	return nil
}

// CheckMessage menolak bila kuota pesan per sesi sudah habis.
func (rl *RateLimiter) CheckMessage(sessionID string) error {
	if !rl.Allow(sessionID, "message") {
		return utils.NewErrRateLimited("Terlalu banyak permintaan. Coba lagi nanti.")
	}
	return nil
}
