package http

// RegisterRouter sets up the HTTP routes for the application.
// It defines a resource group with middleware for metrics and latency checks,
// and registers the v1 API endpoints, including an authenticated POST endpoint
// for managing affiliates.
func (r *controller) RegisterRouter() {

	// Create a resource group with middleware for metrics and latency checks
	resource := r.Router.Group(r.proxyPath, r.metricRequestCounter(), r.metricLatencyRequestChecker())

	// APPLICATION API
	// Define the v1 API group
	v1 := resource.Group("/v1")

	// Register a POST endpoint for affiliates with authentication middleware
	v1.POST("/affiliates", r.authentication())
}
