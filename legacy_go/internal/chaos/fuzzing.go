package chaos

import (
	"encoding/json"
	
	"github.com/medalcode/chaos-api-proxy/internal/models"
)

// ShouldFuzz determines if we should fuzz this response
func (e *Engine) ShouldFuzz(rules models.ChaosRules) bool {
	if rules.ResponseFuzzing == nil || !rules.ResponseFuzzing.Enabled {
		return false
	}
	return e.rng.Float64() < rules.ResponseFuzzing.Probability
}

// FuzzBody mutates the response body based on rules
// Returns mutated body and true if mutation occurred
func (e *Engine) FuzzBody(body []byte, rules models.ChaosRules) ([]byte, bool) {
	if len(body) == 0 {
		return body, false
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Not a JSON body, skip intelligent fuzzing
		return body, false
	}

	// Determine mutation rate
	rate := 0.1 // default
	if rules.ResponseFuzzing != nil && rules.ResponseFuzzing.MutationRate > 0 {
		rate = rules.ResponseFuzzing.MutationRate
	}

	mutated := e.mutateValue(data, rate)
	
	newBody, err := json.Marshal(mutated)
	if err != nil {
		return body, false
	}

	return newBody, true
}

func (e *Engine) mutateValue(v interface{}, rate float64) interface{} {
	// Recursive mutation logic
	switch val := v.(type) {
	case map[string]interface{}:
		for k, innerV := range val {
			val[k] = e.mutateValue(innerV, rate)
		}
		return val
	case []interface{}:
		for i, innerV := range val {
			val[i] = e.mutateValue(innerV, rate)
		}
		return val
	default:
		// Leaf node: Check if we should mutate this specific field
		if e.rng.Float64() < rate {
			return e.applyMutation(val)
		}
		return val
	}
}

func (e *Engine) applyMutation(v interface{}) interface{} {
	// 4 types of mutations:
	// 0. Nullify
	// 1. Type swap
	// 2. Value corruption
	// 3. Bit flip (for logic bugs)

	choice := e.rng.Intn(4)

	switch choice {
	case 0: 
		return nil // Nullify
	case 1:
		// Type swap
		switch v.(type) {
		case string: return 12345
		case float64: return "should_be_number"
		case bool: return 0
		case nil: return "was_null"
		default: return "swapped_type"
		}
	case 2:
		// Value corruption
		switch val := v.(type) {
		case string: return val + "_CHAOS"
		case float64: return val * 9999
		case bool: return !val
		default: return v
		}
	case 3:
		// Edge cases
		switch v.(type) {
		case float64: return -1
		case string: return "" // Empty string
		default: return nil
		}
	}
	return v
}
