package toolpin

import (
	"os"
	"path/filepath"
)

// nativePin reads the pin of a label from the first native file of its
// language that names one. It is empty when none does.
func nativePin(workspaceRoot, label string) string {
	for _, f := range nativeFiles[label] {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, f.name))
		if err != nil {
			continue
		}
		if v := f.parse(string(raw)); v != "" {
			return v
		}
	}
	return ""
}
