// Main HTTP server implementation
//
// Responsibilities:
// - HTTP server setup and configuration
// - Route registration and handler binding
// - Middleware application
// - Server lifecycle management
// - Connection handling optimizations

package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// Server represents the HTTP server for the application
type Server struct {
	server   *http.Server
	listener net.Listener
	options  Options
	router   http.Handler
	mu       sync.Mutex
	started  bool
	stopped  bool
	err      error
}

// New creates a new server instance with the given options and router
func New(options Options, router http.Handler) *Server {
	return &Server{
		options: options,
		router:  router,
	}
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return errors.New("server already started")
	}
	s.started = true
	s.mu.Unlock()

	// Create HTTP server
	s.server = &http.Server{
		Handler:        s.router,
		ReadTimeout:    s.options.ReadTimeout,
		WriteTimeout:   s.options.WriteTimeout,
		IdleTimeout:    s.options.IdleTimeout,
		MaxHeaderBytes: s.options.MaxHeaderBytes,
		TLSConfig:      s.options.TLSConfig,
	}

	// Create listener
	var err error
	s.listener, err = net.Listen("tcp", s.options.Address)
	if err != nil {
		s.err = fmt.Errorf("failed to create listener: %w", err)
		return s.err
	}

	// Start server
	go func() {
		var err error
		if s.options.TLSConfig != nil {
			err = s.server.ServeTLS(s.listener, "", "")
		} else {
			err = s.server.Serve(s.listener)
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.mu.Lock()
			s.err = err
			s.mu.Unlock()
		}
	}()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return errors.New("server not started")
	}
	if s.stopped {
		s.mu.Unlock()
		return errors.New("server already stopped")
	}
	s.stopped = true
	s.mu.Unlock()

	return s.server.Shutdown(ctx)
}

// Addr returns the server's address
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.options.Address
}

// Error returns any error that occurred during server operation
func (s *Server) Error() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}