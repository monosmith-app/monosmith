package cachekey

import (
	"encoding/hex"
	"fmt"
	"io"
)

// KeyBytes is the number of random bytes in a key; the key is twice as many
// hex characters.
const KeyBytes = 32

// Generate makes a key from a source of random bytes.
//
// Parameters: random is the source; production passes crypto/rand.Reader.
// Returns: 64 lowercase hex characters.
// Errors: the source could not give 32 bytes.
func Generate(random io.Reader) (string, error) {
	buf := make([]byte, KeyBytes)
	if _, err := io.ReadFull(random, buf); err != nil {
		return "", fmt.Errorf("the system random source failed: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
