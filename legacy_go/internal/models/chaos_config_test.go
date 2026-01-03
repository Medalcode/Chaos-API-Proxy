package models

import (
	"testing"
)

func TestChaosConfigValidation(t *testing.T) {
	t.Run("Valid config should pass", func(t *testing.T) {
		config := &ChaosConfig{
			Target: "https://api.example.com",
			Rules: ChaosRules{
				InjectFailureRate: 0.5,
				DropConnectionRate: 0.1,
			},
		}

		if err := config.Validate(); err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("Empty target should fail", func(t *testing.T) {
		config := &ChaosConfig{
			Target: "",
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for empty target")
		}
	})

	t.Run("Invalid failure rate should fail", func(t *testing.T) {
		config := &ChaosConfig{
			Target: "https://api.example.com",
			Rules: ChaosRules{
				InjectFailureRate: 1.5, // Invalid: > 1.0
			},
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for invalid failure rate")
		}
	})

	t.Run("Negative failure rate should fail", func(t *testing.T) {
		config := &ChaosConfig{
			Target: "https://api.example.com",
			Rules: ChaosRules{
				InjectFailureRate: -0.1,
			},
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for negative failure rate")
		}
	})

	t.Run("Invalid drop rate should fail", func(t *testing.T) {
		config := &ChaosConfig{
			Target: "https://api.example.com",
			Rules: ChaosRules{
				DropConnectionRate: 1.5,
			},
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for invalid drop rate")
		}
	})
}

func TestChaosConfigSerialization(t *testing.T) {
	t.Run("Should serialize and deserialize correctly", func(t *testing.T) {
		original := &ChaosConfig{
			ID:          "test-id",
			Name:        "Test Config",
			Description: "Test description",
			Target:      "https://api.example.com",
			Enabled:     true,
			Rules: ChaosRules{
				LatencyMs:         500,
				Jitter:            100,
				InjectFailureRate: 0.1,
				ErrorCode:         503,
			},
		}

		// Serialize
		jsonStr, err := original.ToJSON()
		if err != nil {
			t.Fatalf("Failed to serialize: %v", err)
		}

		// Deserialize
		deserialized, err := FromJSON(jsonStr)
		if err != nil {
			t.Fatalf("Failed to deserialize: %v", err)
		}

		// Compare
		if deserialized.ID != original.ID {
			t.Errorf("ID mismatch: expected %s, got %s", original.ID, deserialized.ID)
		}
		if deserialized.Name != original.Name {
			t.Errorf("Name mismatch: expected %s, got %s", original.Name, deserialized.Name)
		}
		if deserialized.Target != original.Target {
			t.Errorf("Target mismatch: expected %s, got %s", original.Target, deserialized.Target)
		}
		if deserialized.Rules.LatencyMs != original.Rules.LatencyMs {
			t.Errorf("LatencyMs mismatch: expected %d, got %d",
				original.Rules.LatencyMs, deserialized.Rules.LatencyMs)
		}
	})
}
