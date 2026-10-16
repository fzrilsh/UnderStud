# ============================================================================
# Build context = folder `project/` (BUKAN `backend/`), karena butuh `frontend/`
# juga untuk membangun & meng-embed FE ke binary Go. Di compose:
#   build: { context: ., dockerfile: backend/Dockerfile }
# ============================================================================

# ---- stage 1: build frontend ----
FROM node:22-alpine AS webbuild
RUN corepack enable
WORKDIR /web
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

# ---- stage 2: build backend (dengan FE ter-embed) ----
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# Salin hasil build FE ke lokasi embed SEBELUM go build.
COPY --from=webbuild /web/dist ./internal/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/cli ./cmd/cli

# ---- run stage ----
FROM alpine:3.20
# tzdata wajib: kode memakai time.LoadLocation("Asia/Jakarta") — tanpa tzdata
# LoadLocation mengembalikan nil dan bisa panic saat menulis tanggal.
# appuser non-root; siapkan direktori upload agar bisa ditulis (spec §8/§15).
RUN adduser -D -u 10001 appuser \
 && apk add --no-cache ca-certificates wget tzdata \
 && mkdir -p /app/data/uploads && chown -R appuser:appuser /app
USER appuser
WORKDIR /app
COPY --from=build /bin/api /app/api
COPY --from=build /bin/cli /app/cli
# Volume lampiran (spec §15: data persist). Tanpa ini, lampiran hilang saat recreate.
VOLUME ["/app/data"]
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=5 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1
CMD ["/app/api"]
