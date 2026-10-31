package derivelock

import (
	"crypto/sha256"
	"encoding/hex"
)

// Hash is the hash of the content smith owns in a file, written
// "sha256:<hex>" so a later version can change the algorithm.
func Hash(content []byte) string {
	sum := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(sum[:])
}
