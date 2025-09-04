package indexer

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/syjn99/leanView/backend/types"
)

const (
	// MaxRecentBlocks is the maximum number of recent blocks to keep in cache
	MaxRecentBlocks = 32
)

// CacheStats represents head cache statistics
type CacheStats struct {
	RecentBlocksCount int    `json:"recent_blocks_count"`
	MaxRecentBlocks   int    `json:"max_recent_blocks"`
	HasCurrentHead    bool   `json:"has_current_head"`
	HasJustified      bool   `json:"has_justified"`
	HasFinalized      bool   `json:"has_finalized"`
	CurrentHeadSlot   uint64 `json:"current_head_slot,omitempty"`
	JustifiedSlot     uint64 `json:"justified_slot,omitempty"`
	FinalizedSlot     uint64 `json:"finalized_slot,omitempty"`
}

// HeadCache maintains current chain head state aligned with Lean consensus
type HeadCache struct {
	// Current chain head
	currentHead *types.BlockHeader

	// Lean consensus checkpoints (from 3SF mini)
	latestJustified *types.Checkpoint // Latest justified checkpoint
	latestFinalized *types.Checkpoint // Latest finalized checkpoint

	// Recent blocks for fork choice (LMD-GHOST)
	recentBlocks map[string]*types.BlockHeader // root hex -> block

	// Synchronization
	mutex sync.RWMutex

	logger logrus.FieldLogger
}

// NewHeadCache creates a new head cache
func NewHeadCache(logger logrus.FieldLogger) *HeadCache {
	return &HeadCache{
		recentBlocks: make(map[string]*types.BlockHeader),
		logger:       logger.WithField("component", "head_cache"),
	}
}

// UpdateHead updates the current head block and maintains recent blocks cache
func (hc *HeadCache) UpdateHead(block *types.BlockHeader) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.currentHead = block

	// Calculate proper block root using SSZ
	blockRoot, err := block.HashTreeRoot()
	if err != nil {
		hc.logger.WithError(err).WithField("slot", block.Slot).Error("Failed to calculate block root")
		return
	}

	// Add to recent blocks cache using proper block root
	rootHex := fmt.Sprintf("%x", blockRoot)
	hc.recentBlocks[rootHex] = block

	// Prune old blocks if we exceed the limit
	if len(hc.recentBlocks) > MaxRecentBlocks {
		hc.pruneOldBlocks()
	}

	hc.logger.WithFields(logrus.Fields{
		"slot":       block.Slot,
		"block_root": rootHex[:8] + "...", // Log first 8 chars
	}).Debug("Updated head cache with new block")
}

// GetCurrentHead returns the current head block (thread-safe)
func (hc *HeadCache) GetCurrentHead() *types.BlockHeader {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	return hc.currentHead
}

// UpdateJustified updates the latest justified checkpoint
func (hc *HeadCache) UpdateJustified(checkpoint *types.Checkpoint) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.latestJustified = checkpoint

	rootHex := fmt.Sprintf("%x", checkpoint.Root)
	hc.logger.WithFields(logrus.Fields{
		"slot": checkpoint.Slot,
		"root": rootHex[:8] + "...",
	}).Info("Updated justified checkpoint")
}

// UpdateFinalized updates the latest finalized checkpoint
func (hc *HeadCache) UpdateFinalized(checkpoint *types.Checkpoint) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.latestFinalized = checkpoint

	rootHex := fmt.Sprintf("%x", checkpoint.Root)
	hc.logger.WithFields(logrus.Fields{
		"slot": checkpoint.Slot,
		"root": rootHex[:8] + "...",
	}).Info("Updated finalized checkpoint")
}

// GetJustifiedCheckpoint returns the latest justified checkpoint
func (hc *HeadCache) GetJustifiedCheckpoint() *types.Checkpoint {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	return hc.latestJustified
}

// GetFinalizedCheckpoint returns the latest finalized checkpoint
func (hc *HeadCache) GetFinalizedCheckpoint() *types.Checkpoint {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	return hc.latestFinalized
}

// GetRecentBlocks returns a copy of recent blocks for fork choice
func (hc *HeadCache) GetRecentBlocks() map[string]*types.BlockHeader {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	// Return a copy to avoid race conditions
	recent := make(map[string]*types.BlockHeader, len(hc.recentBlocks))
	for root, block := range hc.recentBlocks {
		recent[root] = block
	}
	return recent
}

// AddRecentBlock adds a block to the recent blocks cache
func (hc *HeadCache) AddRecentBlock(block *types.BlockHeader) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	// Calculate proper block root using SSZ
	blockRoot, err := block.HashTreeRoot()
	if err != nil {
		hc.logger.WithError(err).WithField("slot", block.Slot).Error("Failed to calculate block root for recent blocks")
		return
	}

	rootHex := fmt.Sprintf("%x", blockRoot)
	hc.recentBlocks[rootHex] = block

	if len(hc.recentBlocks) > MaxRecentBlocks {
		hc.pruneOldBlocks()
	}
}

// GetCacheStats returns cache statistics for monitoring
func (hc *HeadCache) GetCacheStats() *CacheStats {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	stats := &CacheStats{
		RecentBlocksCount: len(hc.recentBlocks),
		MaxRecentBlocks:   MaxRecentBlocks,
		HasCurrentHead:    hc.currentHead != nil,
		HasJustified:      hc.latestJustified != nil,
		HasFinalized:      hc.latestFinalized != nil,
	}

	if hc.currentHead != nil {
		stats.CurrentHeadSlot = hc.currentHead.Slot
	}
	if hc.latestJustified != nil {
		stats.JustifiedSlot = hc.latestJustified.Slot
	}
	if hc.latestFinalized != nil {
		stats.FinalizedSlot = hc.latestFinalized.Slot
	}

	return stats
}

// pruneOldBlocks removes the oldest blocks to maintain cache size
// Must be called with mutex already locked
func (hc *HeadCache) pruneOldBlocks() {
	if len(hc.recentBlocks) <= MaxRecentBlocks {
		return
	}

	// Find the oldest blocks by slot number
	var oldestSlot uint64 = ^uint64(0) // Max uint64
	var oldestRoot string

	for root, block := range hc.recentBlocks {
		if block.Slot < oldestSlot {
			oldestSlot = block.Slot
			oldestRoot = root
		}
	}

	// Remove the oldest block
	delete(hc.recentBlocks, oldestRoot)

	hc.logger.WithFields(logrus.Fields{
		"pruned_slot":       oldestSlot,
		"remaining_blocks":  len(hc.recentBlocks),
		"max_recent_blocks": MaxRecentBlocks,
	}).Debug("Pruned old block from head cache")
}
