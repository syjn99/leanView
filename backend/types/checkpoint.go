package types

// Checkpoint represents a justified or finalized checkpoint in Lean consensus
type Checkpoint struct {
	Root []byte `json:"root"`
	Slot uint64 `json:"slot"`
}
