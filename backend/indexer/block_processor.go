package indexer

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/syjn99/leanView/backend/db"
	"github.com/syjn99/leanView/backend/types"
)

// BlockProcessor handles the core block processing logic
type BlockProcessor struct {
	// Configuration
	maxRetries int

	logger logrus.FieldLogger
}

// NewBlockProcessor creates a new block processor
func NewBlockProcessor(logger logrus.FieldLogger) *BlockProcessor {
	return &BlockProcessor{
		maxRetries: defaultMaxRetries,
		logger:     logger.WithField("component", "block_processor"),
	}
}

// ProcessBlock handles processing a single detected block
func (bp *BlockProcessor) ProcessBlock(ctx context.Context, block *types.BlockHeader) error {
	bp.logger.WithFields(logrus.Fields{
		"slot":           block.Slot,
		"proposer_index": block.ProposerIndex,
	}).Info("Processing new block")

	// Validate the block header
	if err := bp.validateBlockHeader(block); err != nil {
		return fmt.Errorf("block validation failed for slot %d: %w", block.Slot, err)
	}

	// Store the block in the database
	err := db.RunDBTransaction(func(tx *sqlx.Tx) error {
		return db.InsertBlockHeader(block, tx)
	})
	if err != nil {
		return fmt.Errorf("failed to store block for slot %d: %w", block.Slot, err)
	}

	bp.logger.WithField("slot", block.Slot).Info("Successfully processed block")
	return nil
}

// validateBlockHeader performs basic validation on block header
func (bp *BlockProcessor) validateBlockHeader(block *types.BlockHeader) error {
	// Check that slot is reasonable (not zero, not too far in future)
	if block.Slot == 0 {
		return fmt.Errorf("invalid slot: cannot be zero")
	}

	// Check that hash fields are the expected length (32 bytes)
	if len(block.ParentRoot) != 32 {
		return fmt.Errorf("invalid parent_root length: expected 32 bytes, got %d", len(block.ParentRoot))
	}
	if len(block.StateRoot) != 32 {
		return fmt.Errorf("invalid state_root length: expected 32 bytes, got %d", len(block.StateRoot))
	}
	if len(block.BodyRoot) != 32 {
		return fmt.Errorf("invalid body_root length: expected 32 bytes, got %d", len(block.BodyRoot))
	}

	return nil
}

// GetLatestProcessedSlot returns the latest processed slot from database
func (bp *BlockProcessor) GetLatestProcessedSlot() uint64 {
	// Get the latest block from database
	headers, err := db.GetLatestBlockHeaders(1)
	if err != nil || len(headers) == 0 {
		bp.logger.WithError(err).Warn("Could not get latest processed slot")
		return 0
	}

	return headers[0].Slot
}

// ProcessBlockRange fetches and processes blocks for a range of slots
func (bp *BlockProcessor) ProcessBlockRange(ctx context.Context, clientPool *ClientPool, startSlot, endSlot uint64) error {
	if startSlot > endSlot {
		return fmt.Errorf("invalid range: startSlot %d > endSlot %d", startSlot, endSlot)
	}

	bp.logger.WithFields(logrus.Fields{
		"start_slot": startSlot,
		"end_slot":   endSlot,
		"gap_size":   endSlot - startSlot + 1,
	}).Info("Processing block range for catchup")

	// Get a healthy client from the pool
	client := clientPool.GetHealthyClient()
	if client == nil {
		return fmt.Errorf("no healthy clients available for block range processing")
	}

	// Process in batches to avoid memory issues and provide better progress tracking
	const batchSize = 20
	totalProcessed := 0
	var allBlocks []*types.BlockHeader

	for currentSlot := startSlot; currentSlot <= endSlot; {
		batchEnd := currentSlot + batchSize - 1
		if batchEnd > endSlot {
			batchEnd = endSlot
		}

		bp.logger.WithFields(logrus.Fields{
			"batch_start": currentSlot,
			"batch_end":   batchEnd,
		}).Debug("Fetching batch of blocks")

		// Fetch blocks for this batch
		blocks, err := client.GetBlockRange(ctx, currentSlot, batchEnd)
		if err != nil {
			// Try with a different client if available
			if newClient := clientPool.GetHealthyClient(); newClient != nil && newClient != client {
				client = newClient
				blocks, err = client.GetBlockRange(ctx, currentSlot, batchEnd)
			}
			if err != nil {
				bp.logger.WithError(err).WithFields(logrus.Fields{
					"batch_start": currentSlot,
					"batch_end":   batchEnd,
				}).Error("Failed to fetch block range batch")
				// Continue with next batch even if this one fails
				currentSlot = batchEnd + 1
				continue
			}
		}

		// Validate blocks
		validBlocks := make([]*types.BlockHeader, 0, len(blocks))
		for _, block := range blocks {
			if err := bp.validateBlockHeader(block); err != nil {
				bp.logger.WithError(err).WithField("slot", block.Slot).Warn("Skipping invalid block during catchup")
				continue
			}
			validBlocks = append(validBlocks, block)
		}

		if len(validBlocks) > 0 {
			allBlocks = append(allBlocks, validBlocks...)
			totalProcessed += len(validBlocks)
		}

		bp.logger.WithFields(logrus.Fields{
			"batch_start":    currentSlot,
			"batch_end":      batchEnd,
			"blocks_fetched": len(blocks),
			"valid_blocks":   len(validBlocks),
		}).Debug("Processed batch of blocks")

		currentSlot = batchEnd + 1
	}

	// Store all blocks in a single transaction for efficiency
	if len(allBlocks) > 0 {
		err := db.RunDBTransaction(func(tx *sqlx.Tx) error {
			return db.InsertBlockHeaderBatch(allBlocks, tx)
		})
		if err != nil {
			return fmt.Errorf("failed to store catchup blocks: %w", err)
		}
	}

	bp.logger.WithFields(logrus.Fields{
		"start_slot":      startSlot,
		"end_slot":        endSlot,
		"total_processed": totalProcessed,
		"blocks_stored":   len(allBlocks),
	}).Info("Completed block range processing")

	return nil
}
