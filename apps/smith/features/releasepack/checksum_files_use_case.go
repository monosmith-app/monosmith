package releasepack

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// Checksum reads each package file, and writes its checksum file beside it
// (core-0.5.2.tgz.sha256, one line "<hex>  <name>" that sha256sum -c accepts).
//
// Parameters: root is the workspace; files are the workspace-relative package files.
// Returns: the files with their digests, and the workspace-relative checksum
// files, in the same order.
// Errors: a file cannot be read or a checksum file cannot be written.
func Checksum(root string, files []string) ([]File, []string, error) {
	var out []File
	var sidecars []string
	for _, rel := range files {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return nil, nil, fmt.Errorf("reading %s: %w", rel, err)
		}
		sum := sha256.Sum256(raw)
		f := File{Name: path.Base(rel), SHA256: hex.EncodeToString(sum[:]), Size: int64(len(raw))}
		sidecar := rel + ChecksumSuffix
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(sidecar)), []byte(f.SHA256+"  "+f.Name+"\n"), 0o644); err != nil {
			return nil, nil, err
		}
		out = append(out, f)
		sidecars = append(sidecars, sidecar)
	}
	return out, sidecars, nil
}
