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
