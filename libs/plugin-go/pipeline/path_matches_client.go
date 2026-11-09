package pipeline

import (
	"os"
	"path/filepath"
)

// pathMatches lists the files a program name matches in the folders of a PATH
// value, in order: "<dir>/<name>", and on Windows also the .exe, .cmd and .bat
// forms. A folder that holds none adds nothing.
func pathMatches(name, path, goos string) []string {
	suffixes := []string{""}
	if goos == "windows" {
		suffixes = []string{".exe", ".cmd", ".bat"}
	}
	var out []string
	seen := map[string]bool{}
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			continue
		}
		for _, suffix := range suffixes {
			file := filepath.Join(dir, name+suffix)
			if info, err := os.Stat(file); err == nil && !info.IsDir() && !seen[file] {
				seen[file] = true
				out = append(out, file)
			}
		}
	}
	return out
}
