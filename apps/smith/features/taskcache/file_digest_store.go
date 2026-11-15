package taskcache

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// FileDigests memoises file content digests for one smith invocation, so a file
// read by many tasks is hashed once.
type FileDigests struct {
	root string
	mu   sync.Mutex
	memo map[string]string
}

// NewFileDigests hashes files under a workspace root.
func NewFileDigests(root string) *FileDigests {
	return &FileDigests{root: root, memo: map[string]string{}}
}

// Digest returns the SHA-256 of a workspace-relative file. A file that has
// disappeared digests to "missing", which still changes the key.
func (d *FileDigests) Digest(rel string) string {
	d.mu.Lock()
	if v, ok := d.memo[rel]; ok {
		d.mu.Unlock()
		return v
	}
	d.mu.Unlock()
	v := "missing"
	if f, err := os.Open(filepath.Join(d.root, filepath.FromSlash(rel))); err == nil {
		h := sha256.New()
		if _, err := io.Copy(h, f); err == nil {
			v = hex.EncodeToString(h.Sum(nil))
		}
		_ = f.Close()
	}
	d.mu.Lock()
	d.memo[rel] = v
	d.mu.Unlock()
	return v
}

// Forget drops memoised digests, after a task may have changed files.
func (d *FileDigests) Forget() {
	d.mu.Lock()
	d.memo = map[string]string{}
	d.mu.Unlock()
}
