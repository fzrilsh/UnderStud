# ---- build stage ----
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/api

# ---- run stage ----
FROM alpine:3.20
RUN adduser -D -u 10001 appuser && apk add --no-cache ca-certificates wget
USER appuser
WORKDIR /app
COPY --from=build /bin/api /app/api
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=5 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1
CMD ["/app/api"]
