package metricsfiber

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
	"time"
)

type Config struct {
	Next                   func(c *fiber.Ctx) bool
	Registry               prometheus.Registerer
	Namespace              string
	SubSystem              string
	Labels                 prometheus.Labels
	SkipPaths              []string
	DisableMeasureInflight bool
	DisableMeasureSize     bool
}

var ConfigDefault = Config{
	Next:      nil,
	Namespace: "http",
}

func configDefault(config ...Config) Config {
	if len(config) < 1 {
		return ConfigDefault
	}

	cfg := config[0]

	if cfg.Next == nil {
		cfg.Next = ConfigDefault.Next
	}

	if cfg.Namespace == "" {
		cfg.Namespace = ConfigDefault.Namespace
	}

	return cfg
}

type HttpMetrics struct {
	cfg             Config
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
	requestInFlight *prometheus.GaugeVec
	cacheCounter    *prometheus.CounterVec
	skipPathsMap    map[string]struct{}
}

func New(config ...Config) *HttpMetrics {
	cfg := configDefault(config...)

	if cfg.Registry == nil {
		cfg.Registry = prometheus.DefaultRegisterer
	}

	var skip = map[string]struct{}{}
	for _, path := range cfg.SkipPaths {
		skip[path] = struct{}{}
	}

	requestsTotal := promauto.With(cfg.Registry).NewCounterVec(
		prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(cfg.Namespace, cfg.SubSystem, "requests_total"),
			Help:        "Count all http requests by status code, method and path.",
			ConstLabels: cfg.Labels,
		},
		[]string{"status_code", "method", "path"},
	)

	requestDuration := promauto.With(cfg.Registry).NewHistogramVec(prometheus.HistogramOpts{
		Name:        prometheus.BuildFQName(cfg.Namespace, cfg.SubSystem, "request_duration_seconds"),
		Help:        "Duration of all HTTP requests by status code, method and path.",
		ConstLabels: cfg.Labels,
		Buckets: []float64{
			0.000000001, // 1ns
			0.000000002,
			0.000000005,
			0.00000001, // 10ns
			0.00000002,
			0.00000005,
			0.0000001, // 100ns
			0.0000002,
			0.0000005,
			0.000001, // 1µs
			0.000002,
			0.000005,
			0.00001, // 10µs
			0.00002,
			0.00005,
			0.0001, // 100µs
			0.0002,
			0.0005,
			0.001, // 1ms
			0.002,
			0.005,
			0.01, // 10ms
			0.02,
			0.05,
			0.1, // 100 ms
			0.2,
			0.5,
			1.0, // 1s
			2.0,
			5.0,
			10.0, // 10s
			15.0,
			20.0,
			30.0,
		},
	},
		[]string{"status_code", "method", "path"},
	)

	responseSize := promauto.With(cfg.Registry).NewHistogramVec(prometheus.HistogramOpts{
		Name:        prometheus.BuildFQName(cfg.Namespace, cfg.SubSystem, "response_size_bytes"),
		Help:        "The size of the HTTP responses by status code, method and path.",
		ConstLabels: cfg.Labels,
		Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
	},
		[]string{"status_code", "method", "path"},
	)

	requestInFlight := promauto.With(cfg.Registry).NewGaugeVec(prometheus.GaugeOpts{
		Name:        prometheus.BuildFQName(cfg.Namespace, cfg.SubSystem, "requests_in_progress_total"),
		Help:        "All the requests in progress",
		ConstLabels: cfg.Labels,
	}, []string{"method"})

	return &HttpMetrics{
		cfg:             cfg,
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
		responseSize:    responseSize,
		requestInFlight: requestInFlight,
		skipPathsMap:    skip,
	}
}

func (h *HttpMetrics) Serve(c *fiber.Ctx) (err error) {
	if _, exists := h.skipPathsMap[string(c.Request().RequestURI())]; exists {
		return c.Next()
	}

	start := time.Now()

	if !h.cfg.DisableMeasureInflight {
		method := c.Route().Method
		h.requestInFlight.WithLabelValues(method).Inc()
		defer h.requestInFlight.WithLabelValues(method).Dec()
	}

	defer func(start time.Time, cfg Config) {
		status := c.Response().StatusCode()
		if err != nil {
			var e *fiber.Error
			switch {
			case errors.As(err, &e):
				status = e.Code
			default:
				status = fiber.StatusInternalServerError
			}
		}

		statusCode := strconv.Itoa(status)
		method := c.Route().Method
		path := c.Route().Path

		h.requestsTotal.WithLabelValues(statusCode, method, path).Inc()

		elapsed := float64(time.Since(start).Nanoseconds()) / 1e9
		h.requestDuration.WithLabelValues(statusCode, method, path).Observe(elapsed)

		if !cfg.DisableMeasureSize {
			sizeBytes := len(c.Response().Body())
			h.responseSize.WithLabelValues(statusCode, method, path).Observe(float64(sizeBytes))
		}
	}(start, h.cfg)

	return c.Next()
}
