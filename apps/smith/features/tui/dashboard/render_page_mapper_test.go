package dashboard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
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

// cellOf is the cell at which s starts in line, -1 when it is not there.
func cellOf(line, s string) int {
	i := strings.Index(line, s)
	if i < 0 {
		return -1
	}
	return pane.Width(line[:i])
}

// lineOf is the first line of text that holds s, -1 when none does.
func lineOf(text []string, s string) int {
	for i, l := range text {
		if strings.Contains(l, s) {
			return i
		}
	}
	return -1
}

var brokenWorkspace = Facts{
	Checked: true, Failed: 2, Fixable: 6, Problems: []string{"gradle is not on PATH", "ci.yml is out of date"}, Passed: 20,
	ProjectsLoaded: true, Projects: 34, Languages: []string{"go", "go", "node", "python"},
	Release: &Ready{Packages: 8, Name: "smith", From: "1.132.0", To: "1.133.0", Commits: 292},
	Recent:  []Run{{Command: "run test --affected", OK: true, Took: "42s", Index: 0}, {Command: "run lint", Took: "12s", Index: 1}, {Command: "run build", OK: true, Took: "1m 4s", Index: 2}},
}

func TestHomeShowsTheSummaryTheBoxesTheListAndStartSomething(t *testing.T) {
	r := Page(plainTheme(), brokenWorkspace, 140, Away)
	text := plainLines(r)
	all := strings.Join(text, "\n")
	for _, want := range []string{
		"34 projects (2 Go, 1 TypeScript / JavaScript, 1 Python) · 0 changed",
		"Changed", "Last run", "Health", "Release",
		"Needs you now  1/3",
		"│ × Failed   2 checks fail, 6 can be fixed now",
		"gradle is not on PATH · ci.yml is out of date",
		"[Fix (Enter)]  [Open Health (3)]",
		"Ready to release: 8 packages, smith 1.132.0 → 1.133.0 (292 commits)",
		"[Open Release (5)]",
		"Updates not looked at yet",
		"[Open Deps (4)]",
		"Start something",
		"[Add a project (a)]  [Compose a stack (A)]",
	} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q in\n%s", want, all)
		}
	}
	for _, gone := range []string{"Where to go", "Recent runs", "nothing has run yet", "not checked"} {
		if strings.Contains(all, gone) {
			t.Errorf("the old %q is still on the page:\n%s", gone, all)
		}
	}
	order := []string{"34 projects", "╭", "Needs you now", "2 checks fail", "gradle", "Ready to release", "Start something", "[Add a project"}
	for i := 1; i < len(order); i++ {
		if lineOf(text, order[i-1]) >= lineOf(text, order[i]) {
			t.Fatalf("%q does not come before %q:\n%s", order[i-1], order[i], all)
		}
	}
	if strings.Count(all, "Open Release")+strings.Count(all, "[Fix (") != 2 {
		t.Fatalf("an item shows twice:\n%s", all)
	}
}

func TestTheTopLineNamesTheWorkspaceAndLeavesOutWhatIsNotKnown(t *testing.T) {
	f := brokenWorkspace
	f.Workspace, f.Base, f.Affected = "acme", "main", 7
	f.Recent = []Run{{Command: "run test", OK: true, Ago: "3 min ago"}}
	top := plainLines(Page(plainTheme(), f, 140, Away))[1]
	if !strings.Contains(top, "acme  34 projects (2 Go, 1 TypeScript / JavaScript, 1 Python) · 7 changed since main · last run 3 min ago") {
		t.Fatalf("%q", top)
	}
	// A short line drops the languages first.
	if top := plainLines(Page(plainTheme(), f, 80, Away))[1]; !strings.Contains(top, "acme  34 projects · 7 changed since main · last run 3 min ago") {
		t.Fatalf("narrow: %q", top)
	}
	// Before the projects arrive and before any run, only the name is there.
	if top := plainLines(Page(plainTheme(), Facts{Workspace: "acme"}, 140, Away))[1]; strings.TrimSpace(top) != "acme" {
		t.Fatalf("nothing known: %q", top)
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
	for _, w := range []int{72, 80, 100, 120} {
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
				got := textAt(text[tg.Row], tg.From, tg.To)
				switch tg.Kind {
				case TargetButton:
					if !strings.HasPrefix(got, "[") || !strings.HasSuffix(got, "]") {
						t.Errorf("%s at %d: the button of %s reads %q", name, w, tg.Command, got)
					}
				case TargetKey:
					if w >= WideFrom || got == "" || strings.ContainsAny(got, "[] ") {
						t.Errorf("%s at %d: the key of %s reads %q", name, w, tg.Command, got)
					}
				}
			}
		}
	}
}

func TestEachButtonTargetCoversExactlyItsText(t *testing.T) {
	r := Page(plainTheme(), brokenWorkspace, 140, Away)
	text := plainLines(r)
	want := map[string]string{
		"ws.fix": "[Fix (Enter)]", "view.doctor": "[Open Health (3)]", "view.release": "[Open Release (5)]", "view.deps": "[Open Deps (4)]",
		"run.build-all": "[Build all (B)]", "run.lint-all": "[Lint all (L)]", "file.palette": "[Run a task… (Ctrl+K)]",
		"file.add": "[Add a project (a)]", "file.compose": "[Compose a stack (A)]", "view.settings": "[Settings (6)]",
	}
	seen := map[string]bool{}
	for _, tg := range r.Targets {
		if tg.Kind != TargetButton {
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
	text := strings.Join(plainLines(Page(plainTheme(), Facts{ProjectsLoaded: true}, 140, Away)), "\n")
	if strings.Count(text, "[Add a project (") != 1 || strings.Count(text, "[Compose a stack (A)]") != 1 {
		t.Fatalf("the buttons are not shown once:\n%s", text)
	}
	for _, want := range []string{"This workspace has no projects yet", "No projects yet"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "Start something") {
		t.Fatalf("the item already offers what Start something would:\n%s", text)
	}
}

func TestWhileDoctorRunsTheListWaitsAndTheBoxesLoad(t *testing.T) {
	text := strings.Join(plainLines(Page(plainTheme(), Facts{Projects: 5}, 140, Away)), "\n")
	if !strings.Contains(text, "Running the health checks on the 5 projects") {
		t.Fatalf("%s", text)
	}
	if strings.Count(text, "Loading…") != 2 {
		t.Fatalf("the Changed and Health boxes wait for their facts:\n%s", text)
	}
	for _, gone := range []string{"not checked", "Nothing changed", "Recent runs"} {
		if strings.Contains(text, gone) {
			t.Errorf("%q on the page:\n%s", gone, text)
		}
	}
}

func TestANarrowItemShowsOnlyTheKeyOfItsAction(t *testing.T) {
	f := Facts{Checked: true, Failed: 1, Last: &Run{Command: "run lint", Repeatable: true}, DepsChecked: true}
	for _, ch := range []Choice{{Item: 0, Button: 0}, {Item: 0, Button: 1}} {
		r := Render(plainTheme(), f, 80, Away, ch)
		text := plainLines(r)
		all := strings.Join(text, "\n")
		if strings.Contains(all, "[") || strings.Contains(all, "Failed  ") {
			t.Fatalf("a narrow item has no buttons and no tag:\n%s", all)
		}
		var key *Target
		for i, tg := range r.Targets {
			if tg.Kind == TargetKey && tg.Item == 0 {
				key = &r.Targets[i]
			}
		}
		if key == nil || textAt(text[key.Row], key.From, key.To) != "Enter" || key.To != 79 {
			t.Fatalf("the selected item ends in Enter: %+v\n%s", key, all)
		}
		want := map[int]string{0: "x:why-failed", 1: "run.again"}[ch.Button]
		if key.Command != want || key.Button != ch.Button {
			t.Fatalf("Enter on button %d runs %q, want %q", ch.Button, key.Command, want)
		}
	}
	// The other items show the key of their first button, in the same column.
	r := Page(plainTheme(), f, 80, Away)
	text := plainLines(r)
	for _, tg := range r.Targets {
		if tg.Kind == TargetKey && tg.Command == "view.doctor" && textAt(text[tg.Row], tg.From, tg.To) == "3" && tg.To == 79 {
			return
		}
	}
	t.Fatalf("no key 3 for Open Health:\n%s", strings.Join(text, "\n"))
}

// On a wide page every item has its buttons at the right of its title line,
// the last one ending one cell before the edge, and its detail under the title.
func TestEveryItemHasItsButtonsAtTheRightOfItsTitleLine(t *testing.T) {
	f := Facts{Checked: true, Failed: 1, Warnings: 2, Fixable: 3, Problems: []string{"gradle is not on PATH"}, WarningNotes: []string{"lint is old"},
		Affected: 2, DepsChecked: true, Outdated: 3, Last: &Run{Command: "run lint", Repeatable: true}}
	r := Page(plainTheme(), f, 140, Away)
	text := plainLines(r)
	if len(r.Items) != len(Queue(f)) {
		t.Fatalf("%d items drawn for %d", len(r.Items), len(Queue(f)))
	}
	last := map[int]int{}
	for _, tg := range r.Targets {
		if tg.Kind != TargetButton || tg.Item < 0 {
			continue
		}
		if tg.Row != r.Items[tg.Item].From {
			t.Errorf("a button of item %d is on line %d, not on its title line %d", tg.Item, tg.Row, r.Items[tg.Item].From)
		}
		last[tg.Item] = max(last[tg.Item], tg.To)
	}
	for item, to := range last {
		if to != 139 {
			t.Errorf("the buttons of item %d end at %d: %q", item, to, text[r.Items[item].From])
		}
	}
	for i, rows := range r.Items {
		if rows.To-rows.From != 2 {
			t.Errorf("item %d takes %d lines", i, rows.To-rows.From)
		}
	}
	all := strings.Join(text, "\n")
	title := lineOf(text, "1 check fails")
	if cellOf(text[title+1], "gradle") != cellOf(text[title], "1 check fails") {
		t.Errorf("the detail is not under the title:\n%s\n%s", text[title], text[title+1])
	}
	for _, want := range []string{"[Fix (F)]  [Open Health (3)]", "[Open Deps (4)]"} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q in\n%s", want, all)
		}
	}
	if strings.Count(all, "[Fix (") != 2 {
		t.Errorf("a warning offers the same fix and Health buttons as a failure:\n%s", all)
	}
}

func TestTheButtonUnderTheKeyboardIsPrimaryAndCarriesEnter(t *testing.T) {
	th := plainTheme()
	if all := strings.Join(plainLines(Render(th, brokenWorkspace, 140, Away, Choice{})), "\n"); !strings.Contains(all, "[Fix (Enter)]  [Open Health (3)]") {
		t.Fatalf("on the first item:\n%s", all)
	}
	// On the second button Enter moves there, and Fix gets its own key back.
	all := strings.Join(plainLines(Render(th, brokenWorkspace, 140, Away, Choice{Item: 0, Button: 1})), "\n")
	if !strings.Contains(all, "[Fix (F)]  [Open Health (Enter)]") || strings.Count(all, "(Enter)") != 1 {
		t.Fatalf("the keyboard is on Open Health:\n%s", all)
	}
	// On the next item its button carries Enter, and the count follows.
	all = strings.Join(plainLines(Render(th, brokenWorkspace, 140, Away, Choice{Item: 1})), "\n")
	if !strings.Contains(all, "[Open Release (Enter)]") || !strings.Contains(all, "Needs you now  2/3") || strings.Count(all, "(Enter)") != 1 {
		t.Fatalf("the keyboard is on the release:\n%s", all)
	}
	// Without the keyboard no button says Enter and there is no count.
	all = strings.Join(plainLines(PageAt(th, brokenWorkspace, 140, Away, -1)), "\n")
	if strings.Contains(all, "(Enter)") || strings.Contains(all, "1/3") || strings.Contains(all, " │ × Failed") {
		t.Fatalf("no keyboard:\n%s", all)
	}
	// In colour the selected button is the primary one: it differs from the same button at rest.
	col := theme.Default()
	on := strings.Join(Render(col, brokenWorkspace, 140, Away, Choice{Item: 0, Button: 1}).Lines, "\n")
	off := strings.Join(Render(col, brokenWorkspace, 140, Away, Choice{Item: 1}).Lines, "\n")
	if on == off || !strings.Contains(on, col.ButtonActive(" Open Health (Enter) ")) {
		t.Fatal("the selected button is not drawn primary")
	}
}

func TestTheTargetsAreInReadingOrder(t *testing.T) {
	r := Page(plainTheme(), brokenWorkspace, 140, Away)
	buttons := 0
	for i := 1; i < len(r.Targets); i++ {
		a, b := r.Targets[i-1], r.Targets[i]
		if b.Row < a.Row || (b.Row == a.Row && b.From < a.To) {
			t.Fatalf("target %d (%+v) comes before target %d (%+v) in reading order", i, b, i-1, a)
		}
	}
	for _, tg := range r.Targets {
		if tg.Kind == TargetButton {
			buttons++
		}
	}
	// Fix, Open Health, Open Release, Open Deps and the six of Start something.
	if buttons != 10 {
		t.Fatalf("%d buttons: %+v", buttons, r.Targets)
	}
}

func TestTheChosenItemHasTheAccentBarAndAnAccentTitle(t *testing.T) {
	plain := plainLines(Render(plainTheme(), brokenWorkspace, 140, Away, Choice{Item: 1}))
	at := lineOf(plain, "Ready to release")
	if !strings.HasPrefix(plain[at], " │ ") || !strings.HasPrefix(plain[at+1], " │ ") {
		t.Fatalf("no bar on both lines: %q / %q", plain[at], plain[at+1])
	}
	if first := lineOf(plain, "2 checks fail"); strings.HasPrefix(plain[first], " │") {
		t.Fatalf("the other item has the bar too: %q", plain[first])
	}
	th := theme.Default()
	a := Render(th, brokenWorkspace, 140, Away, Choice{Item: 0})
	b := Render(th, brokenWorkspace, 140, Away, Choice{Item: 1})
	if a.Lines[at] == b.Lines[at] || !strings.Contains(b.Lines[at], th.Accent("│")) {
		t.Fatal("the chosen item looks like the others")
	}
}

func TestAClickTargetCoversEachLineOfAnItem(t *testing.T) {
	r := Page(plainTheme(), brokenWorkspace, 140, Away)
	for i, rows := range r.Items {
		for row := rows.From; row < rows.To; row++ {
			found := false
			for _, tg := range r.Targets {
				if tg.Kind == TargetItem && tg.Item == i && tg.Row == row && tg.From == 0 {
					found = true
				}
			}
			if !found {
				t.Errorf("line %d of item %d selects nothing", row, i)
			}
		}
	}
}
