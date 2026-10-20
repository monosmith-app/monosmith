package engineserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/runevents"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
)

// run runs tasks in the workspace, in process, with the run's events sent as
// notifications while it goes, and answers when it ends. The run is cancelled
// by the cancel method, or when the client goes away.
func (s *Server) run(ctx context.Context, raw json.RawMessage) (any, *rpcError) {
	var p RunParams
	if err := decodeParams(raw, &p); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	if len(p.Tasks) == 0 {
		return nil, &rpcError{codeInvalidParams, "run needs the tasks to run: {\"tasks\": [\"build\"]}"}
	}
	if p.Run == "" {
		p.Run = runevents.NewRunID()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := s.runs.begin(p.Run, cancel); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	defer s.runs.end(p.Run)

	runTasks := s.RunTasks
	if runTasks == nil {
		runTasks = func(ctx context.Context, req taskrun.Request) (taskrun.Summary, error) {
			return taskrun.RunInDirectory(ctx, s.dir(), s.Version, os.Stderr, req)
		}
	}
	summary, err := runTasks(ctx, taskrun.Request{
		Tasks: p.Tasks, Projects: p.Projects, Affected: p.Affected, Base: p.Base, Parallel: p.Parallel,
		Bail: p.Bail, NoCache: p.NoCache, Out: io.Discard, Events: s.sink(), RunID: p.Run,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, &rpcError{codeFailed, "cancelled"}
		}
		return nil, &rpcError{codeFailed, err.Error()}
	}
	return RunResult{Run: p.Run, OK: summary.OK(), Summary: summary}, nil
}
