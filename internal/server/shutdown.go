// Graceful shutdown implementation
//
// Handles clean termination:
// - Stop accepting new connections
// - Wait for active requests to complete
// - Timeout for lingering connections
// - Resource cleanup

package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdown handles graceful shutdown of a server when receiving termination signals
type GracefulShutdown struct {
	server          *Server
	shutdownTimeout time.Duration
	signals         []os.Signal
}

// NewGracefulShutdown creates a new graceful shutdown handler for the given server
func NewGracefulShutdown(server *Server, timeout time.Duration) *GracefulShutdown {
	return &GracefulShutdown{
		server:          server,
		shutdownTimeout: timeout,
		signals:         []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	}
}

// WithSignals sets the signals that will trigger a shutdown
func (gs *GracefulShutdown) WithSignals(signals ...os.Signal) *GracefulShutdown {
	gs.signals = signals
	return gs
}

// HandleShutdown starts listening for signals and performs graceful shutdown when received
func (gs *GracefulShutdown) HandleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, gs.signals...)

	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal %s, starting graceful shutdown\n", sig)

		ctx, cancel := context.WithTimeout(context.Background(), gs.shutdownTimeout)
		defer cancel()

		if err := gs.server.Stop(ctx); err != nil {
			fmt.Printf("Error during server shutdown: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Graceful shutdown completed")
		os.Exit(0)
	}()
}

// WaitForShutdown blocks until a shutdown signal is received
func (gs *GracefulShutdown) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, gs.signals...)

	// Wait for signal
	sig := <-sigChan
	fmt.Printf("Received signal %s, starting graceful shutdown\n", sig)

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), gs.shutdownTimeout)
	defer cancel()

	// Attempt to shut down gracefully
	if err := gs.server.Stop(ctx); err != nil {
		fmt.Printf("Error during server shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Graceful shutdown completed")
}