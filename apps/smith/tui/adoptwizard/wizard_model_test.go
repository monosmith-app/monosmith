package adoptwizard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func plain() theme.Theme { return theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}) }

// fakeFlow is smith, for the wizard: it records what it was asked and answers
// from the fields.
type fakeFlow struct {
	found    adoptflow.Found
	groups   []removalplan.Group
	calls    []string
	failOn   map[string]error
	finish   []protocol.FixStep
	closed   bool
	applied  map[string]bool
	declined []map[string]bool
}

func (f *fakeFlow) note(s string) { f.calls = append(f.calls, s) }

func (f *fakeFlow) fail(name string) error {
	if err := f.failOn[name]; err != nil {
		delete(f.failOn, name)
		return err
	}
	return nil
}

func (f *fakeFlow) Found() adoptflow.Found { return f.found }

func (f *fakeFlow) Prepare() (convertchoices.Choices, []convertchoices.Question, error) {
	f.note("prepare")
	def := convertchoices.Choices{Level: convertchoices.LevelFull, Lint: convertchoices.Smith, Test: convertchoices.Jest, Build: convertchoices.Smith, CI: convertchoices.Smith}
	qs := []convertchoices.Question{
		{Key: "lint", Question: "Lint with:", Options: []convertchoices.Option{{Value: "smith", Label: "smith's ESLint"}, {Value: "keep", Label: "keep"}}},
		{Key: "test", Question: "Run tests with:", Options: []convertchoices.Option{{Value: "jest", Label: "Jest"}, {Value: "vitest", Label: "Vitest"}, {Value: "keep", Label: "keep"}}},
		{Key: "build", Question: "Build with:", Options: []convertchoices.Option{{Value: "smith", Label: "smith's build"}, {Value: "keep", Label: "keep"}}},
		{Key: "ci", Question: "CI pipeline:", Options: []convertchoices.Option{{Value: "smith", Label: "smith's pipeline"}, {Value: "keep", Label: "keep"}}},
	}
	return def, qs, f.fail("prepare")
}

func (f *fakeFlow) PlanFull(c convertchoices.Choices, verify bool) (adoptflow.Preview, error) {
	f.note(fmt.Sprintf("plan %s verify=%v", convertchoices.Flags(c), verify))
	return adoptflow.Preview{Text: "● Full conversion: smith becomes the only tool\n● Remove 2 lint packages\n--- package.json\n+++ package.json\n@@ -1 +1 @@\n-eslint\n+smith\n"}, f.fail("plan")
}

func (f *fakeFlow) Convert(_ context.Context, c convertchoices.Choices, out io.Writer) (string, error) {
	f.note("convert " + string(c.Level))
	fmt.Fprintln(out, "… "+adoptflow.ConvertStepName)
	return "Converted 3 projects.", f.fail("convert")
}

func (f *fakeFlow) NextGroup(c convertchoices.Choices, declined map[string]bool) (removalplan.Plan, error) {
	copied := map[string]bool{}
	for k, v := range declined {
		copied[k] = v
	}
	f.declined = append(f.declined, copied)
	f.note("next")
	var plan removalplan.Plan
	for _, g := range f.groups {
		if !declined[g.ID] && !f.applied[g.ID] {
			plan.Groups = append(plan.Groups, g)
		}
	}
	plan.Kept = []legacyinventory.Kept{{What: "rollup.config.js", Why: "build.sh runs rollup"}}
	return plan, nil
}

func (f *fakeFlow) ShowFix(p doctorfix.Planned) (string, error) {
	var b strings.Builder
	b.WriteString("● " + p.Description + "\n")
	for _, e := range p.Files {
		b.WriteString("--- " + e.Path + "\n+++ /dev/null\n")
	}
	for _, s := range p.Steps {
		b.WriteString("  $ " + strings.Join(s.Argv, " ") + "\n")
	}
	return b.String(), nil
}

func (f *fakeFlow) ApplyFix(_ context.Context, fix protocol.Fix, out io.Writer) error {
	f.note("apply " + fix.Description)
	for _, e := range fix.Files {
		if f.applied == nil {
			f.applied = map[string]bool{}
		}
		f.applied[strings.TrimSuffix(e.Path, ".config.js")] = true
	}
	for _, s := range fix.Steps {
		fmt.Fprintln(out, "… "+s.Description)
		if err := f.failOn[s.Description]; err != nil {
			delete(f.failOn, s.Description)
			return &doctorfix.StepError{Number: 1, Total: len(fix.Steps), Step: s, Err: err, Output: "npm ERR! it broke\nline two"}
		}
		fmt.Fprintln(out, "✔ "+s.Description)
	}
	return f.fail("apply " + fix.Description)
}

func (f *fakeFlow) FinishFix(withLint bool) doctorfix.Planned {
	f.note(fmt.Sprintf("finish lint=%v", withLint))
	steps := f.finish
	if steps == nil {
		steps = []protocol.FixStep{{Description: "install", Argv: []string{"npm", "install"}}, {Description: "build every project", Argv: []string{"smith", "run", "build"}}}
	}
	return doctorfix.Planned{Fix: protocol.Fix{Description: "Install, then check", Steps: steps}}
}

func (f *fakeFlow) AdoptFix() doctorfix.Planned {
	return doctorfix.Planned{Fix: protocol.Fix{Description: "Make this repository a smith workspace", Steps: []protocol.FixStep{{Description: "adopt", Argv: []string{"smith", "adopt", "--init"}}}}}
}

func (f *fakeFlow) Close() { f.closed = true }

func (f *fakeFlow) Changes() (adoptflow.Changes, error) {
	f.note("changes")
	return adoptflow.Changes{Created: []string{"smith.toml"}, Edited: []string{"apps/api/package.json", "libs/ui/package.json"}, Removed: []string{"nx.json"}}, f.fail("changes")
}

func (f *fakeFlow) ShowChanges(adoptflow.Changes) (string, error) {
	f.note("show changes")
	return `● The files smith writes: 1 created, 2 edited, 1 removed
--- a/nx.json
+++ b/nx.json
@@ -1 +0,0 @@
-{}
+++ b/smith.toml
+[workspace]
`, nil
}

func (f *fakeFlow) Result() convert.Highlights {
	return convert.Highlights{Projects: "2 projects (1 app, 1 library)", Changed: []string{"build: Rollup becomes esbuild (2 projects)"}, Checks: []string{"azure-pipelines.yml: replaced by a generated pipeline"}}
}

func (f *fakeFlow) Report() (string, error) {
	f.note("report")
	return "Conversion report for acme\n\n• build: Rollup becomes esbuild (2 projects)\n", f.fail("report")
}

func (f *fakeFlow) Install(_ context.Context, out io.Writer) error {
	f.note("install")
	f.found.NeedsInstall, f.found.Blockers = false, nil
	fmt.Fprintln(out, "… install the packages (npm ci)")
	return f.fail("install")
}

func group(id string, files, pkgs int, reason string) removalplan.Group {
	return removalplan.Group{ID: id, Reason: reason, Files: files, Packages: pkgs, Removes: []string{"eslint-plugin-x"},
		Fix: doctorfix.Planned{Fix: protocol.Fix{Description: reason, Files: []protocol.FileEdit{{Path: id + ".config.js", Delete: true}}}, Ask: true}}
}

func nxFound() adoptflow.Found {
	return adoptflow.Found{
		Kind: adoptflow.KindNx, Source: "Nx workspace", Name: "acme", Full: true,
		Projects: []adoptflow.Project{{Name: "api", Root: "apps/api", Language: "node", Kind: "app"}, {Name: "ui", Root: "libs/ui", Language: "node", Kind: "lib"}},
		Tooling:  []string{"Nx", "mnci", "ESLint", "Jest", "rollup"},
	}
}

// walk is the shell: it runs the jobs an outcome asks for, one after the
// other, and reports each end back.
type walk struct {
	t    *testing.T
	m    Model
	last Outcome
	ran  []string
	// progress is what the jobs reported.
	progress []string
}

func start(t *testing.T, f *fakeFlow) *walk {
	t.Helper()
	m, job := New(func(io.Writer) (Flow, error) { return f, nil })
	w := &walk{t: t, m: m.Place(100, 34)}
	w.runJob(&job)
	return w
}

func (w *walk) runJob(job *Job) {
	for job != nil {
		w.ran = append(w.ran, job.Name)
		var out bytes.Buffer
		res, err := job.Run(context.Background(), &out)
		n := 0
		for _, line := range strings.Split(out.String(), "\n") {
			if name, ok := strings.CutPrefix(line, "… "); ok {
				n++
				w.m = w.m.Progress(n, name)
				w.progress = append(w.progress, name)
			}
		}
		var o Outcome
		w.m, o = w.m.Done(res, err)
		w.last = o
		job = o.Job
	}
}

func (w *walk) press(keys ...string) *walk {
	for _, k := range keys {
		var o Outcome
		w.m, o = w.m.Key(k)
		w.last = o
		w.runJob(o.Job)
	}
	return w
}

func (w *walk) screen() string {
	return pane.Strip(strings.Join(w.m.Place(100, 34).Lines(plain(), 0), "\n"))
}

func (w *walk) want(sub ...string) {
	w.t.Helper()
	got := w.screen()
	for _, s := range sub {
		if !strings.Contains(got, s) {
			w.t.Fatalf("the screen lacks %q:\n%s", s, got)
		}
	}
}

func TestTheFirstScreenTellsWhatWasFoundAndNothingHasChanged(t *testing.T) {
	f := &fakeFlow{found: nxFound()}
	w := start(t, f)
	if w.m.Stage() != StageFound {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.want("Bring in · what smith found", "Found an Nx workspace: acme", "2 projects", "Tools it uses today: Nx, mnci, ESLint, Jest, rollup", "api  TypeScript / JavaScript app", "Nothing has changed", "[ Continue  Enter ]")
}

func TestABlockedRepositoryExplainsAndClosesOnEnter(t *testing.T) {
	f := &fakeFlow{found: adoptflow.Found{Kind: adoptflow.KindNotARepo, Name: "x", Source: "repository", Blockers: []string{"This folder is not a git repository"}}}
	w := start(t, f)
	w.want("✖ This folder is not a git repository", "Nothing was changed", "[ Close  Enter ]")
	w.press("enter")
	if !w.last.Close || w.last.Job != nil {
		t.Fatalf("outcome %+v", w.last)
	}
}

func TestACancelBeforeAnythingChangesClosesWithoutANote(t *testing.T) {
	w := start(t, &fakeFlow{found: nxFound()})
	w.press("esc")
	if !w.last.Close || w.last.Note != "" {
		t.Fatalf("outcome %+v", w.last)
	}
}

func TestTheLevelQuestionOffersConservativeFirstAndFull(t *testing.T) {
	w := start(t, &fakeFlow{found: nxFound()}).press("enter")
	w.want("Bring in · how far to go", "How far should smith go?", "1) Conservative", "(suggested)", "2) Full")
	w.press("b")
	if w.m.Stage() != StageFound {
		t.Fatalf("Back from the first question: stage %v", w.m.Stage())
	}
}

func TestTheFullWalkAsksTheChoicesPlansReviewsConvertsAsksEachGroupAndChecks(t *testing.T) {
	f := &fakeFlow{found: nxFound(), groups: []removalplan.Group{
		group("lint", 3, 4, "Use smith's ESLint config: you chose smith's ESLint"),
		group("nx", 2, 5, "Remove Nx: nothing runs Nx any more"),
	}}
	w := start(t, f).press("enter", "2") // Continue, then Full
	if w.m.Stage() != StageChoices {
		t.Fatalf("stage %v after the level", w.m.Stage())
	}
	w.want("your choices 1 of 4", "Lint with:", "1) smith's ESLint  (suggested)")
	w.press("enter", "2", "enter", "enter") // lint smith, test vitest, build smith (default), ci smith (default)
	if w.m.Stage() != StageReview {
		t.Fatalf("stage %v after the choices", w.m.Stage())
	}
	if got := strings.Join(f.calls, "|"); !strings.Contains(got, "plan --level full --lint smith --test vitest --build smith --ci smith verify=true") {
		t.Fatalf("the choices were not planned as answered: %s", got)
	}
	w.want("Bring in · what will change", "Nothing has changed yet", "! Check these before you go ahead (1)", "Behaves differently afterwards (1)", "Remove 2 lint packages", "[ Convert  Enter ]", "[ Back  b ]")

	w.press("enter") // convert, then the first group is asked
	if w.m.Stage() != StageGroup {
		t.Fatalf("stage %v after converting", w.m.Stage())
	}
	w.want("We are going to delete 4 packages and 3 files", "you chose smith's ESLint", "ok?", "[ Yes, delete  Enter ]", "[ No, keep  n ]", "[ Skip rest  s ]")
	if len(w.progress) == 0 || w.progress[0] != adoptflow.ConvertStepName {
		t.Fatalf("the conversion reported %v", w.progress)
	}

	w.press("y") // delete lint, then the nx group
	w.want("We are going to delete 5 packages and 2 files", "Remove Nx")
	w.press("n") // keep nx; nothing is left, so the checks run
	if w.m.Stage() != StageDone {
		t.Fatalf("stage %v after the groups", w.m.Stage())
	}
	w.want("is a smith workspace now", "2 projects (1 app, 1 library)", "Needs you (1)", "azure-pipelines.yml", "Behaves differently now (1)", "Old tools removed: lint", "Left as they were, as you chose: nx", "rollup.config.js: build.sh runs rollup", "It works: everything builds (", "1 thing needs you (listed below).", "[ Open the workspace  Enter ]", "[ Read the report  s ]")
	if strings.Contains(w.screen(), "Open it and build everything") {
		t.Fatalf("the checks already built it: no need to offer a build:\n%s", w.screen())
	}

	// The report opens inside the wizard, and Enter comes back to the result.
	w.press("s")
	w.want("Bring in · the report", "Conversion report for acme", "Rollup becomes esbuild", "[ Back to the result  Enter ]")
	if w.m.Stage() != StageReport {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.press("enter")
	if w.m.Stage() != StageDone {
		t.Fatalf("stage %v after the report", w.m.Stage())
	}
	w.want("It works: everything builds (")

	// The declined group stayed out of the next plan, and the checks ran with verify.
	last := f.declined[len(f.declined)-1]
	if !last["nx"] || last["lint"] {
		t.Fatalf("the last plan was made with declined=%v", last)
	}
	if got := strings.Join(f.calls, "|"); !strings.Contains(got, "finish lint=true") {
		t.Fatalf("calls: %s", got)
	}
	w.press("enter")
	if !w.last.Close || !strings.Contains(w.last.Finished, "is ready") || w.last.Build {
		t.Fatalf("outcome %+v", w.last)
	}
}

func TestOpenItAndBuildEverythingAsksTheShellToBuild(t *testing.T) {
	f := &fakeFlow{found: nxFound(), finish: []protocol.FixStep{{Description: "install", Argv: []string{"npm", "install"}}}} // no check ran
	w := start(t, f).press("enter", "1", "enter")
	w.press("n")
	if !w.last.Close || w.last.Finished == "" || !w.last.Build {
		t.Fatalf("outcome %+v", w.last)
	}
}

func TestSkipKeepsThisGroupAndTheRestAndGoesToTheChecks(t *testing.T) {
	f := &fakeFlow{found: nxFound(), groups: []removalplan.Group{group("lint", 1, 1, "a"), group("tests", 1, 1, "b"), group("nx", 1, 1, "c")}}
	w := start(t, f).press("enter", "2", "enter", "enter", "enter", "enter", "enter") // through the choices, the review, convert
	if w.m.Stage() != StageGroup {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.press("s")
	if w.m.Stage() != StageDone {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.want("Left as they were, as you chose: lint, tests, nx")
}

func TestTheConservativeLevelOfAnNxWorkspaceConvertsThenInstallsWithoutTheChecks(t *testing.T) {
	f := &fakeFlow{found: nxFound()}
	w := start(t, f).press("enter", "1")
	w.want("Convert the Nx workspace \"acme\" to smith, keeping what works", "Files: smith creates 1, edits 2 and removes 1.", "Things to check: 1", "! Check these before you go ahead (1)", "azure-pipelines.yml", "Behaves differently afterwards (1)", "Rollup becomes esbuild", "[ Go ahead  Enter ]")
	if strings.Contains(w.screen(), "stay as they are") || strings.Contains(w.screen(), "with the same tools") {
		t.Fatalf("the review must not promise that the old tooling stays: it removes Nx:\n%s", w.screen())
	}
	w.press("enter")
	if w.m.Stage() != StageDone {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.want("It works: everything builds (", "[ Read the report  s ]")
	got := strings.Join(f.calls, "|")
	if !strings.Contains(got, "convert conservative") || !strings.Contains(got, "finish lint=false") || !strings.Contains(got, "changes") || !strings.Contains(got, "show changes") {
		t.Fatalf("calls: %s", got)
	}
}

func TestAPlainRepositoryAdoptsConservativelyAndSaysWhyThereIsNoFullLevel(t *testing.T) {
	f := &fakeFlow{found: adoptflow.Found{
		Kind: adoptflow.KindPlain, Source: "repository", Name: "solo", NoFull: adoptflow.NoFullForPlain,
		Projects: []adoptflow.Project{{Name: "solo", Root: ".", Language: "go", Kind: "app"}},
	}}
	w := start(t, f)
	w.want("Found a repository: solo", "The full level", "Tracked in")
	w.press("enter")
	if w.m.Stage() != StageReview {
		t.Fatalf("a plain repository goes straight to the review, not a level question: stage %v", w.m.Stage())
	}
	w.want("smith adopt --init", "your build and lint setup stay as they are", "[ Go ahead  Enter ]")
	w.press("enter")
	if w.m.Stage() != StageDone {
		t.Fatalf("stage %v", w.m.Stage())
	}
	if !strings.Contains(strings.Join(f.calls, "|"), "apply Make this repository a smith workspace") {
		t.Fatalf("calls: %v", f.calls)
	}
}

func TestAFailedStepShowsItsOutputAndRetryRunsTheFinishAgain(t *testing.T) {
	f := &fakeFlow{found: nxFound(), failOn: map[string]error{"build every project": errors.New("exited with code 1")}}
	w := start(t, f).press("enter", "1", "enter")
	w.want("Bring in failed", "build every project", "npm ERR! it broke", "[ Retry  r ]", "[ Close  Esc ]")
	w.press("r")
	if w.m.Stage() != StageDone {
		t.Fatalf("stage %v after the retry", w.m.Stage())
	}
	installs := 0
	for _, c := range f.calls {
		if strings.HasPrefix(c, "apply Install") {
			installs++
		}
	}
	if installs != 2 {
		t.Fatalf("the finish ran %d times, want 2: %v", installs, f.calls)
	}
}

func TestAFailedConversionCannotBeRetriedBlindlyAndClosesWithTheWayBack(t *testing.T) {
	f := &fakeFlow{found: nxFound(), failOn: map[string]error{"convert": errors.New("boom")}}
	w := start(t, f).press("enter", "1", "enter")
	got := w.screen()
	if strings.Contains(got, "Retry") || !strings.Contains(got, "boom") || !strings.Contains(got, "Close") {
		t.Fatalf("the failure of the conversion:\n%s", got)
	}
	w.press("r")
	if w.last.Job != nil {
		t.Fatal("r reran the conversion")
	}
	w.press("esc")
	if !w.last.Close || !strings.Contains(w.last.Note, "git") {
		t.Fatalf("outcome %+v", w.last)
	}
}

func TestAnErrorReadingTheRepositoryOffersRetry(t *testing.T) {
	calls := 0
	m, job := New(func(io.Writer) (Flow, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("git is not installed")
		}
		return &fakeFlow{found: nxFound()}, nil
	})
	w := &walk{t: t, m: m.Place(100, 34)}
	w.runJob(&job)
	w.want("git is not installed", "[ Retry  r ]")
	w.press("r")
	if w.m.Stage() != StageFound {
		t.Fatalf("stage %v", w.m.Stage())
	}
}

func TestBackFromTheReviewReturnsToTheChoicesAndFromThereToTheLevel(t *testing.T) {
	w := start(t, &fakeFlow{found: nxFound()}).press("enter", "2", "enter", "enter", "enter", "enter")
	if w.m.Stage() != StageReview {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.press("b")
	if w.m.Stage() != StageChoices {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.press("b") // Back at the first question goes to the level
	if w.m.Stage() != StageLevel {
		t.Fatalf("stage %v", w.m.Stage())
	}
}

func TestEscAfterTheConversionStopsWithTheWayBack(t *testing.T) {
	f := &fakeFlow{found: nxFound(), groups: []removalplan.Group{group("lint", 1, 1, "a")}}
	w := start(t, f).press("enter", "2", "enter", "enter", "enter", "enter", "enter")
	if w.m.Stage() != StageGroup {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.press("esc")
	if !w.last.Close || !strings.Contains(w.last.Note, "git restore") {
		t.Fatalf("outcome %+v", w.last)
	}
}

func TestWhileAJobRunsEscAndCtrlCStopItAndNothingElseDoes(t *testing.T) {
	m, _ := New(func(io.Writer) (Flow, error) { return nil, nil })
	m = m.Place(100, 34)
	if !m.Running() {
		t.Fatal("not running while reading")
	}
	for _, k := range []string{"esc", "ctrl+c"} {
		if _, o := m.Key(k); !o.Stop {
			t.Fatalf("%s did not stop the job", k)
		}
	}
	if _, o := m.Key("enter"); o.Stop || o.Close || o.Job != nil {
		t.Fatalf("enter during a job: %+v", o)
	}
}

func TestProgressNamesTheStepOfTheFinishAndTheDialogFitsASmallScreen(t *testing.T) {
	f := &fakeFlow{found: nxFound()}
	m, job := New(func(io.Writer) (Flow, error) { return f, nil })
	w := &walk{t: t, m: m.Place(40, 12)}
	w.runJob(&job)
	w.press("enter", "1", "enter")
	for _, l := range w.m.Place(40, 12).Lines(plain(), 0) {
		if pane.Width(l) > 40 {
			t.Fatalf("line wider than the screen: %q", pane.Strip(l))
		}
	}
	// Mid-run: the second of two steps.
	w.m = w.m.running(StageFinishing, "Install, then check", Job{Steps: 2}, adoptStyle).Progress(2, "build every project")
	if got := pane.Strip(strings.Join(w.m.Place(100, 20).Lines(plain(), 0), "\n")); !strings.Contains(got, "step 2 of 2: build every project") {
		t.Fatalf("progress:\n%s", got)
	}
}

func TestDisposeClosesTheFlow(t *testing.T) {
	f := &fakeFlow{found: nxFound()}
	w := start(t, f)
	w.m.Dispose()
	if !f.closed {
		t.Fatal("the flow was not closed")
	}
}

func TestMissingPackagesAreOneClickToInstallThenTheWizardGoesOn(t *testing.T) {
	found := nxFound()
	found.Projects = nil
	found.NeedsInstall = true
	found.Blockers = []string{"The packages are not installed, and smith needs them to read your projects."}
	f := &fakeFlow{found: found}
	w := start(t, f)
	w.want("● The packages are not installed", "None of your own files are touched", "[ Install the packages and continue  Enter ]")
	if strings.Contains(w.screen(), "✖") {
		t.Fatalf("a step smith takes itself is not shown as a failure: %s", w.screen())
	}
	if strings.Contains(w.screen(), "No projects found yet") {
		t.Fatalf("the failure that caused it should not be shown as an empty result: %s", w.screen())
	}
	w.press("enter")
	if got := strings.Join(f.calls, "|"); !strings.Contains(got, "install") {
		t.Fatalf("calls: %s", got)
	}
	if w.m.Stage() != StageFound {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.want("Found an Nx workspace", "[ Continue  Enter ]")
}

func TestEachScreenIsTitledByWhatItShowsNotByTheKindOfForm(t *testing.T) {
	w := start(t, &fakeFlow{found: nxFound()})
	w.want("Bring in · what smith found")
	w.press("enter")
	w.want("Bring in · how far to go")
	if strings.Contains(w.screen(), "question 1 of 1") {
		t.Fatal("a generated form title: question 1 of 1")
	}
}

func TestReadingTheRepositoryShowsWhatTheOpenerSaysItIsReading(t *testing.T) {
	m, job := New(func(out io.Writer) (Flow, error) {
		fmt.Fprintln(out, "… reading 1,240 files")
		return &fakeFlow{found: nxFound()}, nil
	})
	var out bytes.Buffer
	if _, err := job.Run(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	name, _ := strings.CutPrefix(strings.TrimSpace(out.String()), "… ")
	m = m.Place(100, 34).Progress(1, name)
	got := pane.Strip(strings.Join(m.Lines(plain(), 0), "\n"))
	if !strings.Contains(got, "Reading 1,240 files…") {
		t.Fatalf("the reading stage:\n%s", got)
	}
}
