package scaffold

import (
	"errors"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// AddOptions configure smith add.
type AddOptions struct {
	Kind string
	Name string
	// Lang selects the language plugin; empty picks the stack's, or the
	// workspace's only language.
	Lang string
	// Stack is the stack the project is built with; empty means the
	// workspace default for its language, if any.
	Stack string
	// Recipe answers the workspace's recipe ([workspace] recipe), which
	// creates the project unless Stack names one. Nil: no answers, no one to
	// ask.
	Recipe *RecipeUse
	Out    io.Writer
}

// AddProject generates a project: its stack's scaffold (or the language
// plugin) writes it and records the stack, then each other plugin of the
// stack and each configured convention that applies adds its part (lint,
// tests), and the pipelines are regenerated so CI knows about the new
// toolchain at once.
//
// With a recipe in smith.toml the stack comes from the recipe: its choices
// are answered, the plugins it picks become a stack declared in smith.toml
// with the recipe's settings, and the project is made exactly as a stack's
// project is. If making it fails, smith.toml is put back.
func AddProject(s *session.Session, opts AddOptions) (result *protocol.GenerateResult, err error) {
	kind := protocol.ProjectKind(opts.Kind)
	switch kind {
	case protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage:
	default:
		return nil, fmt.Errorf("kind %q is not app, lib or package", opts.Kind)
	}
	if err := ValidateName(opts.Kind, opts.Name); err != nil {
		return nil, err
	}
	if s.Graph.Get(opts.Name) != nil {
		return nil, fmt.Errorf("a project named %q already exists (%s)", opts.Name, s.Graph.Get(opts.Name).Root)
	}
	dir := projectkind.Directory(kind) + "/" + opts.Name
	if _, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(dir))); err == nil {
		return nil, fmt.Errorf("%s already exists", dir)
	}
	plan, err := planRecipe(s, opts)
	if err != nil {
		return nil, err
	}
	var language *host.Client
	var st protocol.Stack
	var settings map[string]map[string]any
	if plan != nil {
		language, st, settings = s.Plugins.Get(plan.stack.Language), plan.stack, plan.settings
		var undo func()
		if undo, err = plan.declare(s, opts.Name); err != nil {
			return nil, err
		}
		// A project that could not be made leaves nothing in smith.toml.
		defer func() {
			if err != nil {
				undo()
			}
		}()
	} else if language, st, err = chooseStack(s, opts); err != nil {
		return nil, err
	}
	generator := language
	if st.Scaffold != "" {
		generator = s.Plugins.Get(st.Scaffold)
	}
	if !generator.Supports(protocol.MethodGenerate) {
		return nil, fmt.Errorf("plugin %s cannot generate projects", generator.Info.Plugin.Name)
	}
	how := language.Info.Plugin.Name
	if st.Name != "" && st.Name != protocol.NoStack {
		how = "stack " + st.Name
	}
	fmt.Fprintf(opts.Out, "Adding %s %s (%s) in %s\n", opts.Kind, opts.Name, how, dir)
	var res protocol.GenerateResult
	// A workspace without a module prefix (a converted one) takes the one its
	// git remote gives, so the Go module path is the repository's own.
	modulePrefix := s.Config.Workspace.ModulePrefix
	if modulePrefix == "" {
		modulePrefix = remoteModulePrefix(gitclient.New(s.Root))
	}
	params := protocol.GenerateParams{Kind: kind, Name: opts.Name, Directory: dir, Stack: st.Name, Workspace: protocol.WorkspaceInfo{
		Name: s.Config.Workspace.Name, Scope: s.Config.Workspace.Scope, ModulePrefix: modulePrefix,
	}}
	params.Settings = settings[generator.Info.Plugin.Name]
	if err := generator.Call(protocol.MethodGenerate, params, &res, 20*time.Minute); err != nil {
		return nil, err
	}
	pending := res.InstallPending
	if err := s.Refresh(); err != nil {
		return &res, err
	}
	project := s.Graph.Get(opts.Name)
	if project == nil {
		return &res, errors.New("the plugin generated files, but no project named " + opts.Name + " was discovered")
	}
	if project.Stack != st.Name {
		return &res, fmt.Errorf("%s generated %s, but its manifest does not record stack %s", generator.Info.Plugin.Name, opts.Name, st.Name)
	}
	if err := addParts(s, project, st, params, settings, &res, &pending, opts.Out); err != nil {
		return &res, err
	}
	if err := installPending(language, pending, &res, opts.Out); err != nil {
		return &res, err
	}
	changed, _, err := cipipeline.Regenerate(s)
	if err != nil {
		return &res, err
	}
	res.Files = append(res.Files, changed...)
	sort.Strings(res.Files)
	res.Files = slices.Compact(res.Files)
	return &res, nil
}

// chooseStack resolves the language plugin and the stack of a new project.
func chooseStack(s *session.Session, opts AddOptions) (*host.Client, protocol.Stack, error) {
	if opts.Stack == protocol.NoStack {
		lang, err := chooseLanguage(s, opts)
		if err != nil {
			return nil, protocol.Stack{}, err
		}
		language, err := s.Plugins.ForLanguage(lang)
		return language, protocol.Stack{Name: protocol.NoStack}, err
	}
	if opts.Stack != "" {
		st, err := s.Stacks.Get(opts.Stack)
		if err != nil {
			return nil, st, err
		}
		language := s.Plugins.Get(st.Language)
		if opts.Lang != "" {
			asked, err := s.Plugins.ForLanguage(opts.Lang)
			if err != nil {
				return nil, st, err
			}
			if asked != language {
				return nil, st, fmt.Errorf("stack %s is for %s projects, not %s", st.Name, st.Language, opts.Lang)
			}
		}
		return language, st, nil
	}
	lang, err := chooseLanguage(s, opts)
	if err != nil {
		return nil, protocol.Stack{}, err
	}
	language, err := s.Plugins.ForLanguage(lang)
	if err != nil {
		return nil, protocol.Stack{}, err
	}
	name := s.Config.DefaultStacks[language.Info.Plugin.Name]
	if name == "" {
		return language, protocol.Stack{}, nil
	}
	st, err := s.Stacks.Get(name)
	return language, st, err
}

// addParts asks the stack's other plugins, then each configured convention
// that applies to the new project, to add their files, then rediscovers. Each
// is given the recipe's settings for it, when the project came from a recipe.
func addParts(s *session.Session, project *projectgraph.Project, st protocol.Stack, params protocol.GenerateParams, settings map[string]map[string]any, res *protocol.GenerateResult, pending *bool, out io.Writer) error {
	var parts []*host.Client
	for _, n := range st.Plugins() {
		if n != st.Scaffold { // it generated the project
			parts = append(parts, s.Plugins.Get(n))
		}
	}
	conventions := map[string]bool{}
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		conv, configured := s.Config.Conventions[name]
		if !configured || slices.Contains(st.Plugins(), name) {
			continue
		}
		targets, err := projectgraph.ConventionTargets(name, []*projectgraph.Project{project}, c.Info.Languages, conv)
		if err != nil {
			return err
		}
		if len(targets) > 0 {
			parts = append(parts, c)
			conventions[name] = true
		}
	}
	ran := false
	for _, c := range parts {
		if !c.Supports(protocol.MethodGenerate) {
			continue
		}
		fmt.Fprintf(out, "Adding %s\n", c.Info.Plugin.Name)
		asked := params
		asked.Convention = conventions[c.Info.Plugin.Name]
		asked.Settings = settings[c.Info.Plugin.Name]
		var added protocol.GenerateResult
		if err := c.Call(protocol.MethodGenerate, asked, &added, 20*time.Minute); err != nil {
			return fmt.Errorf("plugin %s: %w", c.Info.Plugin.Name, err)
		}
		res.Files = append(res.Files, added.Files...)
		res.Notes = append(res.Notes, added.Notes...)
		if added.InstallPending {
			*pending = true
		}
		ran = true
	}
	if !ran {
		return nil
	}
	return s.Refresh()
}

// installPending has the language plugin install the workspace once, when any
// plugin of the project left its install to the host: a Node stack of four
// plugins runs npm install once, not four times.
func installPending(language *host.Client, pending bool, res *protocol.GenerateResult, out io.Writer) error {
	if !pending {
		return nil
	}
	name := language.Info.Plugin.Name
	if !language.Supports(protocol.MethodInstall) {
		return fmt.Errorf("plugin %s left the dependencies to install but cannot install them", name)
	}
	var installed protocol.DependencyResult
	if err := language.Call(protocol.MethodInstall, protocol.InstallParams{Quiet: true}, &installed, 20*time.Minute); err != nil {
		return fmt.Errorf("installing dependencies with %s: %w", name, err)
	}
	res.Files = append(res.Files, installed.Files...)
	res.Notes = append(res.Notes, installed.Notes...)
	return nil
}
