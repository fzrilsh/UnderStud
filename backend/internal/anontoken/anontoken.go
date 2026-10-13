package anontoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// New menghasilkan token acak 32 byte (hex) dan hash SHA-256-nya.
// Token asli hanya dikirim ke klien via cookie; server menyimpan hash-nya saja.
func New() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	return raw, Hash(raw), nil
}

// Hash menghitung SHA-256 hex dari token mentah.
func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
