package indexer

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/syjn99/leanView/backend/types"
)

type Indexer struct {
	config         *types.Config
	clientPool     *ClientPool
	blockProcessor *BlockProcessor
	poller         *BlockPoller
	headCache      *HeadCache
	logger         logrus.FieldLogger
}

func NewIndexer(config *types.Config, logger logrus.FieldLogger) *Indexer {
	// Create client pool from endpoint configuration
	clientPool := NewClientPool(config.LeanApi.Endpoints, logger)

	// Create head cache
	headCache := NewHeadCache(logger)

	// Create block processor
	blockProcessor := NewBlockProcessor(headCache, logger)

	// Create block poller with processor
	poller := NewBlockPoller(clientPool, blockProcessor, logger)

	return &Indexer{
		config:         config,
		clientPool:     clientPool,
		blockProcessor: blockProcessor,
		poller:         poller,
		headCache:      headCache,
		logger:         logger,
	}
}

func (i *Indexer) Start(ctx context.Context) error {
	i.logger.Info("Indexer starting...")

	// Start client health checking
	i.clientPool.RunHealthChecks(ctx)

	// Start block polling
	if err := i.poller.Start(ctx); err != nil {
		return fmt.Errorf("failed to start block poller: %w", err)
	}

	i.logger.WithFields(logrus.Fields{
		"client_count": i.clientPool.GetClientCount(),
		"endpoints":    len(i.config.LeanApi.Endpoints),
	}).Info("Indexer started successfully")

	// Wait for context cancellation
	<-ctx.Done()
	return i.Stop()
}

func (i *Indexer) Stop() error {
	i.logger.Info("Indexer stopping...")

	// Stop block polling
	if err := i.poller.Stop(); err != nil {
		i.logger.WithError(err).Warn("Error stopping block poller")
	}

	// Stop client health checking
	i.clientPool.StopHealthChecks()

	i.logger.Info("Indexer stopped successfully")
	return nil
}

// GetHeadCache returns the head cache for external access
func (i *Indexer) GetHeadCache() *HeadCache {
	return i.headCache
}

// GetClientPool returns the client pool for external access
func (i *Indexer) GetClientPool() *ClientPool {
	return i.clientPool
}
