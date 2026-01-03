package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/medalcode/chaos-api-proxy/internal/config"
	"github.com/medalcode/chaos-api-proxy/internal/handler"
	"github.com/medalcode/chaos-api-proxy/internal/middleware"
	"github.com/medalcode/chaos-api-proxy/internal/storage"
	"github.com/medalcode/chaos-api-proxy/internal/ui"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	// ... (logger setup) ...
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	log.Info("Starting Chaos API Proxy...")

    // ... (config load) ...
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

    // ... (redis setup) ...
	store, err := storage.NewRedisStorage(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer store.Close()

    // ... (redis test) ...
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		log.Fatalf("Redis connection test failed: %v", err)
	}
	log.Info("Successfully connected to Redis")

	// Initialize handlers
	proxyHandler := handler.NewProxyHandler(store)
	configHandler := handler.NewConfigHandler(store)
	authMiddleware := middleware.NewAuthMiddleware(cfg.APIKeys)
	
	if cfg.APIKeys != "" {
		log.Info("🔐 Authentication enabled for Admin API")
	} else {
		log.Warn("⚠️  Authentication DISABLED. Set CHAOS_API_KEYS environment variable to secure the Admin API.")
	}

	// Setup router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")
	
	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

    // UI Dashboard
    router.PathPrefix("/dashboard").Handler(http.StripPrefix("/dashboard", ui.Handler()))
    
    // Redirect root to dashboard (if it's not a proxy request via header)
    // NOTE: This conflicts with the catch-all proxy if not careful.
    // The main proxy logic uses PathPrefix("/") at the end.
    // We will let the proxy handler decide if it shows dashboard or proxies based on headers/path.
    // For now, access via /dashboard directly.

	// Config management API (Protected)
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(authMiddleware.Handler)
	api.HandleFunc("/configs", configHandler.CreateConfig).Methods("POST")
	api.HandleFunc("/configs", configHandler.ListConfigs).Methods("GET")
	api.HandleFunc("/configs/{id}", configHandler.GetConfig).Methods("GET")
	api.HandleFunc("/configs/{id}", configHandler.UpdateConfig).Methods("PUT")
	api.HandleFunc("/configs/{id}", configHandler.DeleteConfig).Methods("DELETE")
	api.HandleFunc("/logs", configHandler.GetRequestLogs).Methods("GET")

	// Alias endpoints (Protected)
	// Note: We need a subrouter or wrap individually to apply middleware only here
	// and not to global router if we added it globally.
	// Let's create a subrouter for /rules logic manually or just wrap them.
	// Since /rules is root level, we can't easily use PathPrefix for just /rules without capturing others.
	// Easiest is to wrap the handler functions or use a matcher.
	
	// Better approach:
	rulesRouter := router.PathPrefix("/rules").Subrouter()
	rulesRouter.Use(authMiddleware.Handler)
	// Note: PathPrefix("/rules") captures /rules AND /rules/foo
	// We need to match exact paths inside the subrouter now.
	rulesRouter.HandleFunc("", configHandler.CreateConfig).Methods("POST")
	rulesRouter.HandleFunc("", configHandler.ListConfigs).Methods("GET")
	rulesRouter.HandleFunc("/{id}", configHandler.GetConfig).Methods("GET")
	rulesRouter.HandleFunc("/{id}", configHandler.UpdateConfig).Methods("PUT")
	rulesRouter.HandleFunc("/{id}", configHandler.DeleteConfig).Methods("DELETE")
	
	// Logs alias
	router.Handle("/logs", authMiddleware.Handler(http.HandlerFunc(configHandler.GetRequestLogs))).Methods("GET")

	// Proxy endpoint - path-based routing
	// Format: /proxy/{configID}/{path:.*}
	router.PathPrefix("/proxy/{configID}").Handler(proxyHandler)

	// Catchall for header-based routing (X-Chaos-Config-ID)
	// This must be last to avoid conflicts
	router.PathPrefix("/").Handler(proxyHandler)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Server listening on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped gracefully")
}
