package scaffold

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/recipechoice"
	"github.com/monosmith-app/monosmith/apps/smith/recipestack"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RecipeUse is how `smith add` answers the workspace's recipe.
type RecipeUse struct {
	// Answers are the --choice answers given.
	Answers map[string]string
	// In is where the rest are asked from when Interactive says a person is
	// there; without one a choice not given takes its default.
	In          io.Reader
	Interactive bool
}

// recipePlan is a recipe resolved for one new project, before anything is
// written.
type recipePlan struct {
	ref      string
	res      recipe.Resolution
	stack    protocol.Stack
	reuse    bool
	settings map[string]map[string]any
}

// planRecipe resolves the recipe the workspace names for the project's kind:
// it asks the choices not answered with --choice, checks that the plugins it
// picks work together, and shows what it will use. It writes nothing. It
// returns nil when no recipe applies: the workspace names none, or --stack
// bypasses it (then --choice has nothing to answer, which is an error).
func planRecipe(s *session.Session, opts AddOptions) (*recipePlan, error) {
	use := opts.Recipe
	if use == nil {
		use = &RecipeUse{}
	}
	ref := workspace.RecipeFor(s.Config, opts.Kind)
	switch {
	case opts.Stack != "" && len(use.Answers) > 0:
		return nil, fmt.Errorf("--choice answers a recipe's choices, and --stack bypasses the recipe: use one or the other")
	case opts.Stack != "":
		return nil, nil
	case ref == "" && len(use.Answers) > 0:
		return nil, fmt.Errorf("--choice answers a recipe's choices, but smith.toml names no recipe: set [workspace] recipe")
	case ref == "":
		return nil, nil
	}
	loaded, err := (&recipe.Loader{Root: s.Root}).Load(recipeLocation(s.Root, ref))
	if err != nil {
		return nil, fmt.Errorf("the recipe %s: %w", ref, err)
	}
	in := use.In
	if in == nil {
		in = strings.NewReader("")
	}
	res, err := recipechoice.Asker{In: in, Out: opts.Out, Interactive: use.Interactive}.Resolve(loaded, use.Answers)
	if err != nil {
		return nil, fmt.Errorf("the recipe %s: %w", ref, err)
	}
	catalog := recipe.Catalog{Root: s.Root, Version: s.Version, External: s.Config.Plugins.External}
	if err := recipe.Validate(res, catalog.Lookup); err != nil {
		return nil, fmt.Errorf("the recipe %s: %w", ref, err)
	}
	st, err := recipestack.StackOf(res)
	if err != nil {
		return nil, err
	}
	name, reuse, err := recipestack.Choose(recipestack.NameOf(res), st, func(n string) (protocol.Stack, bool) {
		for _, known := range s.Stacks.All() {
			if known.Name == n {
				return known.Stack, true
			}
		}
		return protocol.Stack{}, false
	})
	if err != nil {
		return nil, err
	}
	st.Name = name
	if err := pluginhost.ValidateStack(st, s.Plugins); err != nil {
		return nil, fmt.Errorf("the recipe %s resolves to a stack this workspace cannot use: %w", ref, err)
	}
	if s.Plugins.Get(st.Language) == nil {
		return nil, fmt.Errorf("the recipe %s uses the language plugin %s, which is not enabled in this workspace", ref, st.Language)
	}
	if opts.Lang != "" {
		asked, err := s.Plugins.ForLanguage(opts.Lang)
		if err != nil {
			return nil, err
		}
		if asked != s.Plugins.Get(st.Language) {
			return nil, fmt.Errorf("--lang %s, but the recipe uses %s for these answers: answer its choices (--choice) instead, or bypass it with --stack", opts.Lang, st.Language)
		}
	}
	fmt.Fprint(opts.Out, recipechoice.Describe(res, s.Root))
	settings, unused := settingsOfProject(res, st)
	if len(unused) > 0 {
		fmt.Fprintf(opts.Out, "  not used here: the recipe's settings for %s (the project does not use that plugin)\n", strings.Join(unused, ", "))
	}
	return &recipePlan{ref: ref, res: res, stack: st, reuse: reuse, settings: settings}, nil
}

// recipeLocation is where a recipe named in smith.toml is: a URL as it is, a
// path relative to smith.toml made absolute.
func recipeLocation(root, ref string) string {
	if strings.Contains(ref, "://") {
		return ref
	}
	return filepath.Join(root, filepath.FromSlash(ref))
}

// settingsOfProject keeps the recipe's settings for the plugins the project
// is built with, and lists the plugins whose settings it leaves out.
func settingsOfProject(res recipe.Resolution, st protocol.Stack) (kept map[string]map[string]any, unused []string) {
	used := map[string]bool{st.Language: true}
	for _, name := range append(st.Plugins(), st.Scaffold) {
		used[name] = true
	}
	for plugin, table := range res.Settings {
		values, ok := table.(map[string]any)
		if !ok {
			continue
		}
		if !used[plugin] {
			unused = append(unused, plugin)
			continue
		}
		if kept == nil {
			kept = map[string]map[string]any{}
		}
		kept[plugin] = values
	}
	sort.Strings(unused)
	return kept, unused
}

// declare writes the plan to smith.toml and reloads the session from it. It
// returns the function that takes it back out when creating the project fails.
func (p *recipePlan) declare(s *session.Session, project string) (func(), error) {
	undo, err := recipestack.Declare(s.Root, recipestack.Declaration{
		Project: project, Stack: p.stack, Reuse: p.reuse, Settings: p.settings,
		Recipe: workspace.ProjectRecipe{Name: p.res.Name, Version: p.res.Version, Source: p.ref, Answers: recipestack.Answers(p.res)},
	})
	if err != nil {
		return nil, err
	}
	rollback := func() {
		if err := undo(); err == nil {
			_ = s.Reload()
		}
	}
	if err := s.Reload(); err != nil {
		rollback()
		return nil, fmt.Errorf("smith.toml with the recipe's stack does not load: %w", err)
	}
	return rollback, nil
}
