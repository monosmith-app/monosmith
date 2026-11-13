package taskrun

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/lintfix"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// runFix is what a lint task does under `smith lint --fix`: it runs the fix its
// plugin declared, learns what is left, and records the project's outcome. It
// never touches the cache. The task fails when a tool could not run or a
// problem is left, so the run exits non-zero.
func (r *runner) runFix(ctx context.Context, t *taskgraph.Task, key string) Result {
	began := time.Now()
	o := r.fix(ctx, t, key)
	r.mu.Lock()
	r.fixes[t.ID] = o
	r.mu.Unlock()
	took := time.Since(began)
	switch {
	case o.Crash != "":
		r.say("%s × could not run: %s\n", t.ID, o.Crash)
		return Result{Task: t.ID, Status: Failed, Key: key, Duration: took, Reason: "could not run: " + o.Crash}
	case o.Problems() > 0:
		r.say("%s ▲ %s\n", t.ID, leftText(o))
		return Result{Task: t.ID, Status: Failed, Key: key, Duration: took, Reason: leftText(o)}
	}
	r.say("%s ✓ nothing left, %s\n", t.ID, took.Round(time.Millisecond))
	return Result{Task: t.ID, Status: Succeeded, Key: key, Duration: took}
}

// leftText says what is left in one project.
func leftText(o lintfix.Outcome) string {
	if !o.Counted {
		return "problems left"
	}
	return fmt.Sprintf("%d problems left", o.Problems())
}

// fix runs the fix of one lint task and judges the result. A tool that counts
// (protocol.TaskFix.Count) is read from its own report; any other is judged by
// the exit code of the task's plain lint, run after the fix.
func (r *runner) fix(ctx context.Context, t *taskgraph.Task, key string) lintfix.Outcome {
	f := t.Def.Fix
	o := lintfix.Outcome{Project: t.Project.Name, Fixed: -1}
	if f.Count == protocol.FixCountESLintJSON {
		return r.fixESLint(ctx, t, key, o)
	}
	if _, _, _, err := r.capture(ctx, t, key, f.Command); err != nil {
		o.Crash = err.Error()
		o.Remedy = "run " + lintfix.LintCommand(o.Project) + " to see the whole error"
		return o
	}
	_, _, code, err := r.capture(ctx, t, key, t.Def.Command)
	if err != nil {
		o.Crash = err.Error()
		o.Remedy = "run " + lintfix.LintCommand(o.Project) + " to see the whole error"
		return o
	}
	o.Left = code != 0
	return o
}

// fixESLint fixes with ESLint and counts from its JSON report: before the fix
// when the task declares a probe, and after it.
func (r *runner) fixESLint(ctx context.Context, t *taskgraph.Task, key string, o lintfix.Outcome) lintfix.Outcome {
	f := t.Def.Fix
	version := lintfix.InstalledESLint(r.opts.Root)
	crash := func(cause string) lintfix.Outcome {
		o.Crash, o.Remedy = cause, lintfix.ESLintRemedy(o.Project, version)
		return o
	}
	if version == "" {
		return crash(lintfix.ESLintNotInstalled)
	}
	before := -1
	if len(f.Probe) > 0 {
		errs, warns, cause := r.countESLint(ctx, t, key, f.Probe)
		if cause != "" {
			return crash(cause)
		}
		before = errs + warns
	}
	errs, warns, cause := r.countESLint(ctx, t, key, f.Command)
	if cause != "" {
		return crash(cause)
	}
	o.Counted, o.Errors, o.Warnings = true, errs, warns
	if before >= 0 {
		o.Fixed = max(0, before-(errs+warns))
	}
	return o
}

// countESLint runs an ESLint command and reads its JSON report. cause is the
// reason it could not run, "" when it did.
func (r *runner) countESLint(ctx context.Context, t *taskgraph.Task, key string, argv []string) (errs, warns int, cause string) {
	stdout, stderr, code, err := r.capture(ctx, t, key, argv)
	if err != nil {
		return 0, 0, err.Error()
	}
	return lintfix.ReadESLint(stdout, stderr, code)
}

// capture runs a command of the task, in the task's folder and environment,
// with its two outputs apart. err is set when the program could not start.
func (r *runner) capture(ctx context.Context, t *taskgraph.Task, key string, argv []string) (stdout, stderr string, code int, err error) {
	var so, se bytes.Buffer
	code, err = RunProcess(ctx, Spec{Argv: r.expand(t, argv), Dir: r.cwd(t), Env: r.env(t, key), Out: &so, Err: &se, Installable: r.opts.Installable})
	return so.String(), se.String(), code, err
}

// fixOutcomes lists the projects' outcomes after a fixing run, in the order of
// the graph: the ones the fix judged, and, for a lint task whose plugin has no
// fix, the plain lint's result.
func (r *runner) fixOutcomes() []lintfix.Outcome {
	var out []lintfix.Outcome
	for _, id := range r.g.Order {
		t := r.g.Tasks[id]
		if t.Name != taskgraph.LintTask {
			continue
		}
		if o, ok := r.fixes[id]; ok {
			out = append(out, o)
			continue
		}
		res := r.results[id]
		out = append(out, lintfix.Outcome{Project: t.Project.Name, Fixed: -1, NoFix: true, Left: res.Status == Failed, Skipped: res.Status == Skipped})
	}
	return out
}

// announceFix is the banner of a fixing run: what it is about to do and in how
// many projects, or why --fix changes nothing when no task is a lint.
func announceFix(out io.Writer, g *taskgraph.Graph, tasks []string) {
	lints := 0
	for _, t := range g.Tasks {
		if t.Name == taskgraph.LintTask && t.Requested {
			lints++
		}
	}
	if lints == 0 {
		fmt.Fprintf(out, "--fix changes only lint, and %s has no lint task: nothing is fixed\n", strings.Join(tasks, ", "))
		return
	}
	fmt.Fprintf(out, "Lint fix: fixing what the linters can in %d projects, without the cache\n", lints)
}
