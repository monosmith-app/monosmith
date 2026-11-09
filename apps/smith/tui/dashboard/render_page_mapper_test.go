package dashboard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// plainTheme draws without colour: buttons keep their text between brackets.
func plainTheme() theme.Theme {
	return theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode})
}

// plainLines are the lines of a page without styles.
func plainLines(r Rendered) []string {
	out := make([]string, len(r.Lines))
	for i, l := range r.Lines {
		out[i] = pane.Strip(l)
	}
	return out
}

// textAt is the cells [from, to) of a line.
func textAt(line string, from, to int) string {
	r := []rune(line)
	return string(r[from:min(to, len(r))])
}

var brokenWorkspace = Facts{
	Checked: true, Failed: 2, Fixable: 6, Problems: []string{"gradle is not on PATH", "ci.yml is out of date"}, Passed: 20,
	ProjectsLoaded: true, Projects: 34, Languages: []string{"go", "go", "node", "python"},
	Release: &Ready{Packages: 8, Name: "smith", From: "1.132.0", To: "1.133.0", Commits: 292},
	Recent:  []Run{{Command: "run test --affected", OK: true, Took: "42s", Index: 0}, {Command: "run lint", Took: "12s", Index: 1}, {Command: "run build", OK: true, Took: "1m 4s", Index: 2}},
}

func TestThePageIsAQueueWithTheFirstItemLarge(t *testing.T) {
	r := Page(plainTheme(), brokenWorkspace, 100, Away)
	text := plainLines(r)
	all := strings.Join(text, "\n")
	for _, want := range []string{
		"2 checks fail, 6 can be fixed now",
		"gradle is not on PATH · ci.yml is out of date",
		"[Fix (Enter)]  [Open Doctor (3)]",
		"Ready to release: 8 packages, smith 1.132.0 → 1.133.0 (292 commits)",
		"[Open Release (5)]",
		"Updates not looked at yet",
		"[Open Deps (4)]",
		"Recent runs   ✓ test --affected 42s   × lint 12s   ✓ build 1m 4s",
		"34 projects (2 Go, 1 TypeScript / JavaScript, 1 Python)",
		"[Add a project (a)]  [Compose a stack (A)]",
	} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q in\n%s", want, all)
		}
	}
	for _, gone := range []string{"Where to go", "Health", "Affected", "Last run", "nothing has run yet", "not checked", "nothing changed"} {
		if strings.Contains(all, gone) {
			t.Errorf("the old %q is still on the page:\n%s", gone, all)
		}
	}
	at := func(s string) int {
		for i, l := range text {
			if strings.Contains(l, s) {
				return i
			}
		}
		return -1
	}
	if at("34 projects") >= at("2 checks fail") || at("2 checks fail") >= at("gradle") || at("gradle") >= at("[Fix (") || at("[Fix (") >= at("Ready to release") || at("Ready to release") >= at("Recent runs") {
		t.Fatalf("the order is wrong:\n%s", all)
	}
	if strings.Count(all, "Open Release")+strings.Count(all, "[Fix (") != 2 {
		t.Fatalf("an item shows twice:\n%s", all)
	}
}

func TestEveryLineFitsAndEveryButtonKeepsItsTextInThePlainTheme(t *testing.T) {
	states := map[string]Facts{
		"broken":   brokenWorkspace,
		"waiting":  {Projects: 34, Recent: brokenWorkspace.Recent},
		"empty":    {ProjectsLoaded: true},
		"good":     {Checked: true, DepsChecked: true, ProjectsLoaded: true, Projects: 3},
		"failed":   {Checked: true, DepsChecked: true, Affected: 2, Last: &Run{Command: "run lint test build typecheck --affected -p a-project-with-a-long-name", Repeatable: true}, SettingsTip: true},
		"many":     {Checked: true, Failed: 12, Fixable: 12, Problems: []string{strings.Repeat("a long problem ", 8)}, Affected: 3, Release: &Ready{Packages: 30, Name: "a-package-with-a-long-name", From: "10.20.30", To: "10.21.0", Commits: 1234}, Outdated: 99, DepsChecked: true, Last: &Run{Command: "run build", Repeatable: true}},
		"checkerr": {CheckErr: strings.Repeat("doctor failed because ", 10)},
	}
	for _, w := range []int{80, 120} {
		for name, f := range states {
			r := Page(plainTheme(), f, w, Away)
			text := plainLines(r)
			for i, l := range text {
				if pane.Width(l) != w {
					t.Errorf("%s at %d: line %d is %d cells: %q", name, w, i, pane.Width(l), l)
				}
			}
			for _, tg := range r.Targets {
				if tg.To > w || tg.Row >= len(text) {
					t.Errorf("%s at %d: target %+v is out of the page", name, w, tg)
					continue
				}
				if tg.Command != "" {
					got := textAt(text[tg.Row], tg.From, tg.To)
					if !strings.HasPrefix(got, "[") || !strings.HasSuffix(got, "]") {
						t.Errorf("%s at %d: the button of %s reads %q", name, w, tg.Command, got)
					}
				}
			}
		}
	}
}

func TestEachButtonTargetCoversExactlyItsText(t *testing.T) {
	r := Page(plainTheme(), brokenWorkspace, 100, Away)
	text := plainLines(r)
	want := map[string]string{
		"ws.fix": "[Fix (Enter)]", "view.doctor": "[Open Doctor (3)]", "view.release": "[Open Release (5)]", "view.deps": "[Open Deps (4)]",
		"file.add": "[Add a project (a)]", "file.compose": "[Compose a stack (A)]",
	}
	seen := map[string]bool{}
	for _, tg := range r.Targets {
		if tg.Command == "" {
			continue
		}
		seen[tg.Command] = true
		if got := textAt(text[tg.Row], tg.From, tg.To); got != want[tg.Command] {
			t.Errorf("%s covers %q, want %q", tg.Command, got, want[tg.Command])
		}
	}
	for cmd := range want {
		if !seen[cmd] {
			t.Errorf("no target for %s", cmd)
		}
	}
}

func TestAnEmptyWorkspaceShowsItsTwoButtonsOnce(t *testing.T) {
	text := strings.Join(plainLines(Page(plainTheme(), Facts{ProjectsLoaded: true}, 100, Away)), "\n")
	if strings.Count(text, "[Add a project (a)]") != 1 || strings.Count(text, "[Compose a stack (A)]") != 1 {
		t.Fatalf("the buttons are not shown once:\n%s", text)
	}
	if !strings.Contains(text, "This workspace has no projects yet") {
		t.Fatalf("%s", text)
	}
}

func TestWhileDoctorRunsThePageHasNoEmptyStateRows(t *testing.T) {
	text := strings.Join(plainLines(Page(plainTheme(), Facts{Projects: 5}, 100, Away)), "\n")
	if !strings.Contains(text, "Running the health checks on the 5 projects") {
		t.Fatalf("%s", text)
	}
	for _, gone := range []string{"nothing has run yet", "not checked", "nothing changed", "Recent runs"} {
		if strings.Contains(text, gone) {
			t.Errorf("%q on the page:\n%s", gone, text)
		}
	}
}

func TestACompactItemMovesItsButtonsDownWhenTheLineIsTooNarrow(t *testing.T) {
	f := Facts{Checked: true, Failed: 1, Last: &Run{Command: "run lint", Repeatable: true}, DepsChecked: true}
	r := Page(plainTheme(), f, 40, Away)
	text := plainLines(r)
	for _, tg := range r.Targets {
		if tg.Command == "run.again" {
			if got := textAt(text[tg.Row], tg.From, tg.To); got != "[Run it again]" {
				t.Fatalf("%q", got)
			}
			return
		}
	}
	t.Fatalf("no button to run again:\n%s", strings.Join(text, "\n"))
}

// Every item is drawn the same way: the headline, an optional note, then its
// buttons on a line of their own, left-aligned under the headline text.
func TestEveryItemKeepsItsButtonsOnTheLineBelowItLeftAligned(t *testing.T) {
	f := Facts{Checked: true, Failed: 1, Warnings: 2, Fixable: 3, Problems: []string{"gradle is not on PATH"}, WarningNotes: []string{"lint is old"},
		Affected: 2, DepsChecked: true, Outdated: 3, Last: &Run{Command: "run lint", Repeatable: true}}
	for _, w := range []int{140, 60} {
		r := Page(plainTheme(), f, w, Away)
		text := plainLines(r)
		if len(r.Targets) == 0 {
			t.Fatal("no buttons")
		}
		first := map[int]int{}
		for _, tg := range r.Targets {
			if from, ok := first[tg.Row]; !ok || tg.From < from {
				first[tg.Row] = tg.From
			}
		}
		for row, from := range first {
			if from != 3 && !strings.Contains(text[row], "[Add a project") {
				t.Errorf("width %d: the buttons of %q start at %d, not under the headline text", w, text[row], from)
			}
		}
	}
	text := strings.Join(plainLines(Page(plainTheme(), f, 140, Away)), "\n")
	for _, want := range []string{"[Fix (F)]  [Open Doctor (3)]", "[Open Deps (4)]"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in\n%s", want, text)
		}
	}
	if strings.Count(text, "[Fix (") != 2 || !strings.Contains(text, "[Fix (F)]  [Open Doctor (3)]") {
		t.Errorf("a warning offers the same fix and Doctor buttons as a failure, in the same wording:\n%s", text)
	}
}

func TestTheTargetUnderTheKeyboardLooksHoveredAndCarriesEnter(t *testing.T) {
	th := plainTheme()
	// Without the keyboard the first button of the queue carries Enter.
	if all := strings.Join(plainLines(PageAt(th, brokenWorkspace, 100, Away, -1)), "\n"); !strings.Contains(all, "[Fix (Enter)]  [Open Doctor (3)]") {
		t.Fatalf("no keyboard:\n%s", all)
	}
	// On the second button Enter moves there, and Fix gets its own key back.
	all := strings.Join(plainLines(PageAt(th, brokenWorkspace, 100, Away, 1)), "\n")
	if !strings.Contains(all, "[Fix (F)]  [Open Doctor (Enter)]") {
		t.Fatalf("the keyboard is on Open Doctor:\n%s", all)
	}
	if strings.Count(all, "(Enter)") != 1 {
		t.Fatalf("only the target with the keyboard says Enter:\n%s", all)
	}
}

func TestTheKeyboardReachesEveryTargetInReadingOrder(t *testing.T) {
	r := Page(plainTheme(), brokenWorkspace, 100, Away)
	for i := 1; i < len(r.Targets); i++ {
		a, b := r.Targets[i-1], r.Targets[i]
		if b.Row < a.Row || (b.Row == a.Row && b.From < a.To) {
			t.Fatalf("target %d (%+v) comes before target %d (%+v) in reading order", i, b, i-1, a)
		}
	}
	// Fix, Open Doctor, Open Release, Open Deps, three runs, Add a project, Compose a stack.
	if len(r.Targets) != 9 {
		t.Fatalf("%d targets: %+v", len(r.Targets), r.Targets)
	}
}

func TestARunUnderTheKeyboardIsDrawnChosen(t *testing.T) {
	th := theme.Default()
	plain := strings.Join(plainLines(PageAt(plainTheme(), brokenWorkspace, 100, Away, 4)), "\n")
	chosen := Page(th, brokenWorkspace, 100, Away)
	at := PageAt(th, brokenWorkspace, 100, Away, 4)
	if plain == "" || strings.Join(chosen.Lines, "") == strings.Join(at.Lines, "") {
		t.Fatal("the run under the keyboard looks like the others")
	}
}
