package db

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/syjn99/leanView/backend/types"
)

// Write Operations (with transactions)

// InsertBlockHeader inserts a new block header into the database
func InsertBlockHeader(header *types.BlockHeader, tx *sqlx.Tx) error {
	_, err := tx.Exec(`
		INSERT OR REPLACE INTO block_headers (
			slot, proposer_index, parent_root, state_root, body_root
		) VALUES (?, ?, ?, ?, ?)`,
		header.Slot, header.ProposerIndex, header.ParentRoot, header.StateRoot, header.BodyRoot)
	if err != nil {
		return fmt.Errorf("error inserting block header for slot %d: %w", header.Slot, err)
	}
	return nil
}

// UpdateBlockHeader updates an existing block header in the database
func UpdateBlockHeader(header *types.BlockHeader, tx *sqlx.Tx) error {
	result, err := tx.Exec(`
		UPDATE block_headers 
		SET proposer_index = ?, parent_root = ?, state_root = ?, body_root = ?
		WHERE slot = ?`,
		header.ProposerIndex, header.ParentRoot, header.StateRoot, header.BodyRoot, header.Slot)
	if err != nil {
		return fmt.Errorf("error updating block header for slot %d: %w", header.Slot, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected for slot %d: %w", header.Slot, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no block header found for slot %d", header.Slot)
	}

	return nil
}

// DeleteBlockHeader deletes a block header by slot
func DeleteBlockHeader(slot uint64, tx *sqlx.Tx) error {
	result, err := tx.Exec(`DELETE FROM block_headers WHERE slot = ?`, slot)
	if err != nil {
		return fmt.Errorf("error deleting block header for slot %d: %w", slot, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected for slot %d: %w", slot, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no block header found for slot %d", slot)
	}

	return nil
}

// InsertBlockHeaderBatch inserts multiple block headers in a single transaction
func InsertBlockHeaderBatch(headers []*types.BlockHeader, tx *sqlx.Tx) error {
	if len(headers) == 0 {
		return nil
	}

	stmt, err := tx.Preparex(`
		INSERT OR REPLACE INTO block_headers (
			slot, proposer_index, parent_root, state_root, body_root
		) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("error preparing batch insert statement: %w", err)
	}
	defer stmt.Close()

	for _, header := range headers {
		_, err := stmt.Exec(header.Slot, header.ProposerIndex, header.ParentRoot, header.StateRoot, header.BodyRoot)
		if err != nil {
			return fmt.Errorf("error inserting block header for slot %d in batch: %w", header.Slot, err)
		}
	}

	return nil
}

// Read Operations (direct ReaderDb)

// GetBlockHeaderBySlot retrieves a block header by its slot number
func GetBlockHeaderBySlot(slot uint64) (*types.BlockHeader, error) {
	header := &types.BlockHeader{}
	err := ReaderDb.Get(header, `
		SELECT slot, proposer_index, parent_root, state_root, body_root
		FROM block_headers
		WHERE slot = ?`, slot)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching block header by slot %d: %w", slot, err)
	}
	return header, nil
}

// GetBlockHeaderByRoot retrieves a block header by one of its roots (parent, state, or body)
func GetBlockHeaderByRoot(root []byte) (*types.BlockHeader, error) {
	header := &types.BlockHeader{}
	err := ReaderDb.Get(header, `
		SELECT slot, proposer_index, parent_root, state_root, body_root
		FROM block_headers
		WHERE parent_root = ? OR state_root = ? OR body_root = ?
		ORDER BY slot DESC
		LIMIT 1`, root, root, root)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching block header by root: %w", err)
	}
	return header, nil
}

// GetBlockHeadersByProposer retrieves block headers by proposer index with a limit
func GetBlockHeadersByProposer(proposerIndex uint64, limit int) ([]*types.BlockHeader, error) {
	headers := []*types.BlockHeader{}
	err := ReaderDb.Select(&headers, `
		SELECT slot, proposer_index, parent_root, state_root, body_root
		FROM block_headers
		WHERE proposer_index = ?
		ORDER BY slot DESC
		LIMIT ?`, proposerIndex, limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching block headers by proposer %d: %w", proposerIndex, err)
	}
	return headers, nil
}

// GetLatestBlockHeaders retrieves the most recent block headers with a limit
func GetLatestBlockHeaders(limit int) ([]*types.BlockHeader, error) {
	headers := []*types.BlockHeader{}
	err := ReaderDb.Select(&headers, `
		SELECT slot, proposer_index, parent_root, state_root, body_root
		FROM block_headers
		ORDER BY slot DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest block headers: %w", err)
	}
	return headers, nil
}

// GetBlockHeadersInRange retrieves block headers within a slot range (inclusive)
func GetBlockHeadersInRange(startSlot, endSlot uint64) ([]*types.BlockHeader, error) {
	if startSlot > endSlot {
		return nil, fmt.Errorf("start slot %d cannot be greater than end slot %d", startSlot, endSlot)
	}

	headers := []*types.BlockHeader{}
	err := ReaderDb.Select(&headers, `
		SELECT slot, proposer_index, parent_root, state_root, body_root
		FROM block_headers
		WHERE slot >= ? AND slot <= ?
		ORDER BY slot ASC`, startSlot, endSlot)
	if err != nil {
		return nil, fmt.Errorf("error fetching block headers in range %d-%d: %w", startSlot, endSlot, err)
	}
	return headers, nil
}
