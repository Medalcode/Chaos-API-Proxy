package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/medalcode/chaos-api-proxy/internal/models"
	"github.com/medalcode/chaos-api-proxy/internal/storage"
	log "github.com/sirupsen/logrus"
)

// ConfigHandler handles configuration management endpoints
type ConfigHandler struct {
	storage *storage.RedisStorage
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(storage *storage.RedisStorage) *ConfigHandler {
	return &ConfigHandler{
		storage: storage,
	}
}

// CreateConfig handles POST /api/v1/configs
func (h *ConfigHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	var config models.ChaosConfig

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		log.WithError(err).Error("Failed to decode request body")
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Generate ID if not provided
	if config.ID == "" {
		config.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	// Set default enabled state
	if !config.Enabled {
		config.Enabled = true
	}

	// Validate config
	if err := config.Validate(); err != nil {
		log.WithError(err).Error("Config validation failed")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Save to storage
	ctx := r.Context()
	if err := h.storage.SaveConfig(ctx, &config); err != nil {
		log.WithError(err).Error("Failed to save config")
		respondError(w, http.StatusInternalServerError, "Failed to save configuration")
		return
	}

	log.WithField("config_id", config.ID).Info("Created new config")
	respondJSON(w, http.StatusCreated, config)
}

// GetConfig handles GET /api/v1/configs/{id}
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ctx := r.Context()
	config, err := h.storage.GetConfig(ctx, id)
	if err != nil {
		log.WithError(err).WithField("config_id", id).Error("Failed to get config")
		respondError(w, http.StatusNotFound, "Configuration not found")
		return
	}

	respondJSON(w, http.StatusOK, config)
}

// ListConfigs handles GET /api/v1/configs
func (h *ConfigHandler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	configs, err := h.storage.ListConfigs(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to list configs")
		respondError(w, http.StatusInternalServerError, "Failed to list configurations")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"configs": configs,
		"count":   len(configs),
	})
}

// UpdateConfig handles PUT /api/v1/configs/{id}
func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var config models.ChaosConfig

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		log.WithError(err).Error("Failed to decode request body")
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ensure ID matches
	config.ID = id

	// Validate config
	if err := config.Validate(); err != nil {
		log.WithError(err).Error("Config validation failed")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update in storage
	ctx := r.Context()
	if err := h.storage.UpdateConfig(ctx, &config); err != nil {
		log.WithError(err).WithField("config_id", id).Error("Failed to update config")
		respondError(w, http.StatusInternalServerError, "Failed to update configuration")
		return
	}

	log.WithField("config_id", id).Info("Updated config")
	respondJSON(w, http.StatusOK, config)
}

// DeleteConfig deletes a chaos configuration
func (h *ConfigHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ctx := r.Context()
	if err := h.storage.DeleteConfig(ctx, id); err != nil {
		log.WithError(err).WithField("config_id", id).Error("Failed to delete config")
		respondError(w, http.StatusInternalServerError, "Failed to delete configuration")
		return
	}

	log.WithField("config_id", id).Info("Deleted config")
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Configuration deleted successfully",
		"id":      id,
	})
}

// GetRequestLogs retrieves the latest request logs
func (h *ConfigHandler) GetRequestLogs(w http.ResponseWriter, r *http.Request) {
	// Parse limit query param
	limitStr := r.URL.Query().Get("limit")
	limit := int64(50) // Default limit
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := h.storage.GetLogs(r.Context(), limit)
	if err != nil {
		log.WithError(err).Error("Failed to get logs")
		respondError(w, http.StatusInternalServerError, "Failed to get logs")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"logs": logs})
}

// Helper functions

func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.WithError(err).Error("Failed to encode response")
	}
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
