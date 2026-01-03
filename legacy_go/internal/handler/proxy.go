```go
package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/medalcode/chaos-api-proxy/internal/chaos"
	"github.com/medalcode/chaos-api-proxy/internal/metrics"
	"github.com/medalcode/chaos-api-proxy/internal/storage"
	log "github.com/sirupsen/logrus"
)

// ProxyHandler handles proxy requests with chaos injection
type ProxyHandler struct {
	storage *storage.RedisStorage
	engine  *chaos.Engine
}

// ... NewProxyHandler ...

// responseWriterWrapper captures the status code
// ...

// ServeHTTP implements http.Handler interface
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	configID := vars["configID"]
	chaosType := "none"

	// Support for header-based config identification
	if configID == "" {
		configID = r.Header.Get("X-Chaos-Config-ID")
		if configID == "" {
			http.Error(w, "Missing configuration ID. Use path /proxy/{configID}/... or header X-Chaos-Config-ID", http.StatusBadRequest)
			return
		}
	}

	// Wrap the response writer to capture status code
	wrapper := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
	
	// Ensure metrics and LOGS are recorded at the end
	defer func() {
		duration := time.Since(start)
		
		// 1. Prometheus Metrics
		metrics.RequestsTotal.WithLabelValues(configID, strconv.Itoa(wrapper.statusCode), chaosType).Inc()
		metrics.RequestDuration.WithLabelValues(configID, chaosType).Observe(duration.Seconds())
		
		// 2. Persistent Logs (Tracing)
		// Only log if we have a configID (valid request attempt)
		if configID != "" {
			reqLog := &storage.RequestLog{
				ID:         uuid.New().String(),
				Timestamp:  start,
				ConfigID:   configID,
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: wrapper.statusCode,
				DurationMs: duration.Milliseconds(),
				ChaosType:  chaosType,
			}
			
			// Fire and forget logging (don't block response on Redis error)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()
				if err := h.storage.LogRequest(ctx, reqLog); err != nil {
					log.WithError(err).Error("Failed to trace request")
				}
			}()
		}
	}()

	// Get configuration
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	config, err := h.storage.GetConfig(ctx, configID)
	if err != nil {
		log.WithError(err).WithField("config_id", configID).Error("Failed to get config")
		wrapper.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Configuration not found"))
		return
	}

	// Check if config is enabled
	if !config.Enabled {
		wrapper.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Configuration is disabled"))
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

	// Determine chaos type for metrics
	if decision.ShouldDropConnection {
		chaosType = "drop_connection"
		metrics.ChaosInjections.WithLabelValues(configID, "drop_connection").Inc()
	} else if decision.ShouldInjectError {
		chaosType = "error"
		metrics.ChaosInjections.WithLabelValues(configID, "error").Inc()
	} else if decision.ShouldInjectLatency {
		// Latency can happen with explicit error or success, but if it's the main feature:
		if chaosType == "none" {
			chaosType = "latency"
		}
		metrics.ChaosInjections.WithLabelValues(configID, "latency").Inc()
	}
	
	if config.Rules.BandwidthLimitKbps > 0 {
		metrics.ChaosInjections.WithLabelValues(configID, "bandwidth_limit").Inc()
	}

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
		wrapper.WriteHeader(http.StatusServiceUnavailable)
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
		wrapper.WriteHeader(decision.ErrorCode)
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
		wrapper.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Invalid target URL"))
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

	// Customize ModifyResponse to add chaos headers and Fuzz body
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("X-Chaos-Proxy", "true")
		resp.Header.Set("X-Chaos-Proxy-Config-ID", configID)

		// Check if we should fuzz response
		if h.engine.ShouldFuzz(config.Rules) {
			// Read body
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.WithError(err).Warn("Failed to read response body for fuzzing")
				return nil // Return original
			}
			resp.Body.Close()

			// Fuzz
			newBody, mutated := h.engine.FuzzBody(body, config.Rules)
			
			// Always restore body reader (whether mutated or not) to avoid "body closed" errors
			resp.Body = ioutil.NopCloser(bytes.NewReader(newBody))

			if mutated {
				resp.ContentLength = int64(len(newBody))
				resp.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
				resp.Header.Set("X-Chaos-Proxy-Fuzzed", "true")
				
				// Update metrics
				metrics.ChaosInjections.WithLabelValues(configID, "response_fuzzing").Inc()
				
				log.WithField("config_id", configID).Info("🔥 Fuzzed response body")
			}
		}
		return nil
	}

	// Error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.WithError(err).Error("Proxy error")
		w.Header().Set("X-Chaos-Proxy-Error", "true")
		wrapper.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Proxy error"))
	}

	// Bandwidth limiting wrapper (if configured)
	if config.Rules.BandwidthLimitKbps > 0 {
		// Note: We use the original 'w' here because bandwidthLimitedWriter needs to wrap the underlying ResponseWriter
		// But passing 'wrapper' is also fine as it implements ResponseWriter
		limitedWriter := &bandwidthLimitedWriter{
			ResponseWriter: wrapper, // Use wrapper to capture writes
			limitKbps:      config.Rules.BandwidthLimitKbps,
			engine:         h.engine,
		}
		proxy.ServeHTTP(limitedWriter, r)
		return
	}

	// Execute proxy with wrapper
	proxy.ServeHTTP(wrapper, r)
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
