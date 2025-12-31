package util

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Sum returns the SHA256 sum of the given data
func SHA256Sum(data string) []byte {
	// Convert the data to bytes
	dataBytes := []byte(data)

	// Create a new buffer for the hex-encoded sum
	hexSum := make([]byte, 64)

	// Calculate the SHA256 sum
	sumRaw := sha256.Sum256(dataBytes)

	// Encode the sum as hex into the buffer
	hex.Encode(hexSum, sumRaw[:])

	// Return the hex encoded sum
	return hexSum
}
