package lineendings

import (
	"os"
)

// WriteFile writes content to path and keeps the line ending the file has now:
// a CRLF file stays CRLF and an LF file stays LF, whatever ending content
// uses. A file that does not exist, or has no line break to learn from, gets
// content as it is.
//
// Parameters: path is the file; content is the new text; perm is the mode of a
// new file.
// Returns: the error of the write.
func WriteFile(path string, content []byte, perm os.FileMode) error {
	if existing, err := os.ReadFile(path); err == nil {
		content = []byte(Match(string(existing), string(content)))
	}
	return os.WriteFile(path, content, perm)
}
