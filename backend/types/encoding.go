package types

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// hexToBytes converts a hex string (with or without 0x prefix) to bytes
func hexToBytes(hexStr string) ([]byte, error) {
	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// Decode hex string to bytes
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string '%s': %w", hexStr, err)
	}

	return bytes, nil
}

// bytesToHex converts bytes to a hex string with 0x prefix
func bytesToHex(data []byte) string {
	return fmt.Sprintf("0x%x", data)
}
