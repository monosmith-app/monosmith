package taskrun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

// Options configure a run.
type Options struct {
	Root string
	// Files are the workspace's files (tracked, or untracked and not ignored).
	Files    []string
	Projects *projectgraph.Graph
	Storage  taskcache.Storage
	// Parallel is the number of tasks running at once; 0 means the CPU count.
	Parallel int
	// Bail stops starting tasks after the first failure.
	Bail bool
	// NoCache runs every task, still saving results.
	NoCache bool
	Out     io.Writer
	// ToolVersion goes into every key.
	ToolVersion string
}

// Run executes a task graph and returns what happened to every task.
func Run(ctx context.Context, g *taskgraph.Graph, opts Options) Summary {
	start := time.Now()
	if opts.Parallel <= 0 {
		opts.Parallel = runtime.NumCPU()
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	r := &runner{opts: opts, g: g, keyer: newKeyer(opts), sem: make(chan struct{}, opts.Parallel),
		done: map[string]chan struct{}{}, results: map[string]Result{}, keys: map[string]string{}}
	for _, id := range g.Order {
		r.done[id] = make(chan struct{})
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	r.cancel = cancel
	var wg sync.WaitGroup
	for _, id := range g.Order {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(r.done[id])
			r.record(r.runTask(ctx, g.Tasks[id]))
		}()
	}
	wg.Wait()
	s := Summary{Duration: time.Since(start)}
	for _, id := range g.Order {
		s.Results = append(s.Results, r.results[id])
	}
	return s
}

type runner struct {
	opts    Options
	g       *taskgraph.Graph
	keyer   *keyer
	sem     chan struct{}
	done    map[string]chan struct{}
	mu      sync.Mutex
	outMu   sync.Mutex
	results map[string]Result
	keys    map[string]string
	failed  bool
	cancel  context.CancelFunc
}

func (r *runner) record(res Result) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results[res.Task] = res
	r.keys[res.Task] = res.Key
	if res.Status == Failed {
		r.failed = true
		if r.opts.Bail {
			r.cancel()
		}
	}
}

func (r *runner) runTask(ctx context.Context, t *taskgraph.Task) Result {
	for _, d := range t.Deps {
		<-r.done[d]
	}
	r.mu.Lock()
	for _, d := range t.Deps {
		if st := r.results[d].Status; st == Failed || st == Skipped {
			r.mu.Unlock()
			return Result{Task: t.ID, Status: Skipped, Reason: "because " + d + " did not succeed"}
		}
	}
	if r.failed && r.opts.Bail {
		r.mu.Unlock()
		return Result{Task: t.ID, Status: Skipped, Reason: "after an earlier failure (--bail)"}
	}
	r.mu.Unlock()

	select {
	case r.sem <- struct{}{}:
	case <-ctx.Done():
		return Result{Task: t.ID, Status: Skipped, Reason: "cancelled"}
	}
	defer func() { <-r.sem }()

	key, inputs := r.key(t)
	prefix := t.ID + " │ "
	if t.Def.Cache && !r.opts.NoCache && r.opts.Storage != nil {
		if entry, err := taskcache.Lookup(r.opts.Storage, key); err == nil {
			log, err := taskcache.Restore(r.opts.Storage, r.opts.Root, t.Project.Root, t.Def.Outputs, entry)
			if err == nil {
				_ = taskcache.Remember(r.opts.Storage, t.ID, key)
				Replay(prefix, log, r.opts.Out, &r.outMu)
				r.say("%s ✔ from cache\n", t.ID)
				return Result{Task: t.ID, Status: Cached, Key: key, Saved: entry.Duration}
			}
		}
	}

	r.say("%s ▶ %s\n", t.ID, strings.Join(t.Def.Command, " "))
	for _, o := range t.Def.Outputs {
		_ = os.RemoveAll(filepath.Join(r.opts.Root, filepath.FromSlash(path.Join(t.Project.Root, o))))
	}
	out := newPrefixWriter(prefix, r.opts.Out, &r.outMu)
	began := time.Now()
	var app *served
	if t.Def.Serve != "" {
		var err error
		if app, err = r.startServe(ctx, t, key); err != nil {
			r.say("%s ✖ %v\n", t.ID, err)
			return Result{Task: t.ID, Status: Failed, Key: key, Duration: time.Since(began), Reason: err.Error()}
		}
	}
	code, err := RunProcess(ctx, Spec{Argv: r.expand(t, t.Def.Command), Dir: r.cwd(t), Env: r.env(t, key), Out: out})
	out.Flush()
	if app != nil {
		app.stop()
		r.say("%s ■ stopped\n", app.id)
	}
	took := time.Since(began)
	r.keyer.digests.Forget()
	if err != nil || code != 0 {
		reason := fmt.Sprintf("exited with code %d", code)
		if err != nil {
			reason = err.Error()
			if errors.Is(err, context.Canceled) {
				reason = "cancelled"
			}
		}
		r.say("%s ✖ %s\n", t.ID, reason)
		return Result{Task: t.ID, Status: Failed, Key: key, Duration: took, Reason: reason}
	}
	if t.Def.Cache && r.opts.Storage != nil {
		if err := taskcache.Save(r.opts.Storage, r.opts.Root, key, t.ID, t.Project.Root, t.Def.Outputs, out.log.Bytes(), took, &inputs); err != nil {
			r.say("%s ! not cached: %v\n", t.ID, err)
		}
	}
	r.say("%s ✔ %s\n", t.ID, took.Round(time.Millisecond))
	return Result{Task: t.ID, Status: Succeeded, Key: key, Duration: took}
}

// key computes the task's cache key once the tasks it depends on have run.
func (r *runner) key(t *taskgraph.Task) (string, taskcache.KeyInputs) {
	r.mu.Lock()
	deps := make(map[string]string, len(t.Deps))
	for _, d := range t.Deps {
		deps[d] = r.keys[d]
	}
	r.mu.Unlock()
	in := r.keyer.inputs(t, deps)
	return taskcache.Key(in), in
}

func (r *runner) expand(t *taskgraph.Task, argv []string) []string {
	root := filepath.ToSlash(r.opts.Root)
	repl := strings.NewReplacer("{smith}", Self(), "{workspaceRoot}", root, "{projectRoot}", root+"/"+t.Project.Root)
	out := make([]string, len(argv))
	for i, a := range argv {
		out[i] = repl.Replace(a)
	}
	return out
}

// Self is the path of the running smith, which {smith} in a command stands
// for: a convention's task runs the same smith that runs it, on PATH or not.
func Self() string {
	exe, err := os.Executable()
	if err != nil {
		return "smith"
	}
	return filepath.ToSlash(exe)
}

func (r *runner) cwd(t *taskgraph.Task) string {
	dir := t.Project.Root
	if t.Def.Cwd != "" {
		dir = t.Def.Cwd
	}
	return filepath.Join(r.opts.Root, filepath.FromSlash(dir))
}

func (r *runner) env(t *taskgraph.Task, key string) []string {
	extra := map[string]string{"SMITH_TASK": t.ID, "SMITH_TASK_KEY": key, "SMITH_WORKSPACE_ROOT": r.opts.Root}
	// The workspace's Python environment comes first on PATH, as
	// node_modules/.bin does for npm scripts, so `python` is the venv's.
	if bin := VenvBin(r.opts.Root); bin != "" {
		extra["PATH"] = bin + string(os.PathListSeparator) + os.Getenv("PATH")
		extra["VIRTUAL_ENV"] = filepath.Dir(bin)
	}
	if t.Def.ServeURL != "" {
		extra["E2E_BASE_URL"] = t.Def.ServeURL
	}
	for k, v := range t.Def.Env {
		extra[k] = r.expand(t, []string{v})[0]
	}
	return Environment(extra)
}

func (r *runner) say(format string, args ...any) {
	r.outMu.Lock()
	defer r.outMu.Unlock()
	fmt.Fprintf(r.opts.Out, format, args...)
}
