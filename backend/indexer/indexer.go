package indexer

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Indexer struct {
	logger logrus.FieldLogger
}

func NewIndexer(logger logrus.FieldLogger) *Indexer {
	return &Indexer{
		logger: logger,
	}
}

func (i *Indexer) Start(ctx context.Context) error {
	i.logger.Info("Indexer started")

	errChan := make(chan error, 1)

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		return i.Stop()
	case err := <-errChan:
		return err
	}
}

func (i *Indexer) Stop() error {
	i.logger.Info("Indexer stopped")

	return nil
}
