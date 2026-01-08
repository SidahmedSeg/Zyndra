package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/intelifox/click-deploy/internal/api"
	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/migrate"
	"github.com/intelifox/click-deploy/internal/store"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database with optimized connection pool
	poolConfig := store.PoolConfig{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: 600, // 10 minutes default
	}
	db, err := store.NewWithConfig(cfg.DatabaseURL, poolConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations automatically
	log.Println("=== STARTING DATABASE MIGRATIONS ===")
	if err := migrate.RunMigrations(db.DB, "migrations"); err != nil {
		log.Printf("❌ CRITICAL: Failed to run migrations: %v", err)
		log.Println("Server will start but API endpoints WILL FAIL")
		log.Println("Please check the error above and fix the migration issue")
		// Don't exit - let server start so we can see the error in logs
	} else {
		log.Println("✅ MIGRATIONS COMPLETED SUCCESSFULLY")
	}
	log.Println("=== MIGRATIONS FINISHED ===")

	// Set up router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(api.CORSMiddlewareFromEnv(cfg.CORSOrigins)) // CORS support
	r.Use(api.SecurityHeadersMiddleware)               // Security headers
	r.Use(api.CompressionMiddleware)                   // Enable response compression
	
	// Add panic recovery with detailed logging
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic recovered: %v\n", err)
					// Log stack trace
					debug.PrintStack()
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	})

	// Health check (no auth required, but rate limited)
	r.Group(func(r chi.Router) {
		r.Use(api.RateLimitMiddleware(10, time.Minute)) // 10 requests per minute for health checks
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			// Basic health check - can be extended to check database connectivity
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
	})

	// Prometheus metrics endpoint (no auth required, but rate limited)
	r.Group(func(r chi.Router) {
		r.Use(api.RateLimitMiddleware(60, time.Minute)) // 60 requests per minute for metrics
		r.Handle("/metrics", promhttp.Handler())
	})

	// OAuth callbacks (public, but will validate state)
	gitHandler := api.NewGitHandler(db, cfg)
	r.Get("/git/callback/github", gitHandler.CallbackGitHub)
	r.Get("/git/callback/gitlab", gitHandler.CallbackGitLab)

	// Authentication routes (public)
	if cfg.DisableAuth {
		// Use mock auth for development
		api.RegisterMockAuthRoutes(r, cfg)
	} else {
		// Use Casdoor OAuth
		api.RegisterAuthRoutes(r, cfg)
	}

	// Initialize auth validator
	// Use mock validator for development/testing
	var authValidator auth.ValidatorInterface
	if cfg.DisableAuth {
		authValidator = auth.NewMockValidator()
	} else {
		authValidator = auth.NewValidator(cfg.CasdoorEndpoint, cfg.CasdoorClientID)
	}

		// API routes (require authentication)
		r.Route("/v1/click-deploy", func(r chi.Router) {
			// Apply authentication middleware to all API routes
			r.Use(auth.Middleware(authValidator))
			// Apply rate limiting (100 requests per minute per user)
			r.Use(api.PerUserRateLimitMiddleware(100, time.Minute))

		// Projects endpoints
		projectHandler := api.NewProjectHandler(db, cfg)
		r.Get("/projects", projectHandler.ListProjects)
		r.Post("/projects", projectHandler.CreateProject)
		r.Get("/projects/{id}", projectHandler.GetProject)
		r.Patch("/projects/{id}", projectHandler.UpdateProject)
		r.Delete("/projects/{id}", projectHandler.DeleteProject)

		// Services endpoints
		serviceHandler := api.NewServiceHandler(db, cfg)
		r.Get("/projects/{id}/services", serviceHandler.ListServices)
		r.Post("/projects/{id}/services", serviceHandler.CreateService)
		r.Get("/services/{id}", serviceHandler.GetService)
		r.Patch("/services/{id}", serviceHandler.UpdateService)
		r.Patch("/services/{id}/position", serviceHandler.UpdateServicePosition)
		r.Delete("/services/{id}", serviceHandler.DeleteService)

		// Git endpoints
		api.RegisterGitRoutes(r, db, cfg)

		// Deployment endpoints
		api.RegisterDeploymentRoutes(r, db, cfg)

		// Database endpoints
		api.RegisterDatabaseRoutes(r, db, cfg)

		// Volume endpoints
		api.RegisterVolumeRoutes(r, db, cfg)

		// Environment variable endpoints
		api.RegisterEnvVarRoutes(r, db, cfg)

		// Realtime (Centrifugo) endpoints
		api.RegisterRealtimeRoutes(r, db, cfg)

		// Rollback endpoints
		api.RegisterRollbackRoutes(r, db, cfg)

		// Custom domain endpoints
		api.RegisterCustomDomainRoutes(r, db, cfg)

		// Metrics endpoints
		api.RegisterMetricsRoutes(r, db, cfg)
	})

	// Webhook endpoints (public, but validated via signature)
	api.RegisterWebhookRoutes(r, db, cfg)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	fmt.Printf("Server starting on :%s\n", cfg.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("Server exited")
}

