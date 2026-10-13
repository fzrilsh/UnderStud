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
	"backend/internal/storage"
	"backend/internal/tokens"
)

// New builds the chi router with middleware chain and all routes mounted.
func New(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) http.Handler {
	queries := db.New(pool)
	tm := tokens.New(cfg.JWTSecret, cfg.StaffTokenTTL)
	authSvc := service.NewAuth(queries, tm)
	authHandler := handler.NewAuthHandlers(authSvc, cfg.StaffTokenTTL)

	anonSvc := service.NewAnon(queries)
	anonHandler := handler.NewAnonHandlers(anonSvc)
	anonMW := middleware.ResolveAnon(queries, cfg.AnonTokenTTL, cfg.CookieSecure)

	rateLimiter := middleware.NewRateLimiter(middleware.RatelimitConfig{
		ChatsPerHour:    cfg.RateLimitChatsPerHour,
		MessagesPerHour: cfg.RateLimitMessagesPerHour,
	})
	chatSvc := service.NewChat(queries, storage.New(cfg.UploadDir), rateLimiter, cfg)
	chatHandler := handler.NewChatHandlers(chatSvc)

	reportSvc := service.NewReport(queries)
	reportHandler := handler.NewReportHandlers(reportSvc)

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

		r.Route("/lapor", func(r chi.Router) {
			r.Use(anonMW)
			r.Get("/", anonHandler.Bootstrap)
			r.Get("/profile", anonHandler.GetProfile)
			r.Put("/profile", anonHandler.UpdateProfile)

			r.Post("/chats", chatHandler.Create)
			r.Get("/chats", chatHandler.List)
			r.Get("/chats/{id}", chatHandler.Detail)
			r.Post("/chats/{id}/messages", chatHandler.Reply)
			r.Delete("/chats/{id}", chatHandler.Hide)
		})

		r.Route("/reports", func(r chi.Router) {
			r.Use(middleware.RequireAuth(tm))
			r.Use(middleware.RequireRole(middleware.RoleBK))
			r.Get("/", reportHandler.Queue)
			r.Get("/{id}", reportHandler.Detail)
			r.Post("/{id}/pin", reportHandler.Pin)
			r.Delete("/{id}/pin", reportHandler.Unpin)
			r.Post("/{id}/draft", reportHandler.Draft)
			r.Post("/{id}/next", reportHandler.Next)
			r.Post("/{id}/messages", reportHandler.Reply)
			r.Delete("/{id}", reportHandler.Hide)
		})
	})

	return r
}
