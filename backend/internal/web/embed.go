package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// dist = hasil build Vite (frontend/dist) yang disalin ke sini oleh `make build-web`
// (atau Dockerfile stage). Placeholder `dist/index.html` disertakan agar `go build`
// tetap sukses sebelum FE di-build.
//
//go:embed all:dist
var distFS embed.FS

// Handler menyajikan aset FE hasil build + SPA fallback (index.html).
// Path yang diawali /api/ atau /healthz TIDAK ditangani di sini — chi sudah
// mendaftarkannya lebih dulu, jadi handler ini hanya menerima sisa path.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return http.NotFoundHandler()
	}
	fileServer := http.FileServer(http.FS(sub))
	index, _ := fs.ReadFile(sub, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(sub, p); err != nil {
			// SPA fallback: rute klien (mis. /students/xyz) → index.html.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(index)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
