package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultPort     = "8080"
	readTimeout     = 15 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 30 * time.Second
)

// Server represents the HTTP server with Connect RPC support
type Server struct {
	logger logrus.FieldLogger

	httpServer *http.Server
}

func NewServer(logger logrus.FieldLogger) *Server {
	mux := http.NewServeMux()

	// Root handler for basic info
	mux.HandleFunc("/", rootHandler(logger))

	httpServer := &http.Server{
		Addr:         ":" + defaultPort,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
	}
}

// Start begins serving HTTP requests
func (s *Server) Start(ctx context.Context) error {
	s.logger.Infof("Starting PQ Devnet Visualizer server on %s", s.httpServer.Addr)

	// Channel to capture server errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		return s.shutdown()
	case err := <-errChan:
		return err
	}
}

// shutdown gracefully shuts down the server
func (s *Server) shutdown() error {
	s.logger.Infof("Shutting down server gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Printf("Server forced to shutdown: %v", err)
		return err
	}

	s.logger.Infof("Server shutdown complete")
	return nil
}

// rootHandler provides basic service information
func rootHandler(logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("Root request from %s", r.RemoteAddr)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := `{"service":"PQ Devnet Visualizer","version":"0.1.0","endpoints":["/health"]}`
		if _, err := w.Write([]byte(response)); err != nil {
			logger.Errorf("Error writing root response: %v", err)
		}
	}
}
