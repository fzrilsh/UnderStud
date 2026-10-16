package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"backend/internal/utils"
)

// OK writes 200 with an envelope {"data": ...} (spec §21.1).
func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

// Created writes 201 with an envelope {"data": ...}.
func Created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, map[string]any{"data": data})
}

// OKMeta writes 200 with {"data": ..., "meta": ...} — dipakai endpoint list.
func OKMeta(w http.ResponseWriter, data any, meta any) {
	writeJSON(w, http.StatusOK, map[string]any{"data": data, "meta": meta})
}

// Accepted writes 202 with {"data": ...} — untuk aksi asinkron (mis. queue AI).
func Accepted(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusAccepted, map[string]any{"data": data})
}

// Error maps any error to the standard shape {"error":{code,message}}.
// AppErr keeps its status/code; unknown errors become 500 (message generic).
func Error(w http.ResponseWriter, err error) {
	var appErr *utils.AppErr
	if errors.As(err, &appErr) {
		writeJSON(w, appErr.Status, map[string]any{"error": map[string]string{
			"code":    appErr.Code,
			"message": appErr.Message,
		}})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]string{
		"code":    utils.CodeInternal,
		"message": "Terjadi kesalahan. Coba lagi.",
	}})
}

// Bind decodes a JSON request body into dst. Empty body → INVALID_INPUT.
func Bind(r *http.Request, dst any) error {
	if r.Body == nil || r.Body == http.NoBody {
		return utils.NewErrInvalidInput("Body permintaan wajib diisi.")
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return utils.NewErrInvalidInput("Format data tidak valid.")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// attachmentHeader menulis Content-Disposition aman (audit L6): buang karakter
// yang bisa memecah header (quote, CR/LF, backslash) dari nama file klien.
func attachmentHeader(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Disposition", `attachment; filename="`+sanitizeHeaderValue(filename)+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func sanitizeHeaderValue(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '"' || r == '\\' || r == '\r' || r == '\n' || r < 0x20:
			// buang; ganti pemisah dengan '_'
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "unduhan"
	}
	return out
}
