package http

import (
	"fmt"
	"github.com/a-aslani/wotop"
	"github.com/a-aslani/wotop/examples/monolith_ddd_simple_app/configs"
	"github.com/a-aslani/wotop/jwt"
	"github.com/a-aslani/wotop/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

// controller represents the HTTP controller for the application.
// It includes the router, logger, configuration, Token handler, and metrics for monitoring.
type controller struct {
	wotop.ControllerStarter                      // Embeds the ControllerStarter interface for starting the controller.
	wotop.UsecaseRegisterer                      // Embeds the UsecaseRegisterer interface for registering use cases.
	Router                  *gin.Engine          // The Gin router instance for handling HTTP requests.
	log                     logger.Logger        // Logger for logging application events.
	cfg                     *configs.Config      // Configuration settings for the application.
	jwt                     jwt.Token            // Token handler for managing JSON Web Tokens.
	reqCounter              prometheus.Counter   // Prometheus counter for tracking HTTP request counts.
	reqLatency              prometheus.Histogram // Prometheus histogram for measuring request latency.
	proxyPath               string               // Proxy path for the application.
	appName                 string               // Name of the application.
}

// NewController creates a new instance of the HTTP controller.
//
// Parameters:
//   - appData: Application data containing metadata about the app.
//   - log: Logger instance for logging application events.
//   - cfg: Configuration settings for the application.
//   - jwt: Token handler for managing JSON Web Tokens.
//
// Returns:
//
//	A wotop.ControllerRegisterer instance for registering the controller.
func NewController(appData wotop.ApplicationData, log logger.Logger, cfg *configs.Config, jwt jwt.Token) wotop.ControllerRegisterer {

	// Create a new Gin router instance.
	router := gin.Default()

	// PING API
	// Define a ping endpoint to check the health of the application.
	router.GET(fmt.Sprintf("%s/ping", cfg.Servers[appData.AppName].ProxyPath), func(c *gin.Context) {
		c.JSON(http.StatusOK, appData)
	})

	// CORS
	// Configure CORS middleware to allow cross-origin requests.
	router.Use(cors.New(cors.Config{
		ExposeHeaders:   []string{"Data-Length"},
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		AllowAllOrigins: true,
		AllowHeaders:    []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		MaxAge:          12 * time.Hour,
	}))

	// Static file serving
	// Serve static files for uploads based on the application name.
	router.Static(fmt.Sprintf("/%s/%s/%s", cfg.Servers[appData.AppName].ProxyPath, "uploads", appData.AppName), fmt.Sprintf("./uploads/%s", appData.AppName))

	// Retrieve the server address from the configuration.
	address := cfg.Servers[appData.AppName].Address

	// Return a new controller instance with the configured router and dependencies.
	return &controller{
		ControllerStarter: NewGracefullyShutdown(log, router, address),
		UsecaseRegisterer: wotop.NewBaseController(),
		Router:            router,
		log:               log,
		cfg:               cfg,
		jwt:               jwt,
		proxyPath:         cfg.Servers[appData.AppName].ProxyPath,
		appName:           appData.AppName,
	}
}
