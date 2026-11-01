package sources

import (
	"context"
	"io"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/affected"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Sources load what the shell's views show; tests replace them.
type Sources struct {
	Doctor  func() doctorview.Result
	Release func() releaseview.Result
	// Projects lists the workspace's projects with their tasks and what the
	// current changes affect.
	Projects func() ProjectsLoaded
	// Deps looks for newer releases and applies the one-version policy.
	Deps func() depsview.Result
	// Plan lists the tasks a run command line will start, in order, so the
	// progress bar knows its total before the first task starts. A nil or
	// failing Plan only means the total grows as tasks appear.
	Plan func(argv []string) []string
	// ShowFix writes a planned fix as the wizard's review: its description,
	// the diff of each file, the commands it runs.
	ShowFix func(doctorfix.Planned) (string, error)
	// ApplyFix writes a fix's files and runs its steps from the workspace
	// root, stopping at the first that fails (a *doctorfix.StepError). It
	// writes a line "… name" to out as each step starts, and honours ctx.
	ApplyFix func(ctx context.Context, fix protocol.Fix, out io.Writer) error
	// Settings is what the Settings tab reads and changes.
	Settings settingsview.Backend
	// Start launches a smith command line as a child process (the smith
	// binary itself) in the workspace.
	Start func(argv []string) (runner.Handle, error)
	// Capture runs a smith command line to its end and returns what it
	// printed, for the plans and errors the shell shows in a dialog.
	Capture func(argv []string) Captured
	// Open opens a file with the program the system chooses (the report).
	Open func(path string) error
}

// ProjectsLoaded is the workspace's projects, loaded in the background.
type ProjectsLoaded struct {
	Projects []projectlist.Project
	Branch   string
	Err      error
}

// WorkspaceSources open the workspace in dir for each view, as the CLI
// commands do; plugin output is dropped so it cannot tear the screen.
func WorkspaceSources(dir, version string) Sources {
	open := func() (*session.Session, error) {
		return session.Open(session.Options{Dir: dir, Version: version, Log: io.Discard})
	}
	return Sources{
		Capture: func(argv []string) Captured { return CaptureCommand(dir, argv) },
		Open:    OpenPath,
		Start: func(argv []string) (runner.Handle, error) {
			p, err := runner.Start(context.Background(), runner.Spec{Argv: withEventStream(argv), Dir: dir})
			if err != nil {
				return nil, err
			}
			return p, nil
		},
		ShowFix: func(p doctorfix.Planned) (string, error) {
			root, err := fixRoot(dir)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			err = doctorfix.ShowFix(&b, root, p)
			return b.String(), err
		},
		ApplyFix: func(ctx context.Context, fix protocol.Fix, out io.Writer) error {
			root, err := fixRoot(dir)
			if err != nil {
				return err
			}
			return doctorfix.Apply(ctx, root, fix, out)
		},
		Projects: func() ProjectsLoaded { return loadProjects(open) },
		Deps:     func() depsview.Result { return depsview.Load(open) },
		Plan:     func(argv []string) []string { return planTasks(open, argv) },
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
			out.Plan, out.Err = release.PlanFor(s)
			return out
		},
	}
}

// fixRoot is the workspace root a fix's paths are relative to.
func fixRoot(dir string) (string, error) { return workspace.FindRoot(dir) }

// loadProjects reads the project graph and what the changes since the release
// branch affect.
func loadProjects(open func() (*session.Session, error)) ProjectsLoaded {
	s, err := open()
	if err != nil {
		return ProjectsLoaded{Err: err}
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
	return ProjectsLoaded{Projects: out, Branch: s.Git.CurrentBranch()}
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

// withEventStream asks a task run for its event stream (`--json-events`): the
// Run tab then reads each task's state and output as structured events instead
// of parsing smith's text. The command shown and copied stays the one the user
// would type; any other command is started as it is.
func withEventStream(argv []string) []string {
	if _, ok := commands.ParseRun(argv); !ok {
		return argv
	}
	for _, a := range argv {
		if a == "--json" || a == "--json-events" {
			return argv
		}
	}
	return append(append([]string(nil), argv...), "--json-events")
}
