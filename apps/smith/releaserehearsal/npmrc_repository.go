package releaserehearsal

import (
	"os"
	"strings"
)

// readNpmrc joins the .npmrc files npm reads, lowest priority first, so the
// last line for a key is the one npm uses.
func readNpmrc(paths ...string) string {
	var out []string
	for _, p := range paths {
		if raw, err := os.ReadFile(p); err == nil {
			out = append(out, string(raw))
		}
	}
	return strings.Join(out, "\n")
}
