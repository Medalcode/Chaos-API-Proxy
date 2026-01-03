package models

import (
	"encoding/json"
	"time"
)

// ChaosConfig defines the chaos engineering rules for a proxy configuration
type ChaosConfig struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Target      string    `json:"target"` // Target API URL (e.g., "https://api.stripe.com")
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Chaos Engineering Parameters
	Rules ChaosRules `json:"rules"`
}

// ChaosRules contains all the chaos injection parameters
type ChaosRules struct {
	// Latency Injection
	LatencyMs int `json:"latency_ms,omitempty"` // Fixed latency in milliseconds
	Jitter    int `json:"jitter,omitempty"`     // Random variation in latency (±jitter ms)

	// Error Injection
	InjectFailureRate float64 `json:"inject_failure_rate,omitempty"` // Probability of failure (0.0 to 1.0)
	ErrorCode         int     `json:"error_code,omitempty"`          // HTTP error code to return (e.g., 500, 503)
	ErrorBody         string  `json:"error_body,omitempty"`          // Custom error response body

	// Connection Chaos
	DropConnection     bool    `json:"drop_connection,omitempty"`      // Close socket without responding
	DropConnectionRate float64 `json:"drop_connection_rate,omitempty"` // Probability of dropping connection

	// Bandwidth Limiting
	BandwidthLimitKbps int `json:"bandwidth_limit_kbps,omitempty"` // Limit bandwidth in KB/s

	// Response Fuzzing (Mutation)
	ResponseFuzzing *FuzzingConfig `json:"response_fuzzing,omitempty"`

	// HTTP Specific
	ModifyHeaders map[string]string `json:"modify_headers,omitempty"` // Headers to add/modify
	RemoveHeaders []string          `json:"remove_headers,omitempty"` // Headers to remove
}

// FuzzingConfig defines how to mutate response bodies
type FuzzingConfig struct {
	Enabled      bool    `json:"enabled"`
	Probability  float64 `json:"probability"`   // 0.0 to 1.0 (likelihood of fuzzing a valid response)
	MutationRate float64 `json:"mutation_rate"` // 0.0 to 1.0 (percentage of fields to mutate)
}

// ToJSON converts ChaosConfig to JSON string
func (c *ChaosConfig) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON creates ChaosConfig from JSON string
func FromJSON(data string) (*ChaosConfig, error) {
	var config ChaosConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Validate checks if the configuration is valid
func (c *ChaosConfig) Validate() error {
	if c.Target == "" {
		return ErrInvalidTarget
	}
	if c.Rules.InjectFailureRate < 0 || c.Rules.InjectFailureRate > 1 {
		return ErrInvalidFailureRate
	}
	if c.Rules.DropConnectionRate < 0 || c.Rules.DropConnectionRate > 1 {
		return ErrInvalidDropRate
	}
	return nil
}

// Custom errors
var (
	ErrInvalidTarget      = &ValidationError{Message: "target URL is required"}
	ErrInvalidFailureRate = &ValidationError{Message: "inject_failure_rate must be between 0 and 1"}
	ErrInvalidDropRate    = &ValidationError{Message: "drop_connection_rate must be between 0 and 1"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
