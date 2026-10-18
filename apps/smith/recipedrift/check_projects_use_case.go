package recipedrift

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/recipestack"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Check compares every project made from a recipe with the recipe as it is
// now, in the graph's order. A project that records no recipe is not checked.
func Check(s *session.Session) []Report {
	loader := &recipe.Loader{Root: s.Root}
	var out []Report
	for _, p := range s.Graph.Projects {
		ov, ok := s.Config.Projects[p.Name]
		if !ok || ov.Recipe == nil {
			continue
		}
		out = append(out, CheckProject(s, loader, p, *ov.Recipe))
	}
	return out
}

// CheckProject resolves the workspace's current recipe for a project with
// the answers it recorded, and compares the result with what the project is
// built with: its stack's plugins, and the settings smith.toml holds for it.
// The recipe is the one smith.toml names for the project's kind, or the one
// the project recorded when smith.toml names none; if that is a different
// recipe from the one the project was made with, the project is not checked.
func CheckProject(s *session.Session, loader *recipe.Loader, p *projectgraph.Project, rec workspace.ProjectRecipe) Report {
	r := Report{Project: p.Name, Recipe: rec.Name, From: rec.Version}
	loaded, ref, problem := load(s, loader, p, rec)
	if problem != "" {
		r.Problem = problem
		return r
	}
	r.Ref = ref
	res, err := recipe.Resolve(loaded, rec.Answers)
	var missing *recipe.MissingAnswerError
	switch {
	case errors.As(err, &missing):
		names := make([]string, 0, len(missing.Choices))
		for _, q := range missing.Choices {
			names = append(names, fmt.Sprintf("%q", q.Name))
		}
		r.Problem = fmt.Sprintf("the recipe now asks %s, which this project has no answer for: add the answer under [projects.%s.recipe.answers] in smith.toml", strings.Join(names, ", "), p.Name)
		return r
	case err != nil:
		r.Problem = "the answers the project recorded no longer fit the recipe: " + err.Error()
		return r
	}
	catalog := recipe.Catalog{Root: s.Root, Version: s.Version, External: s.Config.Plugins.External}
	lookup := func(name string) (protocol.InitializeResult, error) {
		if c := s.Plugins.Get(name); c != nil {
			return c.Info, nil
		}
		return catalog.Lookup(name)
	}
	if err := recipe.Validate(res, lookup); err != nil {
		r.Problem = "the recipe now resolves to plugins that cannot work together: " + err.Error()
		return r
	}
	now, err := recipestack.StackOf(res)
	if err != nil {
		r.Problem = err.Error()
		return r
	}
	if p.Stack == "" {
		r.Problem = "the project records no stack, so there is nothing to compare the recipe with"
		return r
	}
	old, err := s.Stacks.Get(p.Stack)
	if err != nil {
		r.Problem = err.Error()
		return r
	}
	now.Name = p.Stack
	settings, _ := recipestack.ProjectSettings(res, now)
	answers := recipestack.Answers(res)

	r.To, r.Ref, r.Stack, r.Settings, r.Answers, r.Resolution = res.Version, ref, now, settings, answers, res
	stack, added := stackChanges(old, now, res)
	r.Added = added
	r.Changes = append(r.Changes, stack...)
	r.Changes = append(r.Changes, settingsChanges(s.Config.Projects[p.Name].Settings, settings, res, old, now)...)
	r.Changes = append(r.Changes, answerChanges(rec.Answers, answers)...)
	return r
}

// load reads the recipe a project is compared with. smith.toml's recipe for
// the project's kind comes first, then the source the project recorded; the
// first that loads and is the recipe the project was made with is used.
func load(s *session.Session, loader *recipe.Loader, p *projectgraph.Project, rec workspace.ProjectRecipe) (loaded recipe.Loaded, ref, problem string) {
	var refs []string
	for _, candidate := range []string{workspace.RecipeFor(s.Config, string(p.Kind)), rec.Source} {
		if candidate != "" && !slices.Contains(refs, candidate) {
			refs = append(refs, candidate)
		}
	}
	if len(refs) == 0 {
		return recipe.Loaded{}, "", "smith.toml names no recipe and the project recorded no source"
	}
	var failed, others []string
	for _, candidate := range refs {
		l, err := loader.Load(workspace.RecipeLocation(s.Root, candidate))
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", candidate, err))
			continue
		}
		if name := l.Chain[len(l.Chain)-1].Recipe.Name; name != rec.Name {
			others = append(others, fmt.Sprintf("%s is the recipe %q", candidate, name))
			continue
		}
		return l, candidate, ""
	}
	if len(failed) > 0 {
		return recipe.Loaded{}, "", "the recipe could not be loaded (" + strings.Join(failed, "; ") + ")"
	}
	return recipe.Loaded{}, "", "smith.toml now names a different recipe (" + strings.Join(others, "; ") + "), so there is no way to say how this project differs from " + rec.Name
}
