package kindchange

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Outcome is what applying a plan did.
type Outcome struct {
	// Moved says how the folder moved: "git mv", or "rename" where git could
	// not move it (a folder with nothing tracked in it).
	Moved string `json:"moved,omitempty"`
	// Regenerated are the pipeline files that changed.
	Regenerated []string `json:"regenerated,omitempty"`
	// Notes are things to tell the person after the fact.
	Notes []string `json:"notes,omitempty"`
}

// Apply writes the plan's edits, moves the folder, and regenerates the
// pipelines, which name projects by kind.
func Apply(s *session.Session, plan Plan) (Outcome, error) {
	var out Outcome
	if err := doctorfix.ApplyFiles(s.Root, protocol.Fix{Files: plan.Edits}); err != nil {
		return out, err
	}
	if plan.Moves() {
		how, err := move(s, plan.Root, plan.NewRoot)
		if err != nil {
			return out, fmt.Errorf("the files are edited but %s could not move to %s: %w", plan.Root, plan.NewRoot, err)
		}
		out.Moved = how
	}
	if err := s.Reload(); err != nil {
		return out, err
	}
	p := s.Graph.Get(plan.Project)
	switch {
	case p == nil:
		out.Notes = append(out.Notes, fmt.Sprintf("smith no longer finds %s: run smith doctor", plan.Project))
	case p.Kind != plan.To:
		out.Notes = append(out.Notes, fmt.Sprintf("%s is still %s to smith: run smith doctor", plan.Project, article(p.Kind)))
	}
	changed, _, err := cipipeline.Regenerate(s)
	if err != nil {
		return out, err
	}
	out.Regenerated = changed
	sort.Strings(out.Regenerated)
	return out, nil
}

// move puts a folder in its new place with git, so the history follows; a
// folder git has nothing of is renamed.
func move(s *session.Session, from, to string) (string, error) {
	if err := os.MkdirAll(filepath.Join(s.Root, filepath.FromSlash(filepath.Dir(to))), 0o755); err != nil {
		return "", err
	}
	if _, err := s.Git.Run("mv", from, to); err == nil {
		return "git mv", nil
	}
	if err := os.Rename(filepath.Join(s.Root, filepath.FromSlash(from)), filepath.Join(s.Root, filepath.FromSlash(to))); err != nil {
		return "", err
	}
	return "rename", nil
}
