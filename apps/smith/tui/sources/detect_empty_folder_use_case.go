package sources

import "os"

// FolderIsEmpty is true when dir holds nothing but, at most, an empty git
// repository's .git: there is nothing yet to adopt or convert.
func FolderIsEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.Name() != ".git" {
			return false
		}
	}
	return true
}
