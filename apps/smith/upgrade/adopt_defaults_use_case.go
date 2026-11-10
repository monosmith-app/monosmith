package upgrade

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// DefaultsChange is the move of a workspace from one defaults vintage to the
// latest one.
type DefaultsChange struct {
	// From is the vintage smith.toml records; empty when it records none.
	From string `json:"from"`
	// To is the latest vintage this smith knows.
	To string `json:"to"`
}

// Changed reports whether the workspace is not on the latest vintage.
func (c DefaultsChange) Changed() bool { return c.From != c.To }

// AdoptDefaults moves the workspace to the latest defaults vintage by
// recording it in smith.toml (ADR 0054). Only that one value changes; comments
// and layout stay. Today there is one vintage, so the move changes no default:
// it records which vintage the workspace follows.
//
// Parameters: root is the workspace root; check makes it report the change and
// write nothing.
// Returns: the change (From equals To when nothing is to do). It fails when
// smith.toml cannot be read or written, or records a vintage newer than this
// smith knows: upgrade smith first.
func AdoptDefaults(root string, check bool) (DefaultsChange, error) {
	path := filepath.Join(root, workspace.FileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return DefaultsChange{}, err
	}
	cfg, err := workspace.Parse(raw)
	if err != nil {
		return DefaultsChange{}, err
	}
	change := DefaultsChange{From: cfg.Workspace.Defaults, To: workspace.LatestDefaultsVintage()}
	if workspace.DefaultsVintageState(cfg.Workspace.Defaults) == workspace.VintageAhead {
		return change, fmt.Errorf("%s records defaults %s, newer than this smith knows (%s): upgrade smith", workspace.FileName, change.From, change.To)
	}
	if !change.Changed() || check {
		return change, nil
	}
	text := configedit.SetKey(string(raw), "workspace", "defaults", configedit.Quote(change.To))
	return change, os.WriteFile(path, []byte(text), 0o644)
}
