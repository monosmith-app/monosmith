package recipedrift

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Change is one project brought up to date with its recipe.
type Change struct {
	Project string   `json:"project"`
	Recipe  string   `json:"recipe"`
	From    string   `json:"from,omitempty"`
	To      string   `json:"to,omitempty"`
	Changes []string `json:"changes,omitempty"`
}

// Plan is what `smith upgrade` does about recipes: the projects it brings up
// to date, the ones it leaves and why, and smith.toml as it will be.
type Plan struct {
	Changes []Change
	// Skipped say why a project that drifted, or could not be checked, is
	// left as it is.
	Skipped []string
	// Starters say which plugins a recipe now adds whose starter files an
	// upgrade does not write.
	Starters []string
	// Before and After are smith.toml's text; they are equal when nothing
	// changes.
	Before, After string
}

// Changed reports whether the upgrade changes smith.toml.
func (p Plan) Changed() bool { return len(p.Changes) > 0 && p.Before != p.After }

// Projects lists the projects the plan changes, in order.
func (p Plan) Projects() []string {
	names := make([]string, 0, len(p.Changes))
	for _, c := range p.Changes {
		names = append(names, c.Project)
	}
	return names
}

// Apply writes smith.toml and reloads the session from it, putting the old
// text back if the new one does not load.
func (p Plan) Apply(s *session.Session) error {
	if !p.Changed() {
		return nil
	}
	path := filepath.Join(s.Root, workspace.FileName)
	if err := os.WriteFile(path, []byte(p.After), 0o644); err != nil {
		return err
	}
	if err := s.Reload(); err != nil {
		_ = os.WriteFile(path, []byte(p.Before), 0o644)
		_ = s.Reload()
		return err
	}
	return nil
}
