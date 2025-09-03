package types

import (
	"encoding/json"
	"fmt"
)

// BlockHeader represents a Lean Block header
type BlockHeader struct {
	Slot          uint64 `db:"slot" json:"slot"`
	ProposerIndex uint64 `db:"proposer_index" json:"proposer_index"`
	ParentRoot    []byte `db:"parent_root" json:"parent_root"`
	StateRoot     []byte `db:"state_root" json:"state_root"`
	BodyRoot      []byte `db:"body_root" json:"body_root"`
}

// blockHeaderJSON is used for JSON marshaling/unmarshaling with hex strings
type blockHeaderJSON struct {
	Slot          uint64 `json:"slot"`
	ProposerIndex uint64 `json:"proposer_index"`
	ParentRoot    string `json:"parent_root"`
	StateRoot     string `json:"state_root"`
	BodyRoot      string `json:"body_root"`
}

// UnmarshalJSON implements custom JSON unmarshaling for BlockHeader
// This handles the conversion of hex strings to byte arrays
func (bh *BlockHeader) UnmarshalJSON(data []byte) error {
	var jsonHeader blockHeaderJSON
	if err := json.Unmarshal(data, &jsonHeader); err != nil {
		return fmt.Errorf("failed to unmarshal block header JSON: %w", err)
	}

	// Convert basic fields
	bh.Slot = jsonHeader.Slot
	bh.ProposerIndex = jsonHeader.ProposerIndex

	// Convert hex strings to bytes
	var err error

	if bh.ParentRoot, err = hexToBytes(jsonHeader.ParentRoot); err != nil {
		return fmt.Errorf("failed to decode parent_root: %w", err)
	}

	if bh.StateRoot, err = hexToBytes(jsonHeader.StateRoot); err != nil {
		return fmt.Errorf("failed to decode state_root: %w", err)
	}

	if bh.BodyRoot, err = hexToBytes(jsonHeader.BodyRoot); err != nil {
		return fmt.Errorf("failed to decode body_root: %w", err)
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for BlockHeader
// This converts byte arrays back to hex strings for JSON output
func (bh BlockHeader) MarshalJSON() ([]byte, error) {
	jsonHeader := blockHeaderJSON{
		Slot:          bh.Slot,
		ProposerIndex: bh.ProposerIndex,
		ParentRoot:    bytesToHex(bh.ParentRoot),
		StateRoot:     bytesToHex(bh.StateRoot),
		BodyRoot:      bytesToHex(bh.BodyRoot),
	}

	return json.Marshal(jsonHeader)
}
