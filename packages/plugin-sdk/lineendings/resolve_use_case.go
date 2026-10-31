package lineendings

import (
	"os"
	"path/filepath"
)

// Resolve returns content with the line ending the file at rel (slash-separated,
// from the workspace root) should have: the ending of the file that is there,
// else the one NewFileCRLF decides for a new file. It reads the disk and
// writes nothing.
func Resolve(root, rel, content string) string {
	if raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
		if crlf, known := IsCRLF(string(raw)); known {
			return Convert(content, crlf)
		}
	}
	return Convert(content, NewFileCRLF(root, rel))
}
