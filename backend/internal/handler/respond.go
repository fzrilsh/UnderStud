package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
		"message": "internal server error",
	}})
}

// Bind decodes a JSON request body into dst. Empty body → INVALID_INPUT.
func Bind(r *http.Request, dst any) error {
	if r.Body == nil || r.Body == http.NoBody {
		return utils.NewErrInvalidInput("request body is required")
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return utils.NewErrInvalidInput(fmt.Sprintf("invalid JSON body: %v", err))
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
