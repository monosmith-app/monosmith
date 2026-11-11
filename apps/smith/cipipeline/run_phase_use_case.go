package cipipeline

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/affected"
	"github.com/monosmith-app/monosmith/apps/smith/nativebuild"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/releaseassets"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// VerifyTasks are what a pull request must pass.
var VerifyTasks = []string{"lint", "typecheck", "test", "build"}

// AssetsRun is what the assets phase needs to run again for an existing tag.
type AssetsRun struct {
	// Tag is the release tag (project@version); empty means the tags at HEAD.
	Tag string
	// Source is current or tag: the code to build when Tag is not at HEAD.
	Source string
	// Terminal says a person can answer the question for the source on In.
	Terminal bool
	In       io.Reader
	// Manual is set when a person started the release by hand (the release phase).
	Manual ManualRelease
	// NoCache runs every task of the verify phase, even when a cached result
	// exists: a clean CI agent has no cache.
	NoCache bool
}

// PackTask is the task that builds what a project deploys into dist/drop.
const PackTask = "package"

// DropDir is where packed deliverables go, workspace-relative.
const DropDir = "dist/drop"

// Phases lists `smith ci` phases in pipeline order.
var Phases = []string{"setup", "audit", "verify", "pack", "native", "assets", "release"}

// RunPhase runs one CI phase.
func RunPhase(ctx context.Context, s *session.Session, phase string, out io.Writer) error {
	return RunPhaseFor(ctx, s, phase, false, AssetsRun{}, out)
}

// AssetsPrepare returns the step that installs the dependencies in the
// worktree of a tag before its release files are built: the same install
// `smith ci assets` runs, so `smith release assets --source tag` builds from
// the same state (ADR 0075).
//
// Parameters: out receives the install output.
// Returns: a function for releaseassets.Request.Prepare.
func AssetsPrepare(out io.Writer) func(ctx context.Context, ws *session.Session) error {
	return func(ctx context.Context, ws *session.Session) error {
		return RunPhaseFor(ctx, ws, "setup", true, AssetsRun{}, out)
	}
}

// RunPhaseFor runs one CI phase. With assets set, the setup phase installs
// only what the assets job needs: it leaves out the projects whose [release]
// install is false (smith ci setup --assets). A tag (project@version) makes the
// assets phase build for that existing release instead of the ones at HEAD.
func RunPhaseFor(ctx context.Context, s *session.Session, phase string, assets bool, run AssetsRun, out io.Writer) error {
	begin := time.Now()
	log := newPhaseLog(out, phase)
	if !announces(phase) {
		log.fold()
	}
	err := runPhase(ctx, s, phase, assets, run, log)
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

func runPhase(ctx context.Context, s *session.Session, phase string, assets bool, run AssetsRun, log *phaseLog) error {
	out := log.out
	switch phase {
	case "setup", "audit":
		collect := Collect
		if assets && phase == "setup" {
			collect = CollectForAssets
		}
		c, err := collect(s)
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
		var outcomes []stepOutcome
		var sections *SectionLog
		if phase == "setup" {
			// Each plugin's steps are a section (a CI group of their own), not one fold.
			log.heading(intro...)
			sections = &SectionLog{out: out, icons: log.icons, colour: colourOn(), now: time.Now, Remedy: defaultSectionRemedy}
			outcomes, err = runSetupSections(ctx, s.Root, c, sections, out)
		} else {
			log.banner(intro...)
			outcomes, err = runSteps(ctx, s.Root, steps, remembered, out)
		}
		log.close()
		if sections != nil {
			sections.Summary()
		}
		if len(outcomes) > 0 {
			table := stepTable(map[string]string{"setup": "Step", "audit": "Check"}[phase], outcomes, log.icons)
			if sections == nil {
				for _, line := range table.Text("steps done") {
					log.say("%s", line)
				}
			}
			log.markdown = table.Markdown()
		}
		if problems := problemNames(outcomes); len(problems) > 0 {
			summary := fmt.Sprintf("%s reported problems (not blocking): %s", plural(len(problems), "step"), listSome(problems, 3))
			if sections != nil {
				warn(out, summary)
			} else {
				// runSteps raised a warning for each of these steps already:
				// a second one for the same finding is noise in the run summary.
				log.say("%s", summary)
			}
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
		if err := verifyDerived(s, out); err != nil {
			return err
		}
		if err := verifyBoundaries(s, out); err != nil {
			return err
		}
		// Native projects are verified by the native job, on every OS.
		outputs := newTaskOutputs()
		tee := io.MultiWriter(out, outputs)
		summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: VerifyTasks, Projects: nativeExcluded, Affected: base != "", Base: base, NoCache: run.NoCache, Release: true, Out: tee, Colour: taskrun.ToolColour(os.Stdout)})
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
		summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: []string{PackTask}, Projects: nativeExcluded, Release: true, Out: io.MultiWriter(out, outputs), Colour: taskrun.ToolColour(os.Stdout)})
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
		res, err := releaseassets.Run(ctx, s, releaseassets.Request{
			Platform: os.Getenv("SMITH_PLATFORM"), Tag: run.Tag, Source: run.Source, Terminal: run.Terminal, In: run.In, Out: out,
			// A tag's own commit is a fresh checkout without the installed dependencies.
			Prepare: AssetsPrepare(out),
		})
		if err == nil && res.Skipped != "" {
			// The job summary says it too, so a run that did nothing does not read as a success.
			log.conclusion = res.Skipped
		}
		return err
	case "release":
		ensureIdentity(s)
		plan, err := release.PlanFor(s)
		if err != nil {
			return err
		}
		manual := ""
		if run.Manual.Active() {
			var notes []string
			if plan, notes, err = release.AdjustPlan(s, plan, nil, run.Manual.Bumps, run.Manual.Force); err != nil {
				return err
			}
			for _, n := range notes {
				fmt.Fprintln(out, "note: "+n)
			}
			manual = run.Manual.Sentence(Actor(os.Getenv))
		}
		if err := release.CheckAPI(ctx, s, &plan); err != nil {
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
			log.heading(releaseIntro(plan, release.PlanBranch(s), projects)...)
			log.fold()
			for _, item := range plan.Items {
				fmt.Fprintf(out, "%s: %s -> %s (%s) because %s\n", item.Project, cmp.Or(item.Current, "new"), item.Next, bumpOf(item), item.Why())
				for _, w := range item.APIWarnings {
					warn(out, w)
				}
			}
		}
		report, err := release.Execute(ctx, s, plan, release.Options{Push: true, Out: out})
		// A release whose dist-tag waits for npm is done: the steps after it
		// still run, and the run ends with a warning and exit code 2.
		pending := errors.Is(err, release.ErrPropagationPending)
		if pending {
			err = nil
		}
		markReleased(report)
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
		if manual != "" {
			log.say("%s", manual)
			log.markdown = append([]string{"**" + manual + "**", ""}, log.markdown...)
		}
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
			for _, line := range releasedSummary(report, log.icons, colourOn()) {
				log.say("%s", line)
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
		if err == nil && pending {
			for _, command := range report.PendingTags() {
				warn(out, "a dist-tag was not moved: npm did not serve every platform package in time. When it does, run: "+command)
				log.say("%s", "→ when npm serves the packages, run: "+command)
			}
			return release.ErrPropagationPending
		}
		return err
	}
	return fmt.Errorf("unknown phase %q (phases: %s)", phase, strings.Join(Phases, ", "))
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
			return outcomes, errors.New(failureWithOutput(out, step.Name, msg, tail.String()))
		}
		if code == 127 {
			// The shell could not find the tool: nothing was checked.
			outcomes = append(outcomes, stepOutcome{Name: step.Name, Verdict: stepSkipped, Took: time.Since(began)})
			warn(out, step.Name+" was skipped: a tool it needs is not installed, so nothing was checked. Install it, or make the audit blocking for this ecosystem in [ci] audit.")
			continue
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
// output after it. The step's whole captured output (up to tailKeep bytes) is
// printed again in a CI log group named after the step, so a cause the live
// stream scrolled past (pip's conflict block) is one click away. When the
// output names a log file smith can read (npm's debug log stays on the agent),
// the file's end is printed first in its own group.
func failureWithOutput(out io.Writer, step, msg, output string) string {
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
	group(out, step+": full output")
	if !inGroupedLog() {
		fmt.Fprintf(out, "%s: full output\n", step)
	}
	fmt.Fprintf(out, "%s\n", strings.TrimRight(output, "\r\n"))
	endGroup(out)
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
