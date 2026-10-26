package tooldefaults

import (
	"bytes"
	"os"
	"path/filepath"
)

// Dir is where smith materialises the defaults it passes to tools, relative to
// the workspace root. It sits under .smith/, which git ignores: the file is
// derived from the smith binary, never committed, and rewritten when it differs.
const Dir = ".smith/defaults"

// WriteDefault makes sure Dir/name holds content and returns its
// workspace-relative, forward-slashed path. It rewrites the file only when its
// content differs, so timestamps stay stable.
func WriteDefault(workspaceRoot, name, content string) (string, error) {
	rel := Dir + "/" + name
	full := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
	if have, err := os.ReadFile(full); err == nil && bytes.Equal(have, []byte(content)) {
		return rel, nil
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		return "", err
	}
	return rel, nil
}
