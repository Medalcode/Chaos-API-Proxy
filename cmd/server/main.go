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
	"github.com/medalcode/chaos-api-proxy/internal/storage"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	log.Info("Starting Chaos API Proxy...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Redis storage
	store, err := storage.NewRedisStorage(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer store.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		log.Fatalf("Redis connection test failed: %v", err)
	}
	log.Info("Successfully connected to Redis")

	// Initialize handlers
	proxyHandler := handler.NewProxyHandler(store)
	configHandler := handler.NewConfigHandler(store)

	// Setup router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// Config management API
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/configs", configHandler.CreateConfig).Methods("POST")
	api.HandleFunc("/configs", configHandler.ListConfigs).Methods("GET")
	api.HandleFunc("/configs/{id}", configHandler.GetConfig).Methods("GET")
	api.HandleFunc("/configs/{id}", configHandler.UpdateConfig).Methods("PUT")
	api.HandleFunc("/configs/{id}", configHandler.DeleteConfig).Methods("DELETE")

	// Proxy endpoint - catches all other requests
	// Format: /proxy/{config-id}/{path:.*}
	router.PathPrefix("/proxy/{configID}").Handler(proxyHandler)

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
