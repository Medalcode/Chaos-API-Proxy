package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/medalcode/chaos-api-proxy/internal/chaos"
	"github.com/medalcode/chaos-api-proxy/internal/storage"
	log "github.com/sirupsen/logrus"
)

// ProxyHandler handles proxy requests with chaos injection
type ProxyHandler struct {
	storage *storage.RedisStorage
	engine  *chaos.Engine
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(storage *storage.RedisStorage) *ProxyHandler {
	return &ProxyHandler{
		storage: storage,
		engine:  chaos.NewEngine(),
	}
}

// ServeHTTP implements http.Handler interface
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	configID := vars["configID"]
	
	// Support for header-based config identification
	// If configID is empty from path, try to get it from header
	if configID == "" {
		configID = r.Header.Get("X-Chaos-Config-ID")
		if configID == "" {
			http.Error(w, "Missing configuration ID. Use path /proxy/{configID}/... or header X-Chaos-Config-ID", http.StatusBadRequest)
			return
		}
	}

	// Get configuration
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	config, err := h.storage.GetConfig(ctx, configID)
	if err != nil {
		log.WithError(err).WithField("config_id", configID).Error("Failed to get config")
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}

	// Check if config is enabled
	if !config.Enabled {
		log.WithField("config_id", configID).Warn("Config is disabled")
		http.Error(w, "Configuration is disabled", http.StatusForbidden)
		return
	}

	// Make chaos decision
	decision := h.engine.MakeDecision(config.Rules)

	log.WithFields(log.Fields{
		"config_id":           configID,
		"target":              config.Target,
		"inject_latency":      decision.ShouldInjectLatency,
		"inject_error":        decision.ShouldInjectError,
		"drop_connection":     decision.ShouldDropConnection,
	}).Info("Processing proxy request")

	// Handle drop connection
	if decision.ShouldDropConnection {
		log.WithField("config_id", configID).Info("Dropping connection")
		// Close connection without response
		if hijacker, ok := w.(http.Hijacker); ok {
			conn, _, err := hijacker.Hijack()
			if err == nil {
				conn.Close()
				return
			}
		}
		// Fallback if hijacking fails
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Handle error injection
	if decision.ShouldInjectError {
		log.WithFields(log.Fields{
			"config_id":  configID,
			"error_code": decision.ErrorCode,
		}).Info("Injecting error")

		// Add chaos headers
		for k, v := range decision.ModifyHeaders {
			w.Header().Set(k, v)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(decision.ErrorCode)
		w.Write([]byte(decision.ErrorBody))
		return
	}

	// Handle latency injection
	if decision.ShouldInjectLatency {
		log.WithFields(log.Fields{
			"config_id": configID,
			"latency":   decision.LatencyDuration,
		}).Info("Injecting latency")
		time.Sleep(decision.LatencyDuration)
	}

	// Parse target URL
	targetURL, err := url.Parse(config.Target)
	if err != nil {
		log.WithError(err).Error("Failed to parse target URL")
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Customize director to preserve path and modify headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Remove the proxy prefix from path (only if path-based)
		// Path format: /proxy/{configID}/actual/path
		if vars["configID"] != "" {
			prefix := fmt.Sprintf("/proxy/%s", configID)
			req.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
		} else {
			// Header-based: use path as-is
			req.URL.Path = r.URL.Path
		}

		// Apply header modifications
		for k, v := range decision.ModifyHeaders {
			req.Header.Set(k, v)
		}

		// Remove headers
		for _, h := range decision.RemoveHeaders {
			req.Header.Del(h)
		}

		// Set proper host
		req.Host = targetURL.Host

		log.WithFields(log.Fields{
			"original_path": r.URL.Path,
			"proxied_path":  req.URL.Path,
			"target_host":   req.Host,
		}).Debug("Proxying request")
	}

	// Customize ModifyResponse to add chaos headers to response
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("X-Chaos-Proxy", "true")
		resp.Header.Set("X-Chaos-Proxy-Config-ID", configID)
		return nil
	}

	// Error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.WithError(err).Error("Proxy error")
		w.Header().Set("X-Chaos-Proxy-Error", "true")
		http.Error(w, "Proxy error", http.StatusBadGateway)
	}

	// Bandwidth limiting wrapper (if configured)
	if config.Rules.BandwidthLimitKbps > 0 {
		w = &bandwidthLimitedWriter{
			ResponseWriter: w,
			limitKbps:      config.Rules.BandwidthLimitKbps,
			engine:         h.engine,
		}
	}

	// Execute proxy
	proxy.ServeHTTP(w, r)
}

// bandwidthLimitedWriter wraps http.ResponseWriter to limit bandwidth
type bandwidthLimitedWriter struct {
	http.ResponseWriter
	limitKbps int
	engine    *chaos.Engine
}

func (w *bandwidthLimitedWriter) Write(data []byte) (int, error) {
	// Calculate delay based on bandwidth limit
	delay := w.engine.CalculateBandwidthDelay(len(data), w.limitKbps)
	if delay > 0 {
		time.Sleep(delay)
	}
	return w.ResponseWriter.Write(data)
}

// Ensure we implement all necessary interfaces
func (w *bandwidthLimitedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *bandwidthLimitedWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Implement io.ReaderFrom if underlying writer implements it
func (w *bandwidthLimitedWriter) ReadFrom(r io.Reader) (int64, error) {
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(&bandwidthLimitedReader{
			reader:    r,
			limitKbps: w.limitKbps,
			engine:    w.engine,
		})
	}
	// Fallback to copying manually with bandwidth limiting
	return io.Copy(w, r)
}

// bandwidthLimitedReader wraps io.Reader to limit bandwidth
type bandwidthLimitedReader struct {
	reader    io.Reader
	limitKbps int
	engine    *chaos.Engine
}

func (r *bandwidthLimitedReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		delay := r.engine.CalculateBandwidthDelay(n, r.limitKbps)
		if delay > 0 {
			time.Sleep(delay)
		}
	}
	return n, err
}
