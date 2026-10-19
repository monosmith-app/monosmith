package taskrun

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

// ServeTimeout is how long a served app has to answer when the task sets no
// serve_timeout.
const ServeTimeout = 60 * time.Second

// served is an app started around a task.
type served struct {
	id     string
	cancel context.CancelFunc
	exited chan servedExit
}

type servedExit struct {
	code int
	err  error
}

// stop ends the app and its whole process tree, and waits until it is gone.
func (s *served) stop() {
	s.cancel()
	<-s.exited
}

// startServe starts the app t serves, in the background, and waits until its
// serve_url answers with a 2xx. The app's output is prefixed with its own
// task id. The caller stops it with stop, whatever the task's result.
func (r *runner) startServe(ctx context.Context, t *taskgraph.Task, key string) (*served, error) {
	project, name, err := taskgraph.ServeTarget(r.opts.Projects, t.Project.Name, t.Def.Serve)
	if err != nil {
		return nil, err
	}
	timeout := ServeTimeout
	if t.Def.ServeTimeout != "" {
		if timeout, err = time.ParseDuration(t.Def.ServeTimeout); err != nil || timeout <= 0 {
			return nil, fmt.Errorf("serve_timeout %q is not a duration like 60s", t.Def.ServeTimeout)
		}
	}
	st := &taskgraph.Task{ID: taskgraph.ID(project.Name, name), Project: project, Name: name, Def: project.Tasks[name]}
	sctx, cancel := context.WithCancel(ctx)
	app := &served{id: st.ID, cancel: cancel, exited: make(chan servedExit, 1)}
	out := newPrefixWriter(st.ID+" │ ", r.opts.Out, &r.outMu)
	r.say("%s ▶ %s (served for %s)\n", st.ID, strings.Join(st.Def.Command, " "), t.ID)
	go func() {
		code, err := RunProcess(sctx, Spec{Argv: r.expand(st, st.Def.Command), Dir: r.cwd(st), Env: r.env(st, key), Out: out})
		out.Flush()
		app.exited <- servedExit{code, err}
	}()
	if t.Def.ServeURL == "" {
		return app, nil
	}
	if err := waitForURL(sctx, t.Def.ServeURL, timeout, app.exited); err != nil {
		cancel()
		<-app.exited
		return nil, fmt.Errorf("%s: %w", st.ID, err)
	}
	return app, nil
}

// waitForURL polls url until it answers 2xx. It gives up when the app exits
// first (its exit is sent back on the channel for the caller to drain), when
// the timeout passes or when ctx ends.
func waitForURL(ctx context.Context, url string, timeout time.Duration, exited chan servedExit) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()
	client := &http.Client{Timeout: 2 * time.Second}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("serve_url %q: %w", url, err)
		}
		if resp, err := client.Do(req); err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode/100 == 2 {
				return nil
			}
		}
		select {
		case e := <-exited:
			exited <- e
			return fmt.Errorf("exited (code %d) before %s answered", e.code, url)
		case <-deadline.C:
			return fmt.Errorf("%s did not answer with a 2xx within %s", url, timeout)
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
	}
}
