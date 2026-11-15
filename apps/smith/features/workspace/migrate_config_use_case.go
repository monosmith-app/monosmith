package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// Migration rewrites smith.toml text from schema From to From+1. It edits the
// text, so comments and layout survive.
type Migration struct {
	From int
	// Change says what it does, for smith upgrade's report.
	Change string
	Apply  func(text string) (string, error)
}

// Migrations are applied in order. Schema 1 is the first; a change to
// smith.toml that old files cannot be read with adds one here and raises
// SchemaVersion.
var Migrations []Migration

var schemaLine = regexp.MustCompile(`(?m)^schema\s*=\s*(\d+)[ \t]*(#.*)?$`)

// MigrateText brings smith.toml text to SchemaVersion and lists what changed.
// The same text always gives the same result.
func MigrateText(text string) (string, []string, error) {
	return migrateTo(text, SchemaVersion, Migrations)
}

// MigrateFile migrates the workspace's smith.toml in place, unless dryRun.
func MigrateFile(root string, dryRun bool) ([]string, error) {
	path := filepath.Join(root, FileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text, changes, err := MigrateText(string(raw))
	if err != nil || len(changes) == 0 || dryRun {
		return changes, err
	}
	return changes, os.WriteFile(path, []byte(text), 0o644)
}

func migrateTo(text string, target int, migrations []Migration) (string, []string, error) {
	version := 1
	if m := schemaLine.FindStringSubmatch(text); m != nil {
		version, _ = strconv.Atoi(m[1])
	}
	if version > target {
		return text, nil, fmt.Errorf("%s uses schema %d; this smith reads up to %d, upgrade smith", FileName, version, target)
	}
	var changes []string
	for version < target {
		var step *Migration
		for i := range migrations {
			if migrations[i].From == version {
				step = &migrations[i]
			}
		}
		if step == nil {
			return text, changes, fmt.Errorf("no migration from %s schema %d", FileName, version)
		}
		next, err := step.Apply(text)
		if err != nil {
			return text, changes, fmt.Errorf("schema %d → %d: %w", version, version+1, err)
		}
		text = next
		version++
		changes = append(changes, fmt.Sprintf("schema %d → %d: %s", version-1, version, step.Change))
	}
	if len(changes) == 0 {
		return text, nil, nil
	}
	line := "schema = " + strconv.Itoa(version)
	if schemaLine.MatchString(text) {
		text = schemaLine.ReplaceAllString(text, line)
	} else {
		text = line + "\n" + text
	}
	return text, changes, nil
}
