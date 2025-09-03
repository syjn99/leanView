package indexer

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/syjn99/leanView/backend/types"
)

type Indexer struct {
	config     *types.Config
	clientPool *ClientPool
	poller     *BlockPoller
	logger     logrus.FieldLogger
}

func NewIndexer(config *types.Config, logger logrus.FieldLogger) *Indexer {
	// Create client pool from endpoint configuration
	clientPool := NewClientPool(config.LeanApi.Endpoints, logger)

	// Create block poller
	poller := NewBlockPoller(clientPool, logger)

	return &Indexer{
		config:     config,
		clientPool: clientPool,
		poller:     poller,
		logger:     logger,
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
