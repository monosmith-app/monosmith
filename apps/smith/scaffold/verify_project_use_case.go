package scaffold

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
)

// verifyOrder are the tasks that prove a new project works: its build and its
// first test, for the ones it has.
var verifyOrder = []string{"build", "test"}

// verifyTasks are the tasks smith add runs once on a new project: its build
// and its tests, whichever it has (a Bicep project has neither: nothing to
// prove).
func verifyTasks(p *projectgraph.Project) []string {
	if p == nil {
		return nil
	}
	var tasks []string
	for _, task := range verifyOrder {
		if _, ok := p.Tasks[task]; ok {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// VerifyProject runs the new project's build and tests once, so smith add
// never reports success for a scaffold that fails on its first run. It says
// what it ran; if a task fails, the output is printed and an error returned.
func VerifyProject(ctx context.Context, s *session.Session, p *projectgraph.Project, out io.Writer) error {
	tasks := verifyTasks(p)
	if len(tasks) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	what := strings.Join(tasks, " and ")
	fmt.Fprintf(out, "Verifying: smith run %s -p %s\n", strings.Join(tasks, " "), p.Name)
	var log bytes.Buffer
	summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: tasks, Projects: []string{p.Name}, NoCache: true, Out: &log})
	if err == nil && summary.OK() {
		fmt.Fprintf(out, "Verified: %s passed in %s\n", what, summary.Duration.Round(100*time.Millisecond))
		return nil
	}
	fmt.Fprintln(out, strings.TrimRight(log.String(), "\n"))
	if err != nil {
		return fmt.Errorf("the new project %s could not be verified: %w", p.Name, err)
	}
	return fmt.Errorf("the new project %s fails its first %s (smith run %s -p %s shows why)", p.Name, what, strings.Join(tasks, " "), p.Name)
}
