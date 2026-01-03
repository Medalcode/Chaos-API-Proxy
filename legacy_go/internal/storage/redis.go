package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/medalcode/chaos-api-proxy/internal/models"
)

const (
	configPrefix = "chaos:config:"
	configList   = "chaos:configs"
	logsKey      = "chaos:logs:global"
	maxLogs      = 100 // Keep last 100 requests
)

// RequestLog represents a single logged request
type RequestLog struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	ConfigID    string    `json:"config_id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	StatusCode  int       `json:"status_code"`
	DurationMs  int64     `json:"duration_ms"`
	ChaosType   string    `json:"chaos_type"`
}

// RedisStorage implements storage using Redis
type RedisStorage struct {
	client *redis.Client
}
// ... (rest of NewRedisStorage, Ping, Close, CRUD methods remain the same until UpdateConfig) ...

// UpdateConfig updates an existing configuration
func (s *RedisStorage) UpdateConfig(ctx context.Context, config *models.ChaosConfig) error {
	// ... (implementation same as before) ...
	// Check if config exists
	exists, err := s.client.Exists(ctx, configPrefix+config.ID).Result()
	if err != nil {
		return fmt.Errorf("failed to check config existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("config not found: %s", config.ID)
	}

	// Update timestamp
	config.UpdatedAt = time.Now()

	// Save updated config
	return s.SaveConfig(ctx, config)
}

// LogRequest saves a request log entry
func (s *RedisStorage) LogRequest(ctx context.Context, logEntry *RequestLog) error {
	data, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	pipeline := s.client.Pipeline()
	pipeline.LPush(ctx, logsKey, data)
	pipeline.LTrim(ctx, logsKey, 0, maxLogs-1) // Keep list capped
	
	if _, err := pipeline.Exec(ctx); err != nil {
		return fmt.Errorf("failed to save log: %w", err)
	}
	return nil
}

// GetLogs retrieves the latest request logs
func (s *RedisStorage) GetLogs(ctx context.Context, limit int64) ([]*RequestLog, error) {
	if limit <= 0 {
		limit = maxLogs
	}
	
	val, err := s.client.LRange(ctx, logsKey, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	logs := make([]*RequestLog, 0, len(val))
	for _, v := range val {
		var l RequestLog
		if err := json.Unmarshal([]byte(v), &l); err != nil {
			continue // skip corrupted logs
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisStorage{
		client: client,
	}, nil
}

// Ping tests the Redis connection
func (s *RedisStorage) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (s *RedisStorage) Close() error {
	return s.client.Close()
}

// SaveConfig saves a chaos configuration
func (s *RedisStorage) SaveConfig(ctx context.Context, config *models.ChaosConfig) error {
	// Serialize config to JSON
	data, err := config.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Save config
	key := configPrefix + config.ID
	if err := s.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Add to config list
	if err := s.client.SAdd(ctx, configList, config.ID).Err(); err != nil {
		return fmt.Errorf("failed to add to config list: %w", err)
	}

	return nil
}

// GetConfig retrieves a chaos configuration by ID
func (s *RedisStorage) GetConfig(ctx context.Context, id string) (*models.ChaosConfig, error) {
	key := configPrefix + id
	data, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("config not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	config, err := models.FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize config: %w", err)
	}

	return config, nil
}

// ListConfigs retrieves all configurations
func (s *RedisStorage) ListConfigs(ctx context.Context) ([]*models.ChaosConfig, error) {
	// Get all config IDs
	ids, err := s.client.SMembers(ctx, configList).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list config IDs: %w", err)
	}

	configs := make([]*models.ChaosConfig, 0, len(ids))
	for _, id := range ids {
		config, err := s.GetConfig(ctx, id)
		if err != nil {
			// Skip invalid configs but log the error
			continue
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// DeleteConfig removes a chaos configuration
func (s *RedisStorage) DeleteConfig(ctx context.Context, id string) error {
	key := configPrefix + id

	// Delete config
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	// Remove from config list
	if err := s.client.SRem(ctx, configList, id).Err(); err != nil {
		return fmt.Errorf("failed to remove from config list: %w", err)
	}

	return nil
}

// UpdateConfig updates an existing configuration
func (s *RedisStorage) UpdateConfig(ctx context.Context, config *models.ChaosConfig) error {
	// Check if config exists
	exists, err := s.client.Exists(ctx, configPrefix+config.ID).Result()
	if err != nil {
		return fmt.Errorf("failed to check config existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("config not found: %s", config.ID)
	}

	// Update timestamp
	config.UpdatedAt = time.Now()

	// Save updated config
	return s.SaveConfig(ctx, config)
}
