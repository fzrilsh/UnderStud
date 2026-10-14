package utils

import (
	"encoding/json"
	"errors"
	"net/http"
)

// WriteAppError writes an AppErr-shaped error response directly.
// Used by middleware, which lives below the handler layer and cannot import handler.
func WriteAppError(w http.ResponseWriter, err error) {
	appErr, ok := err.(*AppErr)
	if !ok {
		appErr = &AppErr{Status: http.StatusInternalServerError, Code: CodeInternal, Message: "Terjadi kesalahan. Coba lagi."}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{
		"code":    appErr.Code,
		"message": appErr.Message,
	}})
}

// IsAppErr reports whether err is a typed AppErr.
func IsAppErr(err error) bool {
	var appErr *AppErr
	return errors.As(err, &appErr)
}
