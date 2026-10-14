package storage

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

var (
	ErrUnsupportedMedia = errors.New("unsupported media")
	ErrTooLarge         = errors.New("file too large")
)

// MaxFileBytes adalah batas ukuran per file (spec §8: 10 MB).
const MaxFileBytes = 10 * 1024 * 1024

// MaxImagePixels = batas total piksel gambar (audit H3): 25 MP. Mencegah
// decompression bomb (header kecil, dimensi raksasa → alokasi memori besar).
const MaxImagePixels = 25_000_000

type SavedFile struct {
	StoredPath   string // nama file relatif ke dir
	OriginalName string // nama file asli dari klien (untuk Content-Disposition)
	Mime         string // "image/jpeg" untuk gambar; asli untuk pdf/txt
	Size         int
}

type Store struct {
	dir string
}

func New(dir string) *Store {
	return &Store{dir: dir}
}

// SaveOne memvalidasi tipe via magic bytes, menangani gambar (re-encode JPEG,
// strip metadata) dan menyimpan PDF/TXT apa adanya. Nama file = UUID.
func (s *Store) SaveOne(data []byte, originalName string) (SavedFile, error) {
	if len(data) > MaxFileBytes {
		return SavedFile{}, ErrTooLarge
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return SavedFile{}, err
	}

	kind := detectKind(data)
	name := uuid.NewString()
	orig := sanitizeName(originalName)

	switch kind {
	case kindJPEG, kindPNG, kindGIF, kindWebP:
		// H3: batasi dimensi SEBELUM decode — cegah "decompression bomb"
		// (header kecil tapi dimensi raksasa → alokasi multi-GB → OOM).
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return SavedFile{}, ErrUnsupportedMedia
		}
		if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > MaxImagePixels {
			return SavedFile{}, ErrUnsupportedMedia
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return SavedFile{}, ErrUnsupportedMedia
		}
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return SavedFile{}, err
		}
		out := name + ".jpg"
		if err := os.WriteFile(filepath.Join(s.dir, out), buf.Bytes(), 0o644); err != nil {
			return SavedFile{}, err
		}
		return SavedFile{StoredPath: out, OriginalName: orig, Mime: "image/jpeg", Size: buf.Len()}, nil

	case kindPDF:
		out := name + ".pdf"
		if err := os.WriteFile(filepath.Join(s.dir, out), data, 0o644); err != nil {
			return SavedFile{}, err
		}
		return SavedFile{StoredPath: out, OriginalName: orig, Mime: "application/pdf", Size: len(data)}, nil

	case kindTXT:
		out := name + ".txt"
		if err := os.WriteFile(filepath.Join(s.dir, out), data, 0o644); err != nil {
			return SavedFile{}, err
		}
		return SavedFile{StoredPath: out, OriginalName: orig, Mime: "text/plain", Size: len(data)}, nil
	}
	return SavedFile{}, ErrUnsupportedMedia
}

// sanitizeName membuang komponen path dari nama klien (cegah path traversal /
// header injection) dan memberi fallback bila kosong.
func sanitizeName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "" || name == "." || name == "/" {
		return "lampiran"
	}
	return name
}

// Open membuka file tersimpan (untuk serve).
func (s *Store) Open(storedPath string) (*os.File, error) {
	return os.Open(filepath.Join(s.dir, storedPath))
}

// Delete menghapus file tersimpan (best-effort). Dipakai saat transaksi DB
// gagal setelah file ditulis (audit F8) — cegah file orphan.
func (s *Store) Delete(storedPath string) {
	if storedPath == "" {
		return
	}
	_ = os.Remove(filepath.Join(s.dir, storedPath))
}

type kind int

const (
	kindUnknown kind = iota
	kindJPEG
	kindPNG
	kindGIF
	kindWebP
	kindPDF
	kindTXT
)

func detectKind(b []byte) kind {
	switch {
	case len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return kindJPEG
	case len(b) >= 8 && bytes.Equal(b[:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return kindPNG
	case len(b) >= 6 && (bytes.Equal(b[:6], []byte("GIF87a")) || bytes.Equal(b[:6], []byte("GIF89a"))):
		return kindGIF
	case len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")):
		return kindWebP
	case len(b) >= 5 && bytes.Equal(b[:5], []byte("%PDF-")):
		return kindPDF
	case isProbablyText(b) && !looksLikeMarkup(b):
		return kindTXT
	}
	return kindUnknown
}

// isProbablyText: tidak ada byte NUL di 512 byte pertama & tidak kosong.
func isProbablyText(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	n := min(len(b), 512)
	for i := 0; i < n; i++ {
		if b[i] == 0 {
			return false
		}
	}
	return true
}

// looksLikeMarkup menolak SVG/HTML/XML (audit M8/NV-4) — spec memblokir eksplisit
// SVG/HTML/JS. Deteksi: lewati BOM UTF-8 + spasi/baris kosong (dalam urutan
// apa pun), lalu bila dokumen diawali '<' diikuti huruf/!/? → markup.
// Teks biasa jarang diawali '<'.
func looksLikeMarkup(b []byte) bool {
	i := 0
	for {
		// lewati BOM UTF-8 (bisa muncul sebelum spasi)
		if i+3 <= len(b) && bytes.Equal(b[i:i+3], []byte{0xEF, 0xBB, 0xBF}) {
			i += 3
			continue
		}
		if i < len(b) && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n' || b[i] == '\r') {
			i++
			continue
		}
		break
	}
	if i >= len(b) || b[i] != '<' || i+1 >= len(b) {
		return false
	}
	c := b[i+1]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '!' || c == '?'
}

// Describe dipakai untuk pesan error yang lebih jelas (opsional).
func Describe(f SavedFile) string {
	return fmt.Sprintf("%s (%d bytes)", f.Mime, f.Size)
}
