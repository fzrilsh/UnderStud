# ---- build stage ----
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
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
