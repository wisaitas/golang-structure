package promx

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	ServiceName string
	Namespace   string
}

type metrics struct {
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
}

func newMetrics(cfg Config) *metrics {
	ns := cfg.Namespace
	if ns == "" {
		ns = "http"
	}

	return &metrics{
		requestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace:   ns,
			Name:        "requests_total",
			Help:        "Total number of HTTP requests",
			ConstLabels: prometheus.Labels{"service": cfg.ServiceName},
		}, []string{"method", "path", "status_code"}),

		requestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   ns,
			Name:        "request_duration_seconds",
			Help:        "HTTP request duration in seconds",
			ConstLabels: prometheus.Labels{"service": cfg.ServiceName},
			Buckets:     []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"method", "path", "status_code"}),

		requestsInFlight: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace:   ns,
			Name:        "requests_in_flight",
			Help:        "Number of HTTP requests currently being processed",
			ConstLabels: prometheus.Labels{"service": cfg.ServiceName},
		}),
	}
}

func NewMiddleware(app *fiber.App, cfg Config) fiber.Handler {
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	m := newMetrics(cfg)

	return func(c fiber.Ctx) error {
		m.requestsInFlight.Inc()
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Response().StatusCode())
		method := c.Method()
		path := c.Route().Path

		m.requestsTotal.WithLabelValues(method, path, statusCode).Inc()
		m.requestDuration.WithLabelValues(method, path, statusCode).Observe(duration)
		m.requestsInFlight.Dec()

		return err
	}
}
