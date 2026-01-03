package chaos

import (
	"math/rand"
	"time"

	"github.com/medalcode/chaos-api-proxy/internal/models"
)

// Engine handles chaos injection logic
type Engine struct {
	rng *rand.Rand
}

// NewEngine creates a new chaos engine
func NewEngine() *Engine {
	source := rand.NewSource(time.Now().UnixNano())
	return &Engine{
		rng: rand.New(source),
	}
}

// Decision represents a chaos decision
type Decision struct {
	ShouldInjectLatency bool
	LatencyDuration     time.Duration
	ShouldInjectError   bool
	ErrorCode           int
	ErrorBody           string
	ShouldDropConnection bool
	ModifyHeaders       map[string]string
	RemoveHeaders       []string
}

// MakeDecision determines what chaos to inject based on rules
func (e *Engine) MakeDecision(rules models.ChaosRules) *Decision {
	decision := &Decision{
		ModifyHeaders: make(map[string]string),
		RemoveHeaders: make([]string, 0),
	}

	// Check if we should drop connection
	if rules.DropConnection || (rules.DropConnectionRate > 0 && e.rng.Float64() < rules.DropConnectionRate) {
		decision.ShouldDropConnection = true
		return decision // No need to check other rules if dropping connection
	}

	// Check if we should inject error
	if rules.InjectFailureRate > 0 && e.rng.Float64() < rules.InjectFailureRate {
		decision.ShouldInjectError = true
		decision.ErrorCode = rules.ErrorCode
		if decision.ErrorCode == 0 {
			decision.ErrorCode = 500 // Default to 500
		}
		decision.ErrorBody = rules.ErrorBody
		if decision.ErrorBody == "" {
			decision.ErrorBody = `{"error": "Chaos Engineering: Injected failure"}`
		}
		// Add header to indicate this was injected
		decision.ModifyHeaders["X-Chaos-Proxy-Injected"] = "true"
		decision.ModifyHeaders["X-Chaos-Proxy-Type"] = "error"
		return decision
	}

	// Calculate latency
	if rules.LatencyMs > 0 {
		decision.ShouldInjectLatency = true
		latency := rules.LatencyMs

		// Add jitter if specified
		if rules.Jitter > 0 {
			// Random value between -jitter and +jitter
			jitterValue := e.rng.Intn(rules.Jitter*2+1) - rules.Jitter
			latency += jitterValue
		}

		// Ensure non-negative latency
		if latency < 0 {
			latency = 0
		}

		decision.LatencyDuration = time.Duration(latency) * time.Millisecond
	}

	// Copy headers to modify
	if rules.ModifyHeaders != nil {
		for k, v := range rules.ModifyHeaders {
			decision.ModifyHeaders[k] = v
		}
	}

	// Add chaos proxy headers
	decision.ModifyHeaders["X-Chaos-Proxy"] = "true"
	if decision.ShouldInjectLatency {
		decision.ModifyHeaders["X-Chaos-Proxy-Latency-Ms"] = time.Duration(decision.LatencyDuration).String()
	}

	// Copy headers to remove
	if rules.RemoveHeaders != nil {
		decision.RemoveHeaders = append(decision.RemoveHeaders, rules.RemoveHeaders...)
	}

	return decision
}

// CalculateBandwidthDelay calculates delay based on bandwidth limit
func (e *Engine) CalculateBandwidthDelay(bytesRead int, limitKbps int) time.Duration {
	if limitKbps <= 0 {
		return 0
	}

	// Calculate how long it should take to transfer these bytes
	bytesPerSecond := limitKbps * 1024
	seconds := float64(bytesRead) / float64(bytesPerSecond)
	return time.Duration(seconds * float64(time.Second))
}
