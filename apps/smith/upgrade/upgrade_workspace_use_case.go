package upgrade

import (
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Result is what smith upgrade changed, or with Check would change.
type Result struct {
	// Migrations are the smith.toml schema steps applied.
	Migrations []string `json:"migrations,omitempty"`
	// Files are the files written or deleted.
	Files []string `json:"files,omitempty"`
	// Lost are team steps in slots the new pipelines no longer have.
	Lost []string `json:"lost,omitempty"`
	// Recipes are the pinned recipe URLs whose content changed: accepted by
	// the upgrade, or with check, the ones it would accept.
	Recipes []string `json:"recipes,omitempty"`
	// Hints are changes smith suggests but does not make.
	Hints []string `json:"hints,omitempty"`
}

// Changed reports whether anything changed (or would).
func (r Result) Changed() bool {
	return len(r.Migrations) > 0 || len(r.Files) > 0 || len(r.Recipes) > 0
}

// Upgrade migrates smith.toml, then regenerates the pipelines. With check it
// changes nothing and reports what it would do.
func Upgrade(dir, version string, check bool, log io.Writer) (Result, error) {
	root, err := workspace.FindRoot(dir)
	if err != nil {
		return Result{}, err
	}
	var r Result
	if r.Migrations, err = workspace.MigrateFile(root, check); err != nil {
		return r, err
	}
	if len(r.Migrations) > 0 {
		r.Files = append(r.Files, workspace.FileName)
		if check {
			// The configuration must be migrated before the rest can be read.
			return r, nil
		}
	}
	s, err := session.Open(session.Options{Dir: root, Version: version, Log: log})
	if err != nil {
		return r, err
	}
	defer s.Close()
	if r.Recipes, err = recipeChanges(&recipe.Loader{Root: s.Root}, s.Config, check); err != nil {
		return r, err
	}
	if check {
		c, err := cipipeline.Collect(s)
		if err != nil {
			return r, err
		}
		r.Files = append(r.Files, cipipeline.Drift(s.Root, s.Config, cipipeline.NewModel(c, s.Version, s.Config.Release.Branch))...)
	} else {
		changed, lost, err := cipipeline.Regenerate(s)
		if err != nil {
			return r, err
		}
		r.Files, r.Lost = append(r.Files, changed...), lost
	}
	r.Hints = hints(s)
	return r, nil
}

// hints suggest what a workspace made by an older smith may want: today, a
// default stack for Node, without which new Node projects get no lint or
// tests (smith 0.1 had them in the node plugin).
func hints(s *session.Session) []string {
	if _, ok := s.Config.DefaultStacks["node"]; ok {
		return nil
	}
	for _, p := range s.Graph.Projects {
		if p.Plugin == "node" && p.Stack == "" {
			return []string{`Node projects record no stack and smith.toml sets no default: new ones get no lint or tests. Add [default_stacks] node = "node-jest" (https://github.com/monosmith-app/monosmith/blob/main/docs/reference/stacks.md).`}
		}
	}
	return nil
}
