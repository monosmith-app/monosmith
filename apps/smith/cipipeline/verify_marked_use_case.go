package cipipeline

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
)

// verifyMarked runs, after the verify tasks, the tasks a plugin marked for
// verify (protocol.Task.Verify: an app's image), each for the affected
// projects that have it marked, native ones left to the native job. A task
// whose program is not installed is skipped with a warning.
func verifyMarked(ctx context.Context, s *session.Session, base string, out io.Writer) ([]taskrun.Result, error) {
	byTask, missing := MarkedForVerify(s.Graph.Projects, func(program string) bool {
		_, err := exec.LookPath(program)
		return err == nil
	})
	for _, m := range missing {
		warn(out, m)
	}
	names := make([]string, 0, len(byTask))
	for name := range byTask {
		names = append(names, name)
	}
	sort.Strings(names)
	var results []taskrun.Result
	for _, name := range names {
		summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: []string{name}, Projects: byTask[name], Affected: base != "", Base: base, Out: out})
		if err != nil {
			return results, err
		}
		if len(summary.Results) > 0 {
			fmt.Fprintln(out, summary.String())
		}
		results = append(results, summary.Results...)
		if !summary.OK() {
			return results, taskrun.ErrTasksFailed
		}
	}
	return results, nil
}
