package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal cuenta el total de peticiones procesadas
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_proxy_requests_total",
			Help: "The total number of processed requests",
		},
		[]string{"config_id", "status_code", "chaos_type"},
	)

	// RequestDuration mide la latencia de las peticiones
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_proxy_request_duration_seconds",
			Help:    "The duration of requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"config_id", "chaos_type"},
	)

	// ActiveConfigs cuenta cuántas configuraciones están activas
	ActiveConfigs = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "chaos_proxy_active_configs",
			Help: "The number of currently active chaos configurations",
		},
	)

	// ChaosInjections cuenta cuántas veces se inyectó caos específicamente
	ChaosInjections = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_proxy_injections_total",
			Help: "The total number of chaos injections performed",
		},
		[]string{"config_id", "injection_type"}, // type: latency, error, drop, bandwidth
	)
)
