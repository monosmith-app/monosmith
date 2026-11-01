package derive

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Proposal is a file a plugin derives, with the plugin's name.
type Proposal struct {
	Plugin string
	File   protocol.DerivedFile
}

// BuildPlan asks every plugin that derives for its files, reads the files from
// disk and compares them with the record. It writes nothing. It returns an
// error when the record cannot be read, a plugin answers with a path outside
// the workspace, or two renderers claim the same file in the whole-file mode
// with different content.
func BuildPlan(s *session.Session) (Plan, derivelock.Record, error) {
	rec, err := derivelock.Read(s.Root)
	if err != nil {
		return Plan{}, rec, err
	}
	proposals, diverging, err := collect(s, rec.OwnedBySmith())
	if err != nil {
		return Plan{}, rec, err
	}
	plan, err := PlanFor(s.Root, rec, proposals)
	plan.Diverging = diverging
	return plan, rec, err
}

// PlanFor classifies proposals against the files under root and the record.
func PlanFor(root string, rec derivelock.Record, proposals []Proposal) (Plan, error) {
	var plan Plan
	proposals = preferRecorded(rec, proposals)
	whole := map[string]protocol.DerivedFile{}
	owner := map[string]string{}
	for _, p := range proposals {
		f := p.File
		if err := checkPath(f.Path); err != nil {
			return Plan{}, fmt.Errorf("plugin %s: %w", p.Plugin, err)
		}
		if f.Mode == protocol.DeriveFile {
			if other, dup := whole[f.Path]; dup {
				if other.Content == f.Content {
					// Two renderers agree on the file: the first one keeps it.
					continue
				}
				return Plan{}, fmt.Errorf("plugins %s and %s both own the whole file %s", owner[f.Path], p.Plugin, f.Path)
			}
			whole[f.Path], owner[f.Path] = f, p.Plugin
		}
		disk, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return Plan{}, readErr
		}
		item := Classify(f, rec.Find(f.Path, f.Renderer), disk, readErr == nil)
		item.Plugin = p.Plugin
		plan.Items = append(plan.Items, item)
	}
	slices.SortFunc(plan.Items, func(a, b Item) int {
		if c := strings.Compare(a.File.Path, b.File.Path); c != 0 {
			return c
		}
		return strings.Compare(a.File.Renderer, b.File.Renderer)
	})
	return plan, nil
}

// preferRecorded settles a dispute over a whole file. When several renderers
// propose the same file with different content and the record lists exactly one
// of them for it, the others give way: the record says who has written the file.
// Without such a record the dispute stays, and PlanFor reports it.
func preferRecorded(rec derivelock.Record, proposals []Proposal) []Proposal {
	byPath := map[string][]int{}
	for i, p := range proposals {
		if p.File.Mode == protocol.DeriveFile {
			byPath[p.File.Path] = append(byPath[p.File.Path], i)
		}
	}
	drop := map[int]bool{}
	for path, idx := range byPath {
		var recorded []int
		for _, i := range idx {
			if rec.Find(path, proposals[i].File.Renderer) != nil {
				recorded = append(recorded, i)
			}
		}
		if len(idx) < 2 || len(recorded) != 1 {
			continue
		}
		for _, i := range idx {
			if i != recorded[0] && proposals[i].File.Content != proposals[recorded[0]].File.Content {
				drop[i] = true
			}
		}
	}
	if len(drop) == 0 {
		return proposals
	}
	var out []Proposal
	for i, p := range proposals {
		if !drop[i] {
			out = append(out, p)
		}
	}
	return out
}

// collect calls the derive method of each plugin that advertises it.
func collect(s *session.Session, owned []string) ([]Proposal, []Divergence, error) {
	var out []Proposal
	var diverging []Divergence
	for _, c := range s.Plugins.Clients {
		if !c.Info.Capabilities.Derive {
			continue
		}
		params := protocol.DeriveParams{Graph: map[string][]string{}, Owned: owned}
		for _, p := range s.Graph.Projects {
			if p.ServedBy(c.Info.Plugin.Name) {
				params.Projects = append(params.Projects, p.Project)
				params.Graph[p.Name] = p.Deps
			}
		}
		if len(params.Projects) == 0 {
			continue
		}
		var res protocol.DeriveResult
		if err := c.Call(protocol.MethodDerive, params, &res, time.Minute); err != nil {
			return nil, nil, problem.New("plugin "+c.Info.Plugin.Name+" failed to derive: "+err.Error(), "report this to the plugin's maintainers")
		}
		for _, f := range res.Files {
			out = append(out, Proposal{Plugin: c.Info.Plugin.Name, File: f})
		}
		for _, d := range res.Diverging {
			diverging = append(diverging, Divergence{Plugin: c.Info.Plugin.Name, DivergentFile: d})
		}
	}
	slices.SortFunc(diverging, func(a, b Divergence) int { return strings.Compare(a.Path, b.Path) })
	return out, diverging, nil
}

// checkPath rejects a path that leaves the workspace.
func checkPath(p string) error {
	clean := path.Clean(p)
	if p == "" || clean != p || path.IsAbs(p) || strings.HasPrefix(p, "../") || strings.Contains(p, `\`) || strings.Contains(p, ":") {
		return fmt.Errorf("derived path %q is not a clean workspace-relative path with forward slashes", p)
	}
	return nil
}
