package tui

import (
	"context"
	"io"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/affected"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pluginsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/recipesview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Sources load what the shell's views show; tests replace them.
type Sources struct {
	// Inline prepares a task run that takes over the terminal (see
	// shell_inline_run_use_case.go); nil runs it captured in the Run tab.
	Inline   func(argv []string) *runner.Inline
	Doctor   func() doctorview.Result
	Release  func() releaseview.Result
	Settings func() settingsview.Result
	// Projects lists the workspace's projects with their tasks and what the
	// current changes affect.
	Projects func() projectsLoaded
	// Plugins lists the built-in and installed plugins and what uses them.
	Plugins func() pluginsview.Result
	// Deps looks for newer releases and applies the one-version policy.
	Deps func() depsview.Result
	// Recipes lists the recipe files of the workspace and which recipe makes each kind of project.
	Recipes func() recipesview.Loaded
	// DescribeRecipe loads a recipe and resolves it for some answers, without writing anything.
	DescribeRecipe func(ref string, answers map[string]string) recipesview.Detail
	// Plan lists the tasks a run command line will start, in order, so the
	// progress bar knows its total before the first task starts. A nil or
	// failing Plan only means the total grows as tasks appear.
	Plan func(argv []string) []string
	// Start launches a smith command line as a child process (the smith
	// binary itself) in the workspace.
	Start func(argv []string) (runner.Handle, error)
}

// projectsLoaded is the workspace's projects, loaded in the background.
type projectsLoaded struct {
	projects []projectlist.Project
	branch   string
	err      error
}

// WorkspaceSources open the workspace in dir for each view, as the CLI
// commands do; plugin output is dropped so it cannot tear the screen.
func WorkspaceSources(dir, version string) Sources {
	open := func() (*session.Session, error) {
		return session.Open(session.Options{Dir: dir, Version: version, Log: io.Discard})
	}
	return Sources{
		Start: func(argv []string) (runner.Handle, error) {
			p, err := runner.Start(context.Background(), runner.Spec{Argv: argv, Dir: dir})
			if err != nil {
				return nil, err
			}
			return p, nil
		},
		Inline: func(argv []string) *runner.Inline {
			return runner.NewInline(runner.Spec{Argv: argv, Dir: dir})
		},
		Projects: func() projectsLoaded { return loadProjects(open) },
		Plugins:  func() pluginsview.Result { return pluginsview.Load(open, version) },
		Deps:     func() depsview.Result { return depsview.Load(open) },
		Recipes: func() recipesview.Loaded {
			root, err := workspace.FindRoot(dir)
			if err != nil {
				return recipesview.Loaded{Catalog: recipesview.Catalog{Err: err}}
			}
			return recipesview.Loaded{Catalog: recipesview.ReadCatalog(root)}
		},
		DescribeRecipe: func(ref string, answers map[string]string) recipesview.Detail {
			root, err := workspace.FindRoot(dir)
			if err != nil {
				return recipesview.Detail{Ref: ref, Problem: err.Error()}
			}
			return recipesview.Describe(root, ref, answers)
		},
		Plan: func(argv []string) []string { return planTasks(open, argv) },
		Doctor: func() doctorview.Result {
			s, err := open()
			if err != nil {
				return doctorview.Result{Err: err}
			}
			defer s.Close()
			out := doctorview.Result{Report: doctor.Run(s)}
			for _, name := range s.Graph.Names() {
				out.Projects = append(out.Projects, doctorview.ProjectRef{Name: name, Root: s.Graph.Get(name).Root})
			}
			return out
		},
		Settings: func() settingsview.Result {
			root, err := workspace.FindRoot(dir)
			if err != nil {
				return settingsview.Result{Err: err}
			}
			fields, err := configedit.Fields()
			if err != nil {
				return settingsview.Result{Err: err}
			}
			var out settingsview.Result
			for _, f := range fields {
				value, set, err := configedit.Get(root, f.Path)
				if err != nil {
					return settingsview.Result{Err: err}
				}
				out.Settings = append(out.Settings, settingsview.Setting{Field: f, Value: value, Set: set})
			}
			return out
		},
		Release: func() releaseview.Result {
			s, err := open()
			if err != nil {
				return releaseview.Result{Err: err}
			}
			defer s.Close()
			var out releaseview.Result
			for _, f := range release.CheckLineage(s, false) {
				if f.Blocking {
					out.Blockers = append(out.Blockers, f.Message)
				} else {
					out.Warnings = append(out.Warnings, f.Message)
				}
			}
			out.Plan, out.Err = release.MakePlan(s.Git, s.Graph)
			return out
		},
	}
}

// loadProjects reads the project graph and what the changes since the release
// branch affect.
func loadProjects(open func() (*session.Session, error)) projectsLoaded {
	s, err := open()
	if err != nil {
		return projectsLoaded{err: err}
	}
	defer s.Close()
	base := affected.BaseFromEnv()
	if base == "" {
		base = s.Config.Release.Branch
	}
	hit := map[string]bool{}
	for _, n := range affected.Compute(s.Git, s.Graph, base).Projects {
		hit[n] = true
	}
	var out []projectlist.Project
	for _, name := range s.Graph.Names() {
		p := s.Graph.Get(name)
		np := projectlist.Project{
			Name: name, Kind: string(p.Kind), Language: p.Language, Root: p.Root, Version: p.Version,
			Stack: projectgraph.StackOf(p, s.Config), Deps: p.Deps, Dependents: s.Graph.Dependents(name), Affected: hit[name],
		}
		var tasks []string
		for t := range p.Tasks {
			tasks = append(tasks, t)
		}
		sort.Strings(tasks)
		for _, t := range tasks {
			def := p.Tasks[t]
			np.Tasks = append(np.Tasks, projectlist.Task{
				Name: t, Command: strings.Join(def.Command, " "), DependsOn: def.DependsOn, Cache: def.Cache, Description: def.Description,
			})
		}
		out = append(out, np)
	}
	return projectsLoaded{projects: out, branch: s.Git.CurrentBranch()}
}

// planTasks lists the task IDs a run command line will start, in dependency
// order, selecting projects the way `smith run` does; nil when it cannot tell.
func planTasks(open func() (*session.Session, error), argv []string) []string {
	req, ok := commands.ParseRun(argv)
	if !ok {
		return nil
	}
	s, err := open()
	if err != nil {
		return nil
	}
	defer s.Close()
	names, err := s.Graph.Select(req.Projects)
	if err != nil {
		return nil
	}
	if req.Affected {
		base := req.Base
		if base == "" {
			base = affected.BaseFromEnv()
		}
		if base == "" {
			base = s.Config.Release.Branch
		}
		res := affected.Compute(s.Git, s.Graph, base)
		names = slices.DeleteFunc(names, func(n string) bool { return !slices.Contains(res.Projects, n) })
	}
	var withTask []string
	for _, n := range names {
		for _, t := range req.Tasks {
			if _, ok := s.Graph.Get(n).Tasks[t]; ok {
				withTask = append(withTask, n)
				break
			}
		}
	}
	if len(withTask) == 0 {
		return nil
	}
	g, err := taskgraph.Build(s.Graph, withTask, req.Tasks)
	if err != nil {
		return nil
	}
	return slices.Clone(g.Order)
}
