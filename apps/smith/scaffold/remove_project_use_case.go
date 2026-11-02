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
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
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
	// SkipMentions leaves out the scan for files that still name the project.
	SkipMentions bool
	// KeepCopy moves the folder to .smith/removed/<name>-<time>/ instead of
	// deleting it, so nothing is lost (smith remove --keep-copy).
	KeepCopy bool
	// Delete deletes the folder for good, including files git cannot bring
	// back (smith remove --delete). A folder that holds such files is refused
	// unless KeepCopy or Delete is set.
	Delete bool
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
	// Unsaved are the files in the folder that git cannot bring back: untracked
	// files and uncommitted changes.
	Unsaved []string `json:"unsaved,omitempty"`
	// Ignored counts the ignored entries of the folder (node_modules, build output).
	Ignored int `json:"ignored,omitempty"`
	// KeptAt is where the folder was moved (or would be) with KeepCopy.
	KeptAt string `json:"keptAt,omitempty"`
	// Mentions are tracked files that still name the project; smith does not change them.
	Mentions []Mention `json:"mentions,omitempty"`
	// MentionsRootOnly is true when the name is short and only the root path was searched.
	MentionsRootOnly bool `json:"mentionsRootOnly,omitempty"`
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
	unsaved := ListUnsavedFiles(s.Root, p.Root)
	r.Unsaved, r.Ignored = unsaved.Files, unsaved.Ignored
	if opts.KeepCopy {
		r.KeptAt = ".smith/removed/" + p.Name + "-" + time.Now().Format("20060102-150405")
	}
	if !opts.DryRun && len(r.Unsaved) > 0 && !opts.KeepCopy && !opts.Delete {
		return r, problem.New(fmt.Sprintf("%s holds %d file(s) that git cannot bring back (untracked or not committed)", p.Name, len(r.Unsaved)),
			"--keep-copy moves the folder to .smith/removed/ instead of deleting it", "--delete deletes it for good")
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
		dir := filepath.Join(s.Root, filepath.FromSlash(p.Root))
		if opts.KeepCopy {
			target := filepath.Join(s.Root, filepath.FromSlash(r.KeptAt))
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return r, err
			}
			if err := os.Rename(dir, target); err != nil {
				return r, fmt.Errorf("moving %s to %s: %w", p.Root, r.KeptAt, err)
			}
		} else if err := os.RemoveAll(dir); err != nil {
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
		r.findMentions(s.Root, opts)
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
	r.findMentions(s.Root, opts)
	return r, nil
}

// findMentions fills Mentions, skipping the files smith changed.
func (r *Removal) findMentions(root string, opts RemoveOptions) {
	if opts.SkipMentions {
		return
	}
	skip := append([]string(nil), r.Files...)
	if rec, err := derivelock.Read(root); err == nil {
		for _, f := range rec.Files {
			if f.Owner == derivelock.OwnerSmith {
				skip = append(skip, f.Path) // smith regenerates it: no need to review
			}
		}
	}
	r.Mentions, r.MentionsRootOnly = FindMentions(root, r.Project, r.Root, skip)
}
