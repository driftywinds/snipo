package api

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MohamedElashri/snipo/internal/api/handlers"
	"github.com/MohamedElashri/snipo/internal/api/middleware"
	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/services"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	DB              *sql.DB
	Logger          *slog.Logger
	AuthService     *auth.Service
	Version         string
	Commit          string
	RateLimit       int
	RateLimitWindow int // in seconds
}

// NewRouter creates and configures the HTTP router
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery(cfg.Logger))
	r.Use(middleware.Logger(cfg.Logger))
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS)

	// Rate limiting for auth endpoints
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, 60*1000*1000*1000) // 1 minute in nanoseconds (nothing planck scale)

	// Create repositories
	snippetRepo := repository.NewSnippetRepository(cfg.DB)

	// Create services
	snippetService := services.NewSnippetService(snippetRepo, cfg.Logger)

	// Create handlers
	snippetHandler := handlers.NewSnippetHandler(snippetService)
	authHandler := handlers.NewAuthHandler(cfg.AuthService)
	healthHandler := handlers.NewHealthHandler(cfg.DB, cfg.Version, cfg.Commit)

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Health checks
		r.Get("/health", healthHandler.Health)
		r.Get("/ping", healthHandler.Ping)

		// Public snippet access
		r.Get("/api/v1/snippets/public/{id}", snippetHandler.GetPublic)

		// Auth endpoints (with rate limiting)
		r.Group(func(r chi.Router) {
			r.Use(rateLimiter.Middleware)
			r.Post("/api/v1/auth/login", authHandler.Login)
		})

		r.Post("/api/v1/auth/logout", authHandler.Logout)
		r.Get("/api/v1/auth/check", authHandler.Check)
	})

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(cfg.AuthService))

		// Snippet CRUD
		r.Route("/api/v1/snippets", func(r chi.Router) {
			r.Get("/", snippetHandler.List)
			r.Post("/", snippetHandler.Create)
			r.Get("/search", snippetHandler.Search)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", snippetHandler.Get)
				r.Put("/", snippetHandler.Update)
				r.Delete("/", snippetHandler.Delete)
				r.Post("/favorite", snippetHandler.ToggleFavorite)
				r.Post("/duplicate", snippetHandler.Duplicate)
			})
		})
	})

	return r
}
