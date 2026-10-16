package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/ai"
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/storage"
	"backend/internal/tokens"
	"backend/internal/web"
)

// New builds the chi router with middleware chain and all routes mounted.
func New(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) http.Handler {
	// F21: pakai TZ terkonfigurasi (bukan hardcode) untuk semua tanggal/waktu.
	service.SetTimezone(cfg.Timezone)
	queries := db.New(pool)
	tm := tokens.New(cfg.JWTSecret, cfg.StaffTokenTTL)
	authSvc := service.NewAuth(queries, tm)
	authHandler := handler.NewAuthHandlers(authSvc, cfg.StaffTokenTTL, cfg.CookieSecure)

	rateLimiter := middleware.NewRateLimiter(middleware.RatelimitConfig{
		ChatsPerHour:    cfg.RateLimitChatsPerHour,
		MessagesPerHour: cfg.RateLimitMessagesPerHour,
		SessionsPerHour: cfg.RateLimitSessionsPerHour,
		IPMultiplier:    cfg.RateLimitIPMultiplier,
	})

	anonSvc := service.NewAnon(queries)
	anonMW := middleware.ResolveAnon(queries, cfg.AnonTokenTTL, cfg.CookieSecure, rateLimiter, cfg.TrustProxyHeaders)

	chatSvc := service.NewChat(queries, pool, storage.New(cfg.UploadDir), rateLimiter, cfg)
	chatHandler := handler.NewChatHandlers(chatSvc, cfg.TrustProxyHeaders)
	anonHandler := handler.NewAnonHandlers(anonSvc, chatSvc)

	reportSvc := service.NewReport(queries, pool, storage.New(cfg.UploadDir), cfg)
	reportHandler := handler.NewReportHandlers(reportSvc)

	attachmentHandler := handler.NewAttachmentHandlers(chatSvc)

	studentSvc := service.NewStudent(queries)
	observasiSvc := service.NewObservasi(queries, pool)

	aiClient := ai.NewClient(cfg)
	// Observability: tampilkan status AI saat start supaya operator tahu apakah
	// ringkasan aktif (AI_API_KEY terisi) atau nonaktif.
	if aiClient.Configured() {
		logger.Info("AI ringkasan aktif", "base_url", cfg.AIBaseURL, "model", cfg.AIModel)
	} else {
		logger.Info("AI ringkasan nonaktif (AI_API_KEY kosong)")
	}
	aiQueue := ai.NewQueue(queries, aiClient, time.Duration(cfg.AIDebounceSeconds)*time.Second, logger)
	go aiQueue.Start(context.Background())

	summarySvc := service.NewSummary(queries, aiClient.Configured())
	importerSvc := service.NewImporter(queries, pool, aiQueue)
	studentHandler := handler.NewStudentHandlers(studentSvc, observasiSvc, summarySvc, importerSvc).
		WithEnqueue(func(sid, by uuid.UUID, reason string) { aiQueue.Enqueue(sid, by, reason) })
	summaryHandler := handler.NewSummaryHandlers(summarySvc).
		WithEnqueue(func(sid, by uuid.UUID, reason string) { aiQueue.Enqueue(sid, by, reason) })
	dashboardSvc := service.NewDashboard(queries)
	dashboardHandler := handler.NewDashboardHandlers(dashboardSvc)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(chimw.Recoverer)
	// H4: batasi ukuran body request (cegah OOM/DoS tanpa auth). Lampiran chat
	// maks 20 MB total + overhead multipart → 25 MB cukup; JSON jauh di bawah.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.Body = http.MaxBytesReader(w, req.Body, 25<<20)
			next.ServeHTTP(w, req)
		})
	})
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
			// Logout sengaja TANPA auth: hanya membersihkan cookie milik pemanggil
			// (idempotent). Bila token sudah kedaluwarsa, user tetap harus bisa logout.
			r.Post("/logout", authHandler.Logout)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth(tm, queries))
				r.Get("/me", authHandler.Me)
			})
		})

		r.Route("/lapor", func(r chi.Router) {
			r.Use(middleware.RejectStaff(tm))
			r.Use(anonMW)
			r.Get("/", anonHandler.Bootstrap)
			r.Get("/profile", anonHandler.GetProfile)
			r.Put("/profile", anonHandler.UpdateProfile)

			r.Post("/chats", chatHandler.Create)
			r.Get("/chats", chatHandler.List)
			r.Get("/chats/{id}", chatHandler.Detail)
			r.Post("/chats/{id}/messages", chatHandler.Reply)
			r.Delete("/chats/{id}", chatHandler.Hide)

			r.Get("/attachments/{id}", attachmentHandler.Serve)
		})

		r.Route("/attachments", func(r chi.Router) {
			r.Use(middleware.RequireAuth(tm, queries))
			r.Get("/{id}", attachmentHandler.Serve)
		})

		r.Route("/reports", func(r chi.Router) {
			r.Use(middleware.RequireAuth(tm, queries))
			r.Use(middleware.RequireRole(middleware.RoleBK))
			r.Get("/", reportHandler.Queue)
			r.Get("/history", reportHandler.History)
			r.Get("/{id}", reportHandler.Detail)
			r.Post("/{id}/pin", reportHandler.Pin)
			r.Delete("/{id}/pin", reportHandler.Unpin)
			r.Post("/{id}/draft", reportHandler.Draft)
			r.Delete("/{id}/draft", reportHandler.Undraft)
			r.Post("/{id}/next", reportHandler.Next)
			r.Post("/{id}/skip", reportHandler.Skip)
			r.Post("/{id}/messages", reportHandler.Reply)
			r.Delete("/{id}", reportHandler.Hide)
		})

		r.Route("/students", func(r chi.Router) {
			r.Use(middleware.RequireAuth(tm, queries))
			// Read: BK (semua) + Wali Kelas (kelasnya) — scoping di service.
			r.Get("/", studentHandler.List)
			r.Get("/{id}", studentHandler.Detail)
			r.Get("/{id}/observations", studentHandler.ListObservations)
			r.Get("/{id}/summaries", summaryHandler.List)
			r.Get("/{id}/summaries/latest", summaryHandler.Latest)
			r.Get("/{id}/summaries/{sid}/export", summaryHandler.Export)
			// Write: Wali Kelas saja (L14 defense-in-depth; service tetap cek).
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(middleware.RoleWaliKelas))
				r.Post("/", studentHandler.Create)
				r.Get("/import/template", studentHandler.ImportTemplate)
				r.Post("/import/preview", studentHandler.ImportPreview)
				r.Post("/import/commit", studentHandler.ImportCommit)
				r.Patch("/{id}", studentHandler.Update)
				r.Put("/{id}/profile", studentHandler.SaveProfile)
				r.Post("/{id}/archive", studentHandler.Archive)
				r.Post("/{id}/restore", studentHandler.Restore)
				r.Post("/{id}/observations", studentHandler.CreateObservation)
				r.Post("/{id}/summaries", summaryHandler.Generate)
			})
		})

		r.Route("/observations", func(r chi.Router) {
			r.Use(middleware.RequireAuth(tm, queries))
			r.Get("/{id}/revisions", studentHandler.Revisions) // baca: staf scoped
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(middleware.RoleWaliKelas))
				r.Put("/{id}", studentHandler.UpdateObservation)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(middleware.RoleBK))
				r.Post("/{id}/handle", studentHandler.HandleRedFlag)
			})
		})

		r.Route("/red-flags", func(r chi.Router) {
			r.Use(middleware.RequireAuth(tm, queries))
			r.Get("/", studentHandler.RedFlags)
		})

		r.Route("/dashboard", func(r chi.Router) {
			r.Use(middleware.RequireAuth(tm, queries))
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(middleware.RoleBK))
				r.Get("/bk", dashboardHandler.BK)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(middleware.RoleWaliKelas))
				r.Get("/wk", dashboardHandler.WK)
			})
		})

		r.Route("/classes", func(r chi.Router) {
			r.Use(middleware.RequireAuth(tm, queries))
			r.Get("/", dashboardHandler.Classes)
		})

		r.Route("/checklist", func(r chi.Router) {
			r.Use(middleware.RequireAuth(tm, queries))
			r.Get("/", func(w http.ResponseWriter, req *http.Request) {
				handler.OK(w, service.ChecklistCatalog())
			})
		})
	})

	// FE (embedded) — didaftarkan PALING AKHIR: hanya menangkap path yang tidak
	// cocok route mana pun di atas (/api/*, /healthz tetap JSON).
	r.NotFound(web.Handler().ServeHTTP)

	return r
}
