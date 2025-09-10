package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	"github.com/syjn99/leanView/backend/gen/proto/api/v1/apiv1connect"
	"github.com/syjn99/leanView/backend/indexer"
	"github.com/syjn99/leanView/backend/services/block"
	"github.com/syjn99/leanView/backend/services/monitoring"
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
	logger     logrus.FieldLogger
	indexer    *indexer.Indexer
	httpServer *http.Server
}

func NewServer(indexer *indexer.Indexer, logger logrus.FieldLogger) *Server {
	mux := http.NewServeMux()

	// Root handler for basic info
	mux.HandleFunc("/", rootHandler(logger))

	// Health check endpoint for container orchestration
	mux.HandleFunc("/health", healthHandler(logger))

	// Create Block service
	blockService := block.NewBlockService(indexer, logger.(*logrus.Entry).Logger)

	// Register Block service Connect RPC handler
	blockPath, blockHandler := apiv1connect.NewBlockServiceHandler(
		blockService,
		connect.WithInterceptors(
			newLoggingInterceptor(logger),
		),
	)
	mux.Handle(blockPath, blockHandler)

	// Create Monitoring service
	monitoringService := monitoring.NewMonitoringService(indexer, logger.(*logrus.Entry).Logger)

	// Register Monitoring service Connect RPC handler
	monitoringPath, monitoringHandler := apiv1connect.NewMonitoringServiceHandler(
		monitoringService,
		connect.WithInterceptors(
			newLoggingInterceptor(logger),
		),
	)
	mux.Handle(monitoringPath, monitoringHandler)

	// Add CORS for frontend access (Vite dev server)
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(mux)

	httpServer := &http.Server{
		Addr:         ":" + defaultPort,
		Handler:      corsHandler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		indexer:    indexer,
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

// newLoggingInterceptor creates a logging interceptor for Connect RPC
func newLoggingInterceptor(logger logrus.FieldLogger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()

			// Call the handler
			resp, err := next(ctx, req)

			// Log the request
			logger.WithFields(logrus.Fields{
				"method":   req.Spec().Procedure,
				"duration": time.Since(start).Milliseconds(),
				"error":    err != nil,
			}).Debug("RPC request handled")

			return resp, err
		}
	}
}

// rootHandler provides basic service information
func rootHandler(logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("Root request from %s", r.RemoteAddr)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := `{"service":"PQ Devnet Visualizer","version":"0.1.0","endpoints":["/health","/api.v1.BlockService/GetLatestBlockHeader","/api.v1.MonitoringService/GetAllClientsHeads"]}`
		if _, err := w.Write([]byte(response)); err != nil {
			logger.Errorf("Error writing root response: %v", err)
		}
	}
}

// healthHandler provides health status for container orchestration
func healthHandler(logger logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple health check - server is running
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := `{"status":"healthy","service":"PQ Devnet Visualizer"}`
		if _, err := w.Write([]byte(response)); err != nil {
			logger.Errorf("Error writing health response: %v", err)
		}
	}
}
