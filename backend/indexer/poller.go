package indexer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/syjn99/leanView/backend/types"
)

// BlockPoller continuously polls endpoints for new blocks based on slot timing
type BlockPoller struct {
	clientPool     *ClientPool
	blockProcessor *BlockProcessor

	// Polling configuration
	pollInterval time.Duration
	maxRetries   int
	retryDelay   time.Duration

	// State tracking
	lastProcessedSlot uint64
	isRunning         bool

	// Synchronization
	ticker      *time.Ticker
	stopChannel chan bool
	mutex       sync.RWMutex

	logger logrus.FieldLogger
}

// NewBlockPoller creates a new block poller with slot-based timing
func NewBlockPoller(clientPool *ClientPool, blockProcessor *BlockProcessor, logger logrus.FieldLogger) *BlockPoller {
	return &BlockPoller{
		clientPool:     clientPool,
		blockProcessor: blockProcessor,
		pollInterval:   defaultPollingInterval, // 4 seconds per slot
		maxRetries:     defaultMaxRetries,
		retryDelay:     defaultRetryDelay,
		stopChannel:    make(chan bool, 1),
		logger:         logger.WithField("component", "block_poller"),
	}
}

// Start begins the polling process with slot-based timing
func (bp *BlockPoller) Start(ctx context.Context) error {
	bp.mutex.Lock()
	if bp.isRunning {
		bp.mutex.Unlock()
		return fmt.Errorf("poller is already running")
	}
	bp.isRunning = true
	bp.ticker = time.NewTicker(bp.pollInterval)
	bp.mutex.Unlock()

	// Start the polling goroutine
	go bp.pollLoop(ctx)

	bp.logger.WithField("poll_interval", bp.pollInterval).Info("Block poller started")
	return nil
}

// Stop gracefully stops the polling process
func (bp *BlockPoller) Stop() error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	if !bp.isRunning {
		return nil // Already stopped
	}

	bp.isRunning = false

	if bp.ticker != nil {
		bp.ticker.Stop()
	}

	// Signal stop to polling goroutine
	select {
	case bp.stopChannel <- true:
	default: // Non-blocking if channel is full
	}

	bp.logger.Info("Block poller stopped")
	return nil
}

// pollLoop is the main polling loop that runs in a goroutine
func (bp *BlockPoller) pollLoop(ctx context.Context) {
	defer func() {
		if bp.ticker != nil {
			bp.ticker.Stop()
		}
	}()

	for {
		select {
		case <-bp.ticker.C:
			if err := bp.pollForNewBlocks(ctx); err != nil {
				bp.logger.WithError(err).Warn("Failed to poll for new blocks")
			}
		case <-bp.stopChannel:
			bp.logger.Debug("Received stop signal, exiting poll loop")
			return
		case <-ctx.Done():
			bp.logger.Debug("Context cancelled, exiting poll loop")
			return
		}
	}
}

// pollForNewBlocks fetches the latest head block and checks for new slots
func (bp *BlockPoller) pollForNewBlocks(ctx context.Context) error {
	// Get a healthy client from the pool
	client := bp.clientPool.GetHealthyClient()
	if client == nil {
		return fmt.Errorf("no healthy clients available")
	}

	// Fetch the current head block
	headBlock, err := bp.fetchHeadBlockWithRetry(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to fetch head block: %w", err)
	}

	// Check if this is a new slot
	if headBlock.Slot > bp.lastProcessedSlot {
		bp.logger.WithFields(logrus.Fields{
			"new_slot":      headBlock.Slot,
			"previous_slot": bp.lastProcessedSlot,
			"slot_gap":      headBlock.Slot - bp.lastProcessedSlot,
		}).Info("New block detected")

		// Process the detected new block using the block processor
		if err := bp.blockProcessor.ProcessBlock(ctx, headBlock); err != nil {
			bp.logger.WithError(err).WithField("slot", headBlock.Slot).Error("Failed to process new block")
			// Continue and update the slot even if processing failed to avoid getting stuck
		}

		bp.updateLastProcessedSlot(headBlock.Slot)
	} else {
		bp.logger.WithField("current_slot", headBlock.Slot).Debug("No new blocks")
	}

	return nil
}

// fetchHeadBlockWithRetry attempts to fetch the head block with retry logic
func (bp *BlockPoller) fetchHeadBlockWithRetry(ctx context.Context, client *Client) (*types.BlockHeader, error) {
	var lastErr error

	for attempt := 0; attempt < bp.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(bp.retryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			// Try to get a different healthy client
			if newClient := bp.clientPool.GetHealthyClient(); newClient != nil {
				client = newClient
			}
		}

		block, err := client.GetLatestBlock(ctx)
		if err != nil {
			lastErr = err
			bp.logger.WithError(err).WithField("attempt", attempt+1).Warn("Failed to fetch head block")
			continue
		}

		if attempt > 0 {
			bp.logger.WithField("attempt", attempt+1).Info("Successfully fetched head block after retry")
		}

		return block, nil
	}

	return nil, fmt.Errorf("failed to fetch head block after %d attempts: %w", bp.maxRetries, lastErr)
}

// updateLastProcessedSlot safely updates the last processed slot
func (bp *BlockPoller) updateLastProcessedSlot(slot uint64) {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	bp.lastProcessedSlot = slot
}

// GetLastProcessedSlot returns the last processed slot number
func (bp *BlockPoller) GetLastProcessedSlot() uint64 {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()
	return bp.lastProcessedSlot
}

// IsRunning returns whether the poller is currently running
func (bp *BlockPoller) IsRunning() bool {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()
	return bp.isRunning
}
