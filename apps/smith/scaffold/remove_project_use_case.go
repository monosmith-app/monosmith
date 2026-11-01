package scaffold

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RemoveOptions configure smith remove.
type RemoveOptions struct {
	Name string
	// DryRun says what would change, changing nothing.
	DryRun bool
	// Force removes a project other projects still import.
	Force bool
}

// Removal is what removing a project changes (or, in a dry run, would).
type Removal struct {
	Project string `json:"project"`
	Root    string `json:"root"`
	// Dependents still import the project (only with Force).
	Dependents []string `json:"dependents,omitempty"`
	// Files are the other files changed: plugin clean-up, smith.toml, pipelines.
	Files []string `json:"files,omitempty"`
	Notes []string `json:"notes,omitempty"`
}

// RemoveProject deletes a project's directory, lets its plugin clean what it
// owns outside it (go.work, the lockfile, the venv), drops its smith.toml
// tables and regenerates the pipelines. A project others import is refused
// unless Force.
func RemoveProject(s *session.Session, opts RemoveOptions) (Removal, error) {
	p := s.Graph.Get(opts.Name)
	if p == nil {
		return Removal{}, fmt.Errorf("no project named %q (projects: %s).%s", opts.Name, strings.Join(s.Graph.Names(), ", "), nearmatch.Suggestion(opts.Name, s.Graph.Names()))
	}
	r := Removal{Project: p.Name, Root: p.Root, Dependents: s.Graph.Dependents(p.Name)}
	if len(r.Dependents) > 0 && !opts.Force {
		return r, fmt.Errorf("%s is imported by %s: remove those imports first, or pass --force to remove it anyway", p.Name, strings.Join(r.Dependents, ", "))
	}
	cfgPath := filepath.Join(s.Root, workspace.FileName)
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		return r, err
	}
	cfgText, cfgChanged := workspace.RemoveProjectTables(string(raw), p.Name)
	if cfgChanged {
		if _, err := workspace.Parse([]byte(cfgText)); err != nil {
			return r, fmt.Errorf("refusing to write an invalid %s: %w", workspace.FileName, err)
		}
		r.Files = append(r.Files, workspace.FileName)
	}
	if !opts.DryRun {
		if err := os.RemoveAll(filepath.Join(s.Root, filepath.FromSlash(p.Root))); err != nil {
			return r, err
		}
	}
	var cleaned protocol.RemoveResult
	if _, err := s.Plugins.Call(p.Plugin, protocol.MethodRemove, protocol.RemoveParams{Project: p.Project, DryRun: opts.DryRun}, &cleaned, 20*time.Minute); err != nil {
		if opts.DryRun {
			return r, fmt.Errorf("plugin %s: %w", p.Plugin, err)
		}
		// The directory is gone: finish the removal, and say what is left.
		r.Notes = append(r.Notes, fmt.Sprintf("warning: the %s plugin could not clean up (%v); fix that, then run its install again", p.Plugin, err))
	}
	r.Files = append(r.Files, cleaned.Files...)
	r.Notes = append(r.Notes, cleaned.Notes...)
	if opts.DryRun {
		r.Notes = append(r.Notes, "the pipelines would be regenerated")
		sort.Strings(r.Files)
		return r, nil
	}
	if cfgChanged {
		if err := os.WriteFile(cfgPath, []byte(cfgText), 0o644); err != nil {
			return r, err
		}
		if s.Config, err = workspace.Load(s.Root); err != nil {
			return r, err
		}
	}
	if err := s.Refresh(); err != nil {
		return r, err
	}
	changed, _, err := cipipeline.Regenerate(s)
	if err != nil {
		return r, err
	}
	r.Files = append(r.Files, changed...)
	sort.Strings(r.Files)
	return r, nil
}
