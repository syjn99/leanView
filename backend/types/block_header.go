package types

// BlockHeader represents a Lean Block header
type BlockHeader struct {
	Slot          uint64 `db:"slot" json:"slot"`
	ProposerIndex uint64 `db:"proposer_index" json:"proposer_index"`
	ParentRoot    []byte `db:"parent_root" json:"parent_root"`
	StateRoot     []byte `db:"state_root" json:"state_root"`
	BodyRoot      []byte `db:"body_root" json:"body_root"`
}
