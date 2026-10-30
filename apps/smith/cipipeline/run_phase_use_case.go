package cipipeline

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/affected"
	"github.com/monosmith-app/monosmith/apps/smith/nativebuild"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/releaseassets"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// VerifyTasks are what a pull request must pass.
var VerifyTasks = []string{"lint", "typecheck", "test", "build"}

// PackTask is the task that builds what a project deploys into dist/drop.
const PackTask = "package"

// DropDir is where packed deliverables go, workspace-relative.
const DropDir = "dist/drop"

// Phases lists `smith ci` phases in pipeline order.
var Phases = []string{"setup", "audit", "verify", "pack", "native", "assets", "release"}

// RunPhase runs one CI phase.
func RunPhase(ctx context.Context, s *session.Session, phase string, out io.Writer) error {
	begin := time.Now()
	log := newPhaseLog(out, phase)
	if !announces(phase) {
		log.fold()
	}
	err := runPhase(ctx, s, phase, log)
	log.close()
	took := time.Since(begin)
	if err != nil {
		annotate(out, phase, err)
	}
	log.finish(err, took)
	summarise(out, log, err, took)
	return err
}

// summarise adds the phase's result to the job summary of the CI that has
// one: GitHub's step summary file, or an Azure summary section.
func summarise(out io.Writer, log *phaseLog, err error, took time.Duration) {
	state := fmt.Sprintf("%s %s passed in %s", log.icons.Passed, log.title(), formatDuration(took))
	if err != nil {
		state = fmt.Sprintf("%s %s failed after %s", log.icons.Failed, log.title(), formatDuration(took))
	}
	text := "### " + log.title() + "\n\n" + state + "\n"
	if err == nil && log.conclusion != "" {
		text += "\n" + log.conclusion + "\n"
	}
	if len(log.markdown) > 0 {
		text += "\n" + strings.Join(log.markdown, "\n") + "\n"
	}
	text += "\n"
	if file := os.Getenv("GITHUB_STEP_SUMMARY"); file != "" {
		f, openErr := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if openErr != nil {
			return
		}
		defer func() { _ = f.Close() }()
		_, _ = f.WriteString(text)
		return
	}
	if os.Getenv("TF_BUILD") != "" {
		dir := os.Getenv("AGENT_TEMPDIRECTORY")
		if dir == "" {
			dir = os.TempDir()
		}
		path := filepath.Join(dir, "smith-ci-"+log.phase+".md")
		if os.WriteFile(path, []byte(text), 0o644) == nil {
			fmt.Fprintf(out, "##vso[task.uploadsummary]%s\n", path)
		}
	}
}

// announces reports whether a phase opens with a banner of its own
// (phaseLog.banner); the others have their whole output folded.
func announces(phase string) bool {
	switch phase {
	case "setup", "audit", "verify", "pack", "release":
		return true
	}
	return false
}

func runPhase(ctx context.Context, s *session.Session, phase string, log *phaseLog) error {
	out := log.out
	switch phase {
	case "setup", "audit":
		c, err := Collect(s)
		if err != nil {
			return err
		}
		steps := c.Setup
		if phase == "audit" {
			steps = c.Audit
		}
		var remembered *verdicts
		intro := setupIntro(s, c, steps)
		if phase == "audit" {
			remembered = openVerdicts(s)
			intro = auditIntro(s, steps)
		}
		if phase == "setup" {
			intro = append([]string{smithLine(s.Version)}, intro...)
		}
		if len(steps) == 0 {
			log.banner()
			log.close()
			log.say("%s", strings.Join(intro, " "))
			log.say("Nothing to %s: no plugin asked for a %s step.", phase, phase)
			log.conclusion = "nothing to " + phase
			return nil
		}
		log.banner(intro...)
		outcomes, err := runSteps(ctx, s.Root, steps, remembered, out)
		log.close()
		if len(outcomes) > 0 {
			table := stepTable(map[string]string{"setup": "Step", "audit": "Check"}[phase], outcomes, log.icons)
			for _, line := range table.Text("steps done") {
				log.say("%s", line)
			}
			log.markdown = table.Markdown()
		}
		if problems := problemNames(outcomes); len(problems) > 0 {
			warn(out, fmt.Sprintf("%s reported problems (not blocking): %s", plural(len(problems), "step"), listSome(problems, 3)))
			log.say("→ read each step's findings in the folded output above")
		}
		if err == nil {
			log.conclusion = setupConclusion(outcomes, len(s.Graph.Projects))
			if phase == "audit" {
				log.conclusion = auditConclusion(outcomes, len(s.Graph.Projects))
			}
		}
		return err
	case "verify":
		base := affected.BaseFromEnv()
		before := uncommitted(s)
		log.banner(verifyIntro(s, base)...)
		if err := verifyBoundaries(s, out); err != nil {
			return err
		}
		// Native projects are verified by the native job, on every OS.
		outputs := newTaskOutputs()
		tee := io.MultiWriter(out, outputs)
		summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: VerifyTasks, Projects: nativeExcluded, Affected: base != "", Base: base, Out: tee})
		if err != nil {
			return err
		}
		if len(summary.Results) > 0 {
			fmt.Fprintln(out, summary.String())
		}
		results, columns := summary.Results, VerifyTasks
		failed := !summary.OK()
		var markedErr error
		if !failed {
			var marked []taskrun.Result
			marked, markedErr = verifyMarked(ctx, s, base, tee)
			results = append(results, marked...)
			for _, r := range marked {
				if !slices.Contains(columns, taskOf(r.Task)) {
					columns = append(slices.Clone(columns), taskOf(r.Task))
				}
			}
		}
		log.close()
		if len(results) == 0 {
			log.say("Nothing to verify: no selected project has a %s task.", strings.Join(VerifyTasks, ", "))
			log.conclusion = "nothing to verify"
		} else {
			report := reportTasks(results, columns, outputs, log.icons)
			for _, line := range report.Table.Text("passed") {
				log.say("%s", line)
			}
			if slices.ContainsFunc(results, func(r taskrun.Result) bool { return r.Status == taskrun.Cached }) {
				log.say("%s = replayed from the cache", log.icons.Cached)
			}
			for _, line := range failureLines(report.Failed, outputs) {
				log.say("%s", line)
			}
			log.markdown = report.Table.Markdown()
			log.conclusion = report.Conclusion
		}
		if failed {
			return taskrun.ErrTasksFailed
		}
		if markedErr != nil {
			return markedErr
		}
		return verifyClean(s, before, summary)
	case "pack":
		// The directory always exists, so the artifact upload has a path.
		if err := os.MkdirAll(filepath.Join(s.Root, filepath.FromSlash(DropDir)), 0o755); err != nil {
			return err
		}
		packing := false
		for _, p := range s.Graph.Projects {
			if _, ok := p.Tasks[PackTask]; ok {
				packing = true
			}
		}
		if !packing {
			log.banner()
			log.say("Nothing to pack: no project has a %s task, so %s stays empty.", PackTask, DropDir)
			log.conclusion = "nothing to pack"
			return nil
		}
		log.banner(packIntro(s)...)
		// Native projects are packed by the native job, on each OS.
		outputs := newTaskOutputs()
		summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: []string{PackTask}, Projects: nativeExcluded, Out: io.MultiWriter(out, outputs)})
		if err != nil {
			return err
		}
		fmt.Fprintln(out, summary.String())
		log.close()
		report := reportTasks(summary.Results, []string{PackTask}, outputs, log.icons)
		for _, line := range report.Table.Text("packed") {
			log.say("%s", line)
		}
		for _, line := range failureLines(report.Failed, outputs) {
			log.say("%s", line)
		}
		if !summary.OK() {
			return taskrun.ErrTasksFailed
		}
		files := listDropFiles(s.Root)
		for _, line := range dropTable(files).Text("files") {
			log.say("%s", line)
		}
		log.markdown = append(report.Table.Markdown(), append([]string{""}, dropTable(files).Markdown()...)...)
		log.conclusion = fmt.Sprintf("%s packed into %s (%s)", plural(len(files), "file"), DropDir, formatSize(totalSize(files)))
		return nil
	case "native":
		return nativebuild.Run(ctx, s, out)
	case "assets":
		done, err := releaseassets.Run(ctx, s, releaseassets.Request{Platform: os.Getenv("SMITH_PLATFORM"), Out: out})
		if err == nil && len(done) == 0 {
			fmt.Fprintln(out, "no release files to build: no project with a build command was released by this commit on this platform")
		}
		return err
	case "release":
		ensureIdentity(s)
		plan, err := release.MakePlan(s.Git, s.Graph)
		if err != nil {
			return err
		}
		projects := len(s.Graph.Projects)
		var c Contributions
		if len(plan.Items) == 0 {
			// Nothing is published, so the plugins can be asked first: fold only
			// when a post-release step has output to fold.
			if c, err = Collect(s); err != nil {
				return err
			}
			log.heading("Looking for commits that release (feat, fix, perf or a breaking change) since each project's last tag.")
			if len(c.PostRelease) > 0 {
				log.fold()
			}
		} else {
			log.heading(releaseIntro(plan, s.Config.Release.Branch, projects)...)
			log.fold()
			for _, item := range plan.Items {
				fmt.Fprintf(out, "%s: %s -> %s (%s) because %s\n", item.Project, cmp.Or(item.Current, "new"), item.Next, bumpOf(item), item.Why())
			}
		}
		report, err := release.Execute(ctx, s, plan, release.Options{Push: true, Out: out})
		var after []stepOutcome
		if err == nil {
			// Work plugins want after every push to the release branch, released
			// or not (a docs-only commit releases nothing).
			if len(plan.Items) > 0 {
				c, err = Collect(s)
			}
			if err == nil {
				after, err = runSteps(ctx, s.Root, c.PostRelease, nil, out)
			}
		}
		log.close()
		log.markdown = strings.Split(strings.TrimSpace(strings.TrimPrefix(report.Markdown(), "### Release")), "\n")
		if len(plan.Items) == 0 {
			log.say("Nothing to release: none of the %s has a releasing commit since its last release.", plural(projects, "project"))
			log.say("→ commit with a message like \"feat: add search\" or \"fix: handle empty input\"; docs, chore, test and ci commits do not release")
			log.conclusion = "nothing to release"
		} else {
			for _, line := range releaseTable(plan, report, log.icons).Text("packages") {
				log.say("%s", line)
			}
			if line := unchangedLine(report.Unchanged); line != "" {
				log.say("%s", paint(line, tintDim, colourOn()))
			}
			for _, res := range report.Results {
				for _, image := range res.Images {
					log.say("Image pushed for %s: %s", res.Project, release.ImageLine(image))
				}
				if res.GitHubRelease != "" {
					log.say("GitHub Release for %s: %s", res.Project, res.GitHubRelease)
				}
			}
			if len(after) > 0 {
				for _, line := range stepTable("After the release", after, log.icons).Text("steps") {
					log.say("%s", line)
				}
			}
			if err != nil {
				log.say("%s", releaseConclusion(report))
				log.say("→ re-run the release: the packages it finished are kept and the rest resume")
			} else {
				log.conclusion = releaseConclusion(report)
				notice(out, releaseConclusion(report))
			}
		}
		return err
	}
	return fmt.Errorf("unknown phase %q (phases: %s)", phase, strings.Join(Phases, ", "))
}

// uncommitted lists the tracked files changed in the working tree, as the
// release's preflight sees them; nil when git cannot say.
func uncommitted(s *session.Session) []string {
	out, err := s.Git.Run("status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return nil
	}
	return release.Uncommitted(out)
}

// verifyClean fails verify when its tasks left tracked files changed: the
// release refuses a dirty tree, and the pull request is where to find out.
// A file already changed before verify fails it only in CI, where checkout
// starts clean; on a laptop it is the developer's own edit.
func verifyClean(s *session.Session, before []string, summary taskrun.Summary) error {
	after := uncommitted(s)
	if len(after) == 0 {
		return nil
	}
	ran := make([]string, 0, len(summary.Results))
	for _, r := range summary.Results {
		ran = append(ran, r.Task)
	}
	var changes []TrackedChange
	for _, c := range TrackedChanges(before, after, s.Graph.Projects, ran) {
		if !c.Before || os.Getenv("CI") != "" {
			changes = append(changes, c)
		}
	}
	if len(changes) == 0 {
		return nil
	}
	return errors.New(ExplainTrackedChanges(changes))
}

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

// taskOf is the task of a task id ("core:test" is "test").
func taskOf(id string) string {
	_, task := projectOf(id)
	return task
}

// MarkedForVerify lists, by task, the projects whose task a plugin marked
// for verify, leaving out native projects, the verify tasks themselves and
// tasks whose program is not installed (named in missing, for a warning).
func MarkedForVerify(projects []*projectgraph.Project, installed func(string) bool) (byTask map[string][]string, missing []string) {
	byTask = map[string][]string{}
	for _, p := range projects {
		if slices.Contains(p.Tags, release.NativeTag) {
			continue
		}
		tasks := make([]string, 0, len(p.Tasks))
		for name := range p.Tasks {
			tasks = append(tasks, name)
		}
		sort.Strings(tasks)
		for _, name := range tasks {
			t := p.Tasks[name]
			if !t.Verify || slices.Contains(VerifyTasks, name) || len(t.Command) == 0 {
				continue
			}
			if program := t.Command[0]; !strings.HasPrefix(program, "{") && !installed(program) {
				missing = append(missing, fmt.Sprintf("%s:%s not verified: %s is not installed", p.Name, name, program))
				continue
			}
			byTask[name] = append(byTask[name], p.Name)
		}
	}
	return byTask, missing
}

func runSteps(ctx context.Context, root string, steps []protocol.Step, remembered *verdicts, out io.Writer) ([]stepOutcome, error) {
	var failures []string
	var outcomes []stepOutcome
	for _, step := range steps {
		var key string
		var inputs taskcache.KeyInputs
		if remembered != nil {
			var ok bool
			if key, inputs, ok = remembered.keyOf(step); ok {
				if since, hit := remembered.passed(key); hit {
					replay(out, step, since)
					outcomes = append(outcomes, stepOutcome{Name: step.Name, Verdict: stepUnchanged, Since: since.UTC().Format("2006-01-02")})
					continue
				}
			}
		}
		fmt.Fprintf(out, "▶ %s\n", step.Name)
		argv := make([]string, len(step.Command))
		for i, a := range step.Command {
			argv[i] = strings.NewReplacer("{smith}", taskrun.Self(), "{workspaceRoot}", filepath.ToSlash(root)).Replace(a)
		}
		dir := root
		if step.Dir != "" {
			dir = filepath.Join(root, filepath.FromSlash(step.Dir))
		}
		began := time.Now()
		tail := &tailWriter{}
		code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: argv, Dir: dir, Env: os.Environ(), Out: io.MultiWriter(out, tail)})
		if err == nil && code == 0 {
			if key != "" {
				if err := remembered.remember(key, step, inputs, time.Since(began)); err != nil {
					warn(out, step.Name+": the verdict was not remembered: "+err.Error())
				}
			}
			outcomes = append(outcomes, stepOutcome{Name: step.Name, Took: time.Since(began)})
			continue
		}
		msg := fmt.Sprintf("%s (exit %d)", step.Name, code)
		if err != nil {
			msg = step.Name + ": " + err.Error()
		}
		if step.Blocking {
			return outcomes, errors.New(failureWithOutput(out, msg, tail.String()))
		}
		failures = append(failures, msg)
		outcomes = append(outcomes, stepOutcome{Name: step.Name, Verdict: stepProblem, Took: time.Since(began)})
		warn(out, msg+" — not blocking")
	}
	if len(failures) > 0 {
		fmt.Fprintf(out, "%d non-blocking step(s) reported problems\n", len(failures))
	}
	return outcomes, nil
}

// failureWithOutput is the failure of a blocking step with the end of its
// output after it. When that output names a log file smith can read (npm's
// debug log stays on the agent), the file's end is printed first, in a CI log
// group, so the cause can be found from the pipeline.
func failureWithOutput(out io.Writer, msg, output string) string {
	if path := logFileNamed(output); path != "" {
		if raw, err := os.ReadFile(path); err == nil {
			group(out, "log file "+path)
			fmt.Fprintf(out, "last %d lines of %s\n%s\n", logFileLines, path, lastLines(string(raw), logFileLines))
			endGroup(out)
		}
	}
	if strings.TrimSpace(output) == "" {
		return msg
	}
	return msg + "\n" + lastLines(output, outputTailLines)
}

// ensureIdentity gives CI a committer identity for annotated release tags.
func ensureIdentity(s *session.Session) {
	if v, _ := s.Git.Run("config", "user.email"); v == "" {
		_, _ = s.Git.Run("config", "user.email", "smith-release@users.noreply.github.com")
	}
	if v, _ := s.Git.Run("config", "user.name"); v == "" {
		_, _ = s.Git.Run("config", "user.name", "smith release")
	}
}

func group(out io.Writer, name string) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintf(out, "::group::%s\n", name)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##[group]%s\n", name)
	}
}

func endGroup(out io.Writer) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintln(out, "::endgroup::")
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintln(out, "##[endgroup]")
	}
}

func annotate(out io.Writer, phase string, err error) {
	msg := strings.ReplaceAll(err.Error(), "\n", " ")
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintf(out, "::error title=smith ci %s::%s\n", phase, msg)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##vso[task.logissue type=error]smith ci %s: %s\n", phase, msg)
	}
}

// notice puts a result in the run summary of the CI that has notices.
func notice(out io.Writer, msg string) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		fmt.Fprintf(out, "::notice title=smith ci release::%s\n", msg)
	}
}

func warn(out io.Writer, msg string) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintf(out, "::warning::%s\n", msg)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##vso[task.logissue type=warning]%s\n", msg)
	default:
		fmt.Fprintln(out, "warning: "+msg)
	}
}
