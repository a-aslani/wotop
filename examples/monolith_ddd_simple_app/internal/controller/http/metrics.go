package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterMetrics sets up Prometheus metrics for the HTTP server.
//
// Parameters:
//   - serviceName: The name of the service for which metrics are being registered.
//
// This function registers a `/metrics` endpoint for Prometheus to scrape metrics.
// It also initializes a request counter and a latency histogram for monitoring HTTP requests.
func (r *controller) RegisterMetrics(serviceName string) {

	// Register the `/metrics` endpoint to expose Prometheus metrics.
	r.Router.GET("/metrics", prometheusHandler())

	// Initialize a Prometheus counter to track the number of HTTP requests.
	r.reqCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "http_request_counter",
		Name:      serviceName,
		Help:      fmt.Sprintf("Count of request to the %s service", serviceName),
	})

	// Initialize a Prometheus histogram to measure the latency of HTTP requests.
	r.reqLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "http_request_latency",
		Name:      serviceName,
		Buckets:   []float64{0.1, 0.5, 1.0},
	})

}

// prometheusHandler returns a Gin middleware handler for serving Prometheus metrics.
//
// This function wraps the Prometheus HTTP handler and integrates it with the Gin framework.
func prometheusHandler() gin.HandlerFunc {
	// Create the Prometheus HTTP handler.
	h := promhttp.Handler()
	return func(c *gin.Context) {
		// Serve the Prometheus metrics using the HTTP handler.
		h.ServeHTTP(c.Writer, c.Request)
	}
}
