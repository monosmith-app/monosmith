package tui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
	"github.com/monosmith-app/monosmith/apps/smith/tui/adoptwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// adoptFake is smith, for the shell: a repository that is an Nx workspace,
// one group of removals, and the calls it was given.
type adoptFake struct {
	mu       sync.Mutex
	calls    []string
	hold     chan struct{}
	stopped  bool
	applied  bool
	closed   bool
	found    adoptflow.Found
	finishes int
}

func (f *adoptFake) note(s string) {
	f.mu.Lock()
	f.calls = append(f.calls, s)
	f.mu.Unlock()
}

func (f *adoptFake) Calls() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return strings.Join(f.calls, "|")
}

func (f *adoptFake) Found() adoptflow.Found { return f.found }

func (f *adoptFake) Prepare() (convertchoices.Choices, []convertchoices.Question, error) {
	f.note("prepare")
	def := convertchoices.Choices{Level: convertchoices.LevelFull, Lint: convertchoices.Smith, Test: convertchoices.Jest, Build: convertchoices.Smith, CI: convertchoices.Smith}
	return def, []convertchoices.Question{
		{Key: "lint", Question: "Lint with:", Options: []convertchoices.Option{{Value: "smith", Label: "smith's ESLint"}, {Value: "keep", Label: "keep"}}},
	}, nil
}

func (f *adoptFake) PlanFull(c convertchoices.Choices, verify bool) (adoptflow.Preview, error) {
	f.note("plan " + convertchoices.Flags(c))
	return adoptflow.Preview{Text: "● Full conversion: smith becomes the only tool\n● Remove 1 lint package\n"}, nil
}

func (f *adoptFake) Convert(ctx context.Context, c convertchoices.Choices, out io.Writer) (string, error) {
	f.note("convert " + string(c.Level))
	fmt.Fprintln(out, "… "+adoptflow.ConvertStepName)
	if f.hold != nil {
		select {
		case <-f.hold:
		case <-ctx.Done():
			f.mu.Lock()
			f.stopped = true
			f.mu.Unlock()
			return "", ctx.Err()
		}
	}
	return "Converted 2 projects.", nil
}

func (f *adoptFake) NextGroup(c convertchoices.Choices, declined map[string]bool) (removalplan.Plan, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.applied || declined["lint"] {
		return removalplan.Plan{}, nil
	}
	g := removalplan.Group{ID: "lint", Reason: "Use smith's ESLint config: you chose smith's ESLint", Files: 3, Packages: 4,
		Fix: doctorfix.Planned{Fix: protocol.Fix{Description: "Use smith's ESLint config", Files: []protocol.FileEdit{{Path: ".eslintrc", Delete: true}}}, Ask: true}}
	return removalplan.Plan{Groups: []removalplan.Group{g}}, nil
}

func (f *adoptFake) ShowFix(p doctorfix.Planned) (string, error) {
	return "● " + p.Description + "\n", nil
}

func (f *adoptFake) ApplyFix(_ context.Context, fix protocol.Fix, out io.Writer) error {
	f.note("apply " + fix.Description)
	f.mu.Lock()
	if len(fix.Files) > 0 {
		f.applied = true
	}
	f.mu.Unlock()
	for _, s := range fix.Steps {
		fmt.Fprintln(out, "… "+s.Description)
	}
	return nil
}

func (f *adoptFake) FinishFix(verify bool) doctorfix.Planned {
	f.note(fmt.Sprintf("finish verify=%v", verify))
	return doctorfix.Planned{Fix: protocol.Fix{Description: "Install, then check", Steps: []protocol.FixStep{{Description: "install", Argv: []string{"npm", "install"}}}}}
}

func (f *adoptFake) AdoptFix() doctorfix.Planned { return doctorfix.Planned{} }

func (f *adoptFake) Close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
}

func nxRepository() adoptflow.Found {
	return adoptflow.Found{
		Kind: adoptflow.KindNx, Source: "Nx workspace", Name: "acme", Full: true,
		Projects: []adoptflow.Project{{Name: "api", Root: "apps/api", Language: "node", Kind: "app"}, {Name: "ui", Root: "libs/ui", Language: "node", Kind: "lib"}},
		Tooling:  []string{"Nx", "ESLint", "Jest"},
	}
}

// newAdoptLab is the shell in a folder that has files and is not a workspace.
func newAdoptLab(t *testing.T, f *adoptFake, empty bool, w, h int) *lab {
	t.Helper()
	l := &lab{t: t, clock: time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)}
	// As Choose makes the menu: what needs a workspace or content is left out.
	var actions []guide.Action
	for _, a := range guide.Actions([]string{"go", "node"}, []string{"go", "node"}, []string{"go", "node"}) {
		if !a.NeedsWorkspace && !(a.NeedsContent && empty) {
			actions = append(actions, a)
		}
	}
	if !empty {
		actions = guide.AdoptionFirst(actions)
	}
	cfg := Config{
		Version: "1.2.3", Workspace: "no workspace", Theme: theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}),
		Commands: smithCommands, Actions: actions, EmptyFolder: empty,
		Adopt: func() (adoptwizard.Flow, error) { return f, nil },
	}
	sh := NewShell(cfg)
	sh.now = func() time.Time { return l.clock }
	l.m = sh
	l.send(tea.WindowSizeMsg{Width: w, Height: h})
	return l
}

func TestAFolderThatIsNotAWorkspaceLeadsWithAdoptionAndEnterStartsIt(t *testing.T) {
	f := &adoptFake{found: nxRepository()}
	l := newAdoptLab(t, f, false, 110, 34)
	l.has("Not a smith workspace yet: Enter starts adoption", "Adopt this repository", "Not a smith workspace yet. Press a to adopt this folder")
	l.keys("enter")
	l.has("Adopt · review", "Found a Nx workspace: acme", "2 projects", "Old tooling in use: Nx, ESLint, Jest", "[ Continue  Enter ]")
	if l.sh().adopt == nil {
		t.Fatal("the wizard is not open")
	}
}

func TestTheAKeyAndThePaletteOpenTheSameWizard(t *testing.T) {
	f := &adoptFake{found: nxRepository()}
	l := newAdoptLab(t, f, false, 110, 34)
	l.keys("a")
	l.has("Found a Nx workspace")
	l.keys("esc")
	if l.sh().adopt != nil {
		t.Fatal("esc left the wizard open")
	}
	l.has("Cancelled: nothing was changed")
	l.keys("ctrl+p")
	l.typeText("Adopt")
	l.has("Adopt…")
	l.keys("enter")
	l.has("Found a Nx workspace")
}

func TestAnEmptyFolderHasNoAdoptionToOffer(t *testing.T) {
	f := &adoptFake{found: nxRepository()}
	l := newAdoptLab(t, f, true, 110, 34)
	l.hasNot("Enter starts adoption", "Adopt this repository")
	l.keys("a")
	if l.sh().adopt != nil {
		t.Fatal("a opened the wizard in an empty folder")
	}
	l.has("Nothing here yet")
}

func TestTheFullAdoptionRunsOffTheUIGoroutineAndLeavesWithTheNextStep(t *testing.T) {
	f := &adoptFake{found: nxRepository()}
	l := newAdoptLab(t, f, false, 120, 36)
	l.keys("a", "enter", "2")              // continue, then the full level
	l.has("question 1 of 1", "Lint with:") // the choices
	l.keys("enter")                        // smith's ESLint, the suggested one
	l.has("Nothing has changed yet", "[ Convert  Enter ]")
	l.keys("enter") // convert, then the group is asked
	l.has("We are going to delete 4 packages and 3 files", "ok?")
	l.keys("y") // delete it, then the closing steps run and the summary shows
	l.has("is a smith workspace now", "Converted 2 projects.", "Old tooling removed: lint")
	calls := f.Calls()
	for _, want := range []string{"prepare", "plan --level full --lint smith --test jest --build smith --ci smith", "convert full", "apply Use smith's ESLint config", "finish verify=true", "apply Install, then check"} {
		if !strings.Contains(calls, want) {
			t.Fatalf("calls lack %q: %s", want, calls)
		}
	}
	l.keys("enter")
	if !l.quit || !strings.Contains(l.sh().ExitSummary(), "run smith again to open it") {
		t.Fatalf("quit=%v summary=%q", l.quit, l.sh().ExitSummary())
	}
	if !f.closed {
		t.Fatal("the flow was not closed")
	}
}

func TestCtrlCStopsAConversionThatIsRunning(t *testing.T) {
	f := &adoptFake{found: nxRepository(), hold: make(chan struct{})}
	l := newAdoptLab(t, f, false, 120, 36)
	l.keys("a", "enter", "1", "enter") // conservative, then go ahead: the conversion holds
	l.has("Converting…", "step 1 of 1: "+adoptflow.ConvertStepName, "Ctrl+C")
	if l.sh().adopting == nil {
		t.Fatal("no job is running")
	}
	l.keys("ctrl+c")
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		stopped := f.stopped
		f.mu.Unlock()
		if stopped {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.stopped {
		t.Fatal("the job did not see its context cancelled")
	}
}

func TestTheAdoptKeyIsInTheHelpTheFooterAndTheKeymapOfTheGuide(t *testing.T) {
	c, ok := commands.ByID("file.adopt")
	if !ok || c.Kind != commands.OpenAdoptWizard || len(c.Argv) != 0 || c.Confirm {
		t.Fatalf("file.adopt is %+v", c)
	}
	st := commands.State{View: commands.TabGuide}
	if got, _, ok := commands.ByKey("a", st); !ok || got.ID != "file.adopt" {
		t.Fatalf("a in the guide: %+v %v", got, ok)
	}
	if got, _, ok := commands.ByKey("a", commands.State{View: commands.TabProjects, InWorkspace: true}); ok && got.ID == "file.adopt" {
		t.Fatal("a adopts inside a workspace")
	}
	help := ""
	for _, r := range commands.HelpRows(st, "Guide") {
		help += r.Title + "|" + r.Shortcut + "\n"
	}
	if !strings.Contains(help, "Adopt…|a") {
		t.Fatalf("the help lacks the adopt key:\n%s", help)
	}
	view, _ := commands.Hints(st)
	found := false
	for _, h := range view {
		found = found || h.Label == "adopt"
	}
	if !found {
		t.Fatalf("the footer lacks adopt: %+v", view)
	}
}

func TestAClickOnTheBannerOpensTheWizard(t *testing.T) {
	f := &adoptFake{found: nxRepository()}
	l := newAdoptLab(t, f, false, 110, 34)
	sh := l.sh()
	in := sh.mainInner()
	l.send(tea.MouseMsg{X: in.X + 2, Y: in.Y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, Type: tea.MouseLeft})
	if l.sh().adopt == nil {
		t.Fatal("a click on the banner did nothing")
	}
}
