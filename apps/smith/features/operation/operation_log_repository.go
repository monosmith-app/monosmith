package operation

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// logPath is where the log lives: .smith is git-ignored, so the log belongs to
// one clone, as the commits it names may differ between clones.
func logPath(root string) string {
	return filepath.Join(root, ".smith", "operations.json")
}

// ReadLog loads the operation log of a workspace. A workspace with no log has
// an empty one. It fails when the file is not valid or comes from a newer smith.
//
// Parameters: root is the workspace root.
func ReadLog(root string) (Log, error) {
	raw, err := os.ReadFile(logPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return Log{Schema: SchemaVersion}, nil
	}
	if err != nil {
		return Log{}, err
	}
	var l Log
	if err := json.Unmarshal(raw, &l); err != nil {
		return Log{}, fmt.Errorf(".smith/operations.json is not valid JSON: %w", err)
	}
	if l.Schema > SchemaVersion {
		return Log{}, fmt.Errorf(".smith/operations.json has schema %d, this smith reads up to %d: upgrade smith", l.Schema, SchemaVersion)
	}
	l.Schema = SchemaVersion
	return l, nil
}

// WriteLog saves the operation log, creating .smith when needed.
//
// Parameters: root is the workspace root; l is the log to save.
func WriteLog(root string, l Log) error {
	l.Schema = SchemaVersion
	raw, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(logPath(root)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(logPath(root), append(raw, '\n'), 0o644)
}
