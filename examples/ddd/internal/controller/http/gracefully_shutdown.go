package http

import (
	"context"
	"errors"
	"github.com/a-aslani/wotop"
	"github.com/a-aslani/wotop/wotop_logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// gracefullyShutdown handles the HTTP server with a graceful shutdown mechanism.
type gracefullyShutdown struct {
	httpServer *http.Server        // The HTTP server instance.
	log        wotop_logger.Logger // Logger for logging server events.
}

// NewGracefullyShutdown creates a new instance of gracefullyShutdown.
//
// Parameters:
//   - log: The logger instance for logging server events.
//   - handler: The HTTP handler to process incoming requests.
//   - address: The address on which the server will listen.
//
// Returns:
//
//	A wotop.ControllerStarter instance for starting the server.
func NewGracefullyShutdown(log wotop_logger.Logger, handler http.Handler, address string) wotop.ControllerStarter {
	return &gracefullyShutdown{
		httpServer: &http.Server{
			Addr:    address,
			Handler: handler,
		},
		log: log,
	}
}

// Start begins the HTTP server and listens for termination signals to shut down gracefully.
//
// The method starts the server in a separate goroutine and listens for SIGINT or SIGTERM signals.
// Upon receiving a termination signal, it shuts down the server with a timeout of 5 seconds.
func (r *gracefullyShutdown) Start() {

	// Start the HTTP server in a separate goroutine.
	go func() {
		if err := r.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			r.log.Error(context.Background(), "listen: %s", err)
			os.Exit(1)
		}
	}()

	// Log that the server is running.
	r.log.Info(context.Background(), "server is running at %v", r.httpServer.Addr)

	// Create a channel to listen for OS signals.
	quit := make(chan os.Signal, 1)

	// Notify the channel on SIGINT or SIGTERM signals.
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received.

	// Log that the server is shutting down.
	r.log.Info(context.Background(), "Shutting down server...")

	// Create a context with a timeout for the shutdown process.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server.
	if err := r.httpServer.Shutdown(ctx); err != nil {
		r.log.Error(context.Background(), "Server forced to shutdown: %v", err.Error())
		os.Exit(1)
	}

	// Log that the server has stopped.
	r.log.Info(context.Background(), "Server stopped.")
}
