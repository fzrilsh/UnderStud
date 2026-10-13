package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/tokens"
)

// New builds the chi router with middleware chain and all routes mounted.
func New(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) http.Handler {
	queries := db.New(pool)
	tm := tokens.New(cfg.JWTSecret, cfg.StaffTokenTTL)
	authSvc := service.NewAuth(queries, tm)
	authHandler := handler.NewAuthHandlers(authSvc, cfg.StaffTokenTTL)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", handler.Health(pool))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth(tm))
				r.Get("/me", authHandler.Me)
			})
		})
	})

	return r
}
