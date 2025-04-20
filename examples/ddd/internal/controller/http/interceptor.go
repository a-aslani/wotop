package http

import (
	"context"
	"github.com/a-aslani/wotop/wotop_logger"
	"github.com/a-aslani/wotop/wotop_util"
	"github.com/gin-gonic/gin"
	"time"
)

// metricRequestCounter is a middleware that increments the request counter metric
// for each incoming HTTP request.
func (r *controller) metricRequestCounter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment the request counter metric.
		r.reqCounter.Inc()
		// Proceed to the next middleware or handler.
		c.Next()
	}
}

// metricLatencyRequestChecker is a middleware that measures the latency of each
// HTTP request and records it in the latency metric.
func (r *controller) metricLatencyRequestChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record the start time of the request.
		t := time.Now()

		// Proceed to the next middleware or handler.
		c.Next()

		// Calculate the latency of the request.
		latency := time.Since(t)

		// Record the latency in the latency metric.
		r.reqLatency.Observe(latency.Seconds())
	}
}

// authentication is a middleware that sets authentication-related data in the
// request context, such as trace ID, data, ID, role, and expiration time.
func (r *controller) authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a unique trace ID for the request.
		traceID := wotop_util.GenerateID(16)

		// Set the trace ID in the logger context.
		ctx := wotop_logger.SetTraceID(context.Background(), traceID)
		_ = ctx // The context is not used further in this function.

		// Set authentication-related data in the Gin context.
		c.Set("Data", "DATA")
		c.Set("ID", "ID")
		c.Set("Role", "ROLE")
		c.Set("ExpiredAt", "_")

		// Proceed to the next middleware or handler.
		c.Next()
	}
}
