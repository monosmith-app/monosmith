package toolsdialog

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// plan builds the plan of one missing tool with the given install options.
func plan(name string, options ...toolinstall.Option) toolinstall.ToolPlan {
	p := toolinstall.ToolPlan{
		Status: toolinstall.Status{
			Need:  toolinstall.Need{ToolNeed: protocol.ToolNeed{Name: name, Title: strings.ToUpper(name), Why: "smith needs it", Consequence: "builds fail", Download: "https://example.com/" + name}},
			State: toolinstall.Missing,
		},
		Options: options,
	}
	p.Options = append(p.Options,
		toolinstall.Option{Kind: toolinstall.OptionPage, Value: "page", Label: "Use the official page", Hint: "go to the page"},
		toolinstall.Option{Kind: toolinstall.OptionMyself, Value: "myself", Label: "I will install it myself"},
		toolinstall.Option{Kind: toolinstall.OptionSkip, Value: "skip", Label: "Skip it"},
	)
	p.Default = "myself"
	for _, o := range options {
		if !o.Admin {
			p.Default = o.Value
			break
		}
	}
	if p.Default == "myself" && len(options) > 0 {
		p.Default = options[0].Value // as BuildPlan does: every option needs rights, the first is marked
	}
	return p
}

func install(manager string, admin bool, argv ...string) toolinstall.Option {
	return toolinstall.Option{Kind: toolinstall.OptionInstall, Value: "install:" + manager, Manager: manager, Command: argv, Admin: admin, Label: "Install with " + manager}
}

func press(m Model, keys ...string) (Model, Outcome) {
	var out Outcome
	for _, k := range keys {
		m, out = m.Key(k)
	}
	return m, out
}

func text(m Model) string { return strings.Join(m.Lines(theme.Default()), "\n") }

func TestTheFirstQuestionMarksTheDefaultAndShowsTheExactCommand(t *testing.T) {
	m := New([]toolinstall.ToolPlan{plan("go", install("brew", false, "brew", "install", "go"), install("apt", true, "sudo", "apt-get", "install", "golang-go"))}).Place(100, 30)
	got := text(m)
	for _, want := range []string{"GO is missing", "Install with brew (default)", "Command: brew install go", "esc cancel (nothing is installed)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("the dialog lacks %q:\n%s", want, got)
		}
	}
	m, _ = press(m, "down")
	if got := text(m); !strings.Contains(got, "Command: sudo apt-get install golang-go") || !strings.Contains(got, "administrator rights") {
		t.Fatalf("the command of the option under the cursor is shown with its rights:\n%s", got)
	}
}

func TestAnInstallThatNeedsNoRightsRunsWithoutASecondQuestion(t *testing.T) {
	m := New([]toolinstall.ToolPlan{plan("go", install("brew", false, "brew", "install", "go"))}).Place(100, 30)
	m, out := press(m, "enter")
	if !slices.Equal(out.Run, []string{"tools", "install", "go", "--manager", "brew"}) || !m.Running() {
		t.Fatalf("outcome %+v running %v", out, m.Running())
	}
}

func TestACommandThatNeedsRightsAsksASecondQuestionAndNeverRunsOnItsOwn(t *testing.T) {
	m := New([]toolinstall.ToolPlan{plan("go", install("winget", true, "winget", "install", "--id", "GoLang.Go", "-e"))}).Place(100, 30)
	m, out := press(m, "enter")
	if out.Run != nil || m.Running() || !strings.Contains(text(m), "needs administrator rights") {
		t.Fatalf("a second question comes before any run: %+v\n%s", out, text(m))
	}
	// The default answer is to run it yourself: enter alone starts nothing.
	m2, out := press(m, "enter")
	if out.Run != nil || out.Handover != nil || len(m2.Results()) != 1 || !strings.Contains(m2.Results()[0].Line, "Left to you") {
		t.Fatalf("outcome %+v results %+v", out, m2.Results())
	}
	// Letting smith run it passes --with-admin and streams.
	m3, out := press(m, "down", "enter")
	if !slices.Equal(out.Run, []string{"tools", "install", "go", "--manager", "winget", "--with-admin"}) || !m3.Running() {
		t.Fatalf("outcome %+v", out)
	}
}

func TestASudoCommandIsHandedOverToTheTerminal(t *testing.T) {
	m := New([]toolinstall.ToolPlan{plan("go", install("apt", true, "sudo", "apt-get", "install", "golang-go"))}).Place(100, 30)
	_, out := press(m, "enter", "down", "enter")
	if !out.Done || !slices.Equal(out.Handover, []string{"tools", "install", "go", "--manager", "apt", "--with-admin"}) || out.Run != nil {
		t.Fatalf("a password needs a terminal: %+v", out)
	}
}

func TestEscCancelsWithoutRunningAndBacksOutOfTheSecondQuestion(t *testing.T) {
	m := New([]toolinstall.ToolPlan{plan("go", install("winget", true, "winget", "install", "go"))}).Place(100, 30)
	if _, out := press(m, "esc"); !out.Done || out.Run != nil {
		t.Fatalf("%+v", out)
	}
	back, _ := press(m, "enter", "esc")
	if !strings.Contains(text(back), "How do you want to get it?") {
		t.Fatalf("esc goes back to the first question:\n%s", text(back))
	}
}

func TestTheOutputStreamsAndEachToolEndsWithAResult(t *testing.T) {
	m := New([]toolinstall.ToolPlan{
		plan("go", install("brew", false, "brew", "install", "go")),
		plan("node", install("brew", false, "brew", "install", "node")),
	}).Place(100, 30)
	m, _ = press(m, "enter")
	m = m.Output([]string{"Downloading go", "Pouring go"})
	if got := text(m); !strings.Contains(got, "Running: brew install go") || !strings.Contains(got, "Pouring go") {
		t.Fatalf("the output streams into the dialog:\n%s", got)
	}
	if _, out := press(m, "x"); out.Stop || out.Done {
		t.Fatalf("a stray key does nothing while it runs: %+v", out)
	}
	if _, out := press(m, "esc"); !out.Stop {
		t.Fatalf("esc stops the run: %+v", out)
	}
	m = m.Finished(0, false)
	if !strings.Contains(text(m), "NODE is missing") {
		t.Fatalf("the next tool is asked after the first:\n%s", text(m))
	}
	m, _ = press(m, "enter")
	m = m.Output([]string{"brew: no network"}).Finished(1, false)
	res := m.Results()
	if len(res) != 2 || res[0].Failed || res[0].Line != "Installed." || !res[1].Failed || res[1].Output[0] != "brew: no network" {
		t.Fatalf("results %+v", res)
	}
	got := text(m)
	for _, want := range []string{"done", "Installed.", "brew: no network", "smith tools"} {
		if !strings.Contains(got, want) {
			t.Fatalf("the result lacks %q:\n%s", want, got)
		}
	}
	if _, out := press(m, "enter"); !out.Done {
		t.Fatalf("enter closes the result: %+v", out)
	}
}

func TestThePageAndSkipAnswersRunNothingAndNameTheConsequence(t *testing.T) {
	m := New([]toolinstall.ToolPlan{plan("go"), plan("node")}).Place(100, 30)
	m, out := press(m, "up", "enter")
	if out.Run != nil || !strings.Contains(m.Results()[0].Line, "https://example.com/go") {
		t.Fatalf("the page answer gives the address: %+v", m.Results())
	}
	m, out = press(m, "down", "enter")
	if out.Run != nil || !strings.Contains(m.Results()[1].Line, "builds fail") {
		t.Fatalf("the skip answer says what stops working: %+v", m.Results())
	}
}

func TestWithNoMissingToolTheDialogSaysSo(t *testing.T) {
	m := New(nil).Place(100, 30)
	if got := text(m); !strings.Contains(got, "Every tool the workspace needs is installed.") {
		t.Fatalf("%s", got)
	}
}

func TestAClickPicksAnOptionAndADoubleClickConfirmsIt(t *testing.T) {
	m := New([]toolinstall.ToolPlan{plan("go", install("brew", false, "brew", "install", "go"), install("apt", false, "apt-get", "install", "golang-go"))}).Place(100, 30)
	th := theme.Default()
	y := -1
	for i, l := range m.Lines(th) {
		if strings.Contains(l, "Install with apt") {
			y = m.Rect().Y + i
		}
	}
	if y < 0 {
		t.Fatal("option not drawn")
	}
	m, out := m.Click(th, m.Rect().X+5, y, false)
	if m.cursor != 1 || out.Run != nil {
		t.Fatalf("a single click only moves the cursor: %d %+v", m.cursor, out)
	}
	if _, out = m.Click(th, m.Rect().X+5, y, true); !slices.Equal(out.Run, []string{"tools", "install", "go", "--manager", "apt"}) {
		t.Fatalf("a double click confirms the option: %+v", out)
	}
	if !strings.Contains(text(m), "click an entry, double-click to select") {
		t.Fatal("hint missing")
	}
}
