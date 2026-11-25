package projectstandard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Apply writes the standard files: the copy of each whole file it replaces
// first, then the new content.
//
// Parameters: root is the workspace; diffs are the differences to apply; out
// gets one line per file.
// Returns: the workspace-relative paths written, copies included.
// Errors: a file cannot be written; the files written before stay.
func Apply(root string, diffs []Difference, out io.Writer) ([]string, error) {
	var written []string
	for _, d := range diffs {
		for _, e := range d.Edits() {
			full := filepath.Join(root, filepath.FromSlash(e.Path))
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				return written, err
			}
			if err := os.WriteFile(full, []byte(e.Content), 0o644); err != nil {
				return written, err
			}
			written = append(written, e.Path)
		}
		fmt.Fprintf(out, "  standard: %s\n", d.Summary())
	}
	return written, nil
}
