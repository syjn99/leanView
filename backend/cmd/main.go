package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/syjn99/leanView/backend/db"
	"github.com/syjn99/leanView/backend/indexer"
	"github.com/syjn99/leanView/backend/server"
	"github.com/syjn99/leanView/backend/types"
	"github.com/syjn99/leanView/backend/utils"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	flag.Parse()

	// Initialize logger
	logger := utils.NewLogger()
	logger.Infof("Starting PQ Devnet Visualizer backend...")

	// Setup graceful shutdown context
	ctx, cancel := setupSignalHandling(logger)
	defer cancel()

	// Parse config file
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}

	// Initialize database instances
	db.InitDB(&cfg.Database)

	indexerInstance := indexer.NewIndexer(cfg, logger.WithField("service", "indexer"))
	serverInstance := server.NewServer(indexerInstance, logger.WithField("service", "http"))

	go func() {
		if err := indexerInstance.Start(ctx); err != nil {
			logger.WithError(err).Fatalf("Indexer error")
		}
	}()

	if err := serverInstance.Start(ctx); err != nil {
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
