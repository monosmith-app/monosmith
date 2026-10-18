package scaffold

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
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
	Out   io.Writer
}

// AddProject generates a project: its stack's scaffold (or the language
// plugin) writes it and records the stack, then each other plugin of the
// stack and each configured convention that applies adds its part (lint,
// tests), and the pipelines are regenerated so CI knows about the new
// toolchain at once.
func AddProject(s *session.Session, opts AddOptions) (*protocol.GenerateResult, error) {
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
	language, st, err := chooseStack(s, opts)
	if err != nil {
		return nil, err
	}
	generator := language
	if st.Scaffold != "" {
		generator = s.Plugins.Get(st.Scaffold)
	}
	if !generator.Supports(protocol.MethodGenerate) {
		return nil, fmt.Errorf("plugin %s cannot generate projects", generator.Info.Plugin.Name)
	}
	dir := projectkind.Directory(kind) + "/" + opts.Name
	if _, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(dir))); err == nil {
		return nil, fmt.Errorf("%s already exists", dir)
	}
	how := language.Info.Plugin.Name
	if st.Name != "" && st.Name != protocol.NoStack {
		how = "stack " + st.Name
	}
	fmt.Fprintf(opts.Out, "Adding %s %s (%s) in %s\n", opts.Kind, opts.Name, how, dir)
	var res protocol.GenerateResult
	params := protocol.GenerateParams{Kind: kind, Name: opts.Name, Directory: dir, Stack: st.Name, Workspace: protocol.WorkspaceInfo{
		Name: s.Config.Workspace.Name, Scope: s.Config.Workspace.Scope, ModulePrefix: s.Config.Workspace.ModulePrefix,
	}}
	if err := generator.Call(protocol.MethodGenerate, params, &res, 20*time.Minute); err != nil {
		return nil, err
	}
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
	if err := addParts(s, project, st, params, &res, opts.Out); err != nil {
		return &res, err
	}
	changed, _, err := cipipeline.Regenerate(s)
	if err != nil {
		return &res, err
	}
	res.Files = append(res.Files, changed...)
	sort.Strings(res.Files)
	return &res, nil
}

// chooseStack resolves the language plugin and the stack of a new project.
func chooseStack(s *session.Session, opts AddOptions) (*host.Client, protocol.Stack, error) {
	if opts.Stack == protocol.NoStack {
		lang, err := chooseLanguage(s, opts.Lang)
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
	lang, err := chooseLanguage(s, opts.Lang)
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

func chooseLanguage(s *session.Session, lang string) (string, error) {
	if lang != "" {
		return lang, nil
	}
	used := map[string]bool{}
	for _, p := range s.Graph.Projects {
		used[p.Plugin] = true
	}
	if len(used) == 1 {
		for l := range used {
			return l, nil
		}
	}
	return "", fmt.Errorf("choose a language with --lang (%s) or a stack with --stack", strings.Join(s.Plugins.Languages(), ", "))
}

// addParts asks the stack's other plugins, then each configured convention
// that applies to the new project, to add their files, then rediscovers.
func addParts(s *session.Session, project *projectgraph.Project, st protocol.Stack, params protocol.GenerateParams, res *protocol.GenerateResult, out io.Writer) error {
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
		var added protocol.GenerateResult
		if err := c.Call(protocol.MethodGenerate, asked, &added, 20*time.Minute); err != nil {
			return fmt.Errorf("plugin %s: %w", c.Info.Plugin.Name, err)
		}
		res.Files = append(res.Files, added.Files...)
		res.Notes = append(res.Notes, added.Notes...)
		ran = true
	}
	if !ran {
		return nil
	}
	return s.Refresh()
}
