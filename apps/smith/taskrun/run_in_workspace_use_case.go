package taskrun

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/affected"
	"github.com/monosmith-app/monosmith/apps/smith/runevents"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

// Request is what to run in an open workspace.
type Request struct {
	Tasks []string
	// Projects are selection patterns; empty means every project.
	Projects []string
	// Affected narrows the selection to projects affected since Base.
	Affected bool
	Base     string
	Parallel int
	Bail     bool
	NoCache  bool
	Out      io.Writer
	// Events, when set, receives the run's event stream. RunID names the run
	// in it; empty means a generated one.
	Events runevents.Sink
	RunID  string
}

// Command is the smith command line the request stands for.
func (r Request) Command() string {
	parts := append([]string{"smith", "run"}, r.Tasks...)
	if len(r.Projects) > 0 {
		parts = append(parts, "-p", strings.Join(r.Projects, ","))
	}
	if r.Affected {
		parts = append(parts, "--affected")
		if r.Base != "" {
			parts = append(parts, "--base", r.Base)
		}
	}
	if r.Bail {
		parts = append(parts, "--bail")
	}
	if r.NoCache {
		parts = append(parts, "--no-cache")
	}
	return strings.Join(parts, " ")
}

// RunInDirectory opens the workspace containing dir, runs the request in it and
// closes it again: what `smith run` and `smith serve` both do. With Affected
// and no Base, the base is the pull request's target, else the release branch.
func RunInDirectory(ctx context.Context, dir, version string, log io.Writer, req Request) (Summary, error) {
	s, err := session.Open(session.Options{Dir: dir, Version: version, Log: log})
	if err != nil {
		return Summary{}, err
	}
	defer s.Close()
	if req.Affected && req.Base == "" {
		req.Base = affected.BaseFromEnv()
		if req.Base == "" {
			req.Base = s.Config.Release.Branch
		}
	}
	return RunInWorkspace(ctx, s, req)
}

// RunInWorkspace selects projects, builds the task graph and runs it.
func RunInWorkspace(ctx context.Context, s *session.Session, req Request) (Summary, error) {
	if req.RunID == "" {
		req.RunID = runevents.NewRunID()
	}
	stream := runevents.NewRunStream(req.RunID, req.Events)
	// A run with nothing to do is still a run: it starts and finishes.
	nothing := func() (Summary, error) {
		ok := true
		stream.Emit(runevents.Event{Event: runevents.RunStarted, Command: req.Command()})
		stream.Emit(runevents.Event{Event: runevents.RunFinished, OK: &ok, Summary: runevents.Summarize(0, 0, 0, 0, 0)})
		return Summary{}, nil
	}
	names, err := s.Graph.Select(req.Projects)
	if err != nil {
		return Summary{}, err
	}
	if req.Affected {
		res := affected.Compute(s.Git, s.Graph, req.Base)
		fmt.Fprintf(req.Out, "affected: %s → %d project(s)\n", res.Reason, len(res.Projects))
		keep := map[string]bool{}
		for _, n := range res.Projects {
			keep[n] = true
		}
		var narrowed []string
		for _, n := range names {
			if keep[n] {
				narrowed = append(narrowed, n)
			}
		}
		names = narrowed
	}
	if len(names) == 0 {
		fmt.Fprintln(req.Out, "nothing to run")
		return nothing()
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
		fmt.Fprintf(req.Out, "no selected project has a %v task\n", req.Tasks)
		return nothing()
	}
	g, err := taskgraph.Build(s.Graph, withTask, req.Tasks)
	if err != nil {
		return Summary{}, err
	}
	store, err := s.Storage()
	if err != nil {
		return Summary{}, err
	}
	return Run(ctx, g, Options{
		Root: s.Root, Files: s.Files, Projects: s.Graph, Storage: store,
		Parallel: req.Parallel, Bail: req.Bail, NoCache: req.NoCache, Out: req.Out, ToolVersion: s.Version,
		Events: stream, Command: req.Command(),
	}), nil
}
