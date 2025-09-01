package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/syjn99/leanView/backend/indexer"
	"github.com/syjn99/leanView/backend/server"
	"github.com/syjn99/leanView/backend/utils"
)

func main() {
	// Initialize logger
	logger := utils.NewLogger()
	logger.Infof("Starting PQ Devnet Visualizer backend...")

	// Setup graceful shutdown context
	ctx, cancel := setupSignalHandling(logger)
	defer cancel()

	server := server.NewServer(logger.WithField("service", "http"))
	indexer := indexer.NewIndexer(logger.WithField("service", "indexer"))

	go func() {
		if err := indexer.Start(ctx); err != nil {
			logger.WithError(err).Fatalf("Indexer error")
		}
	}()

	if err := server.Start(ctx); err != nil {
		logger.WithError(err).Fatalf("Server error")
	}

	logger.Infof("PQ Devnet Visualizer backend terminated")
}

// setupSignalHandling creates a context that cancels on interrupt signals
func setupSignalHandling(logger logrus.FieldLogger) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		logger.Infof("Received interrupt signal, initiating graceful shutdown...")
		cancel()
	}()

	return ctx, cancel
}
