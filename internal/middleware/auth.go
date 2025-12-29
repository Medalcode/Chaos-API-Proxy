package middleware

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// AuthMiddleware handles API key validation
type AuthMiddleware struct {
	apiKeys map[string]bool
	enabled bool
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(keys string) *AuthMiddleware {
	keyMap := make(map[string]bool)
	enabled := false

	if keys != "" {
		enabled = true
		for _, k := range strings.Split(keys, ",") {
			key := strings.TrimSpace(k)
			if key != "" {
				keyMap[key] = true
			}
		}
	}

	return &AuthMiddleware{
		apiKeys: keyMap,
		enabled: enabled,
	}
}

// Handler implements the middleware logic
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If auth is disabled, skip check
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Check header
		clientKey := r.Header.Get("X-API-Key")
		
		// Also check query param for easier browser testing if needed
		if clientKey == "" {
			clientKey = r.URL.Query().Get("api_key")
		}

		if clientKey == "" || !m.apiKeys[clientKey] {
			log.WithFields(log.Fields{
				"ip":     r.RemoteAddr,
				"method": r.Method,
				"path":   r.URL.Path,
			}).Warn("Unauthorized API access attempt")
			
			http.Error(w, "Unauthorized: Invalid or missing API Key", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
