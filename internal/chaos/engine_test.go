package chaos

import (
	"testing"
	"time"

	"github.com/medalcode/chaos-api-proxy/internal/models"
)

func TestEngineMakeDecision(t *testing.T) {
	engine := NewEngine()

	t.Run("Should inject latency with fixed value", func(t *testing.T) {
		rules := models.ChaosRules{
			LatencyMs: 500,
		}

		decision := engine.MakeDecision(rules)

		if !decision.ShouldInjectLatency {
			t.Error("Expected latency injection")
		}
		if decision.LatencyDuration != 500*time.Millisecond {
			t.Errorf("Expected 500ms latency, got %v", decision.LatencyDuration)
		}
	})

	t.Run("Should inject latency with jitter", func(t *testing.T) {
		rules := models.ChaosRules{
			LatencyMs: 500,
			Jitter:    100,
		}

		decision := engine.MakeDecision(rules)

		if !decision.ShouldInjectLatency {
			t.Error("Expected latency injection")
		}

		// Latency should be between 400ms and 600ms
		minLatency := 400 * time.Millisecond
		maxLatency := 600 * time.Millisecond

		if decision.LatencyDuration < minLatency || decision.LatencyDuration > maxLatency {
			t.Errorf("Expected latency between %v and %v, got %v",
				minLatency, maxLatency, decision.LatencyDuration)
		}
	})

	t.Run("Should inject error based on failure rate", func(t *testing.T) {
		rules := models.ChaosRules{
			InjectFailureRate: 1.0, // Always fail
			ErrorCode:         503,
		}

		decision := engine.MakeDecision(rules)

		if !decision.ShouldInjectError {
			t.Error("Expected error injection with 100% failure rate")
		}
		if decision.ErrorCode != 503 {
			t.Errorf("Expected error code 503, got %d", decision.ErrorCode)
		}
	})

	t.Run("Should drop connection", func(t *testing.T) {
		rules := models.ChaosRules{
			DropConnection: true,
		}

		decision := engine.MakeDecision(rules)

		if !decision.ShouldDropConnection {
			t.Error("Expected connection drop")
		}
	})

	t.Run("Should modify headers", func(t *testing.T) {
		rules := models.ChaosRules{
			ModifyHeaders: map[string]string{
				"X-Custom-Header": "test-value",
			},
		}

		decision := engine.MakeDecision(rules)

		if decision.ModifyHeaders["X-Custom-Header"] != "test-value" {
			t.Error("Expected custom header to be set")
		}
		if decision.ModifyHeaders["X-Chaos-Proxy"] != "true" {
			t.Error("Expected X-Chaos-Proxy header to be set")
		}
	})
}

func TestCalculateBandwidthDelay(t *testing.T) {
	engine := NewEngine()

	t.Run("Should calculate delay for bandwidth limit", func(t *testing.T) {
		// 100 KB/s limit, 10240 bytes (10 KB) read
		delay := engine.CalculateBandwidthDelay(10240, 100)

		// Should take ~0.1 seconds for 10 KB at 100 KB/s
		expectedDelay := 100 * time.Millisecond
		tolerance := 10 * time.Millisecond

		if delay < expectedDelay-tolerance || delay > expectedDelay+tolerance {
			t.Errorf("Expected delay around %v, got %v", expectedDelay, delay)
		}
	})

	t.Run("Should return zero delay when no limit", func(t *testing.T) {
		delay := engine.CalculateBandwidthDelay(10240, 0)
		if delay != 0 {
			t.Errorf("Expected zero delay with no limit, got %v", delay)
		}
	})
}
