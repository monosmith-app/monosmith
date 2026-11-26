package dashboard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func TestTheBoxesFillTheWidthOfThePage(t *testing.T) {
	for _, w := range []int{72, 76, 87, 88, 91, 121} {
		text := plainLines(Page(plainTheme(), brokenWorkspace, w, Away))
		for _, i := range []int{lineOf(text, "╭"), lineOf(text, "Last run")} {
			if got := pane.Width(strings.TrimRight(text[i], " ")); got != w-1 {
				t.Errorf("width %d: the row of boxes ends at %d: %q", w, got, text[i])
			}
		}
	}
}

// boxLines are the lines of the page that hold a box border or a box text.
func boxLines(r Rendered) []string {
	var out []string
	for _, l := range plainLines(r) {
		if strings.ContainsAny(l, "╭╰") || strings.Count(l, "│") >= 2 {
			out = append(out, l)
		}
	}
	return out
}

func runsFacts(commands ...string) Facts {
	f := Facts{Checked: true, DepsChecked: true}
	for i, c := range commands {
		f.Recent = append(f.Recent, Run{Command: c, OK: true, Took: "1s", Index: i})
	}
	return f
}

func TestTheBoxesStandInOneRowWhenWideAndInAGridWhenNarrow(t *testing.T) {
	f := brokenWorkspace
	wide := plainLines(Page(plainTheme(), f, 140, Away))
	row := lineOf(wide, "Changed")
	for _, label := range []string{"Last run", "Health", "Release"} {
		if !strings.Contains(wide[row], label) {
			t.Fatalf("wide: %q is not in the row of the labels: %q", label, wide[row])
		}
	}
	if strings.Count(wide[row-1], "╭") != 4 || strings.Count(wide[row+3], "╰") != 4 {
		t.Fatalf("wide: four boxes of five lines:\n%s", strings.Join(wide[row-1:row+4], "\n"))
	}
	narrow := plainLines(Page(plainTheme(), f, 80, Away))
	first, second := lineOf(narrow, "Changed"), lineOf(narrow, "Health")
	if !strings.Contains(narrow[first], "Last run") || !strings.Contains(narrow[second], "Release") || second != first+4 {
		t.Fatalf("narrow: a 2x2 grid of boxes of four lines:\n%s", strings.Join(narrow, "\n"))
	}
	if strings.Count(narrow[first-1], "╭") != 2 || strings.Count(narrow[first+2], "╰") != 2 {
		t.Fatalf("narrow: two boxes in a row:\n%s", strings.Join(narrow[first-1:first+3], "\n"))
	}
	// Narrow boxes keep the label and the value, not the sub-line.
	if strings.Contains(strings.Join(narrow, "\n"), "0 warnings") || !strings.Contains(strings.Join(wide, "\n"), "0 warnings") {
		t.Fatal("the sub-line shows only on a wide page")
	}
}

func TestABoxThatLoadsWaitsAndABoxWithoutFactsSaysSo(t *testing.T) {
	loading := strings.Join(boxLines(Page(plainTheme(), Facts{ReleaseLoading: true}, 140, Away)), "\n")
	if strings.Count(loading, "Loading…") != 3 || !strings.Contains(loading, "Nothing ran yet") {
		t.Fatalf("Changed, Health and Release load; Last run has nothing:\n%s", loading)
	}
	none := strings.Join(boxLines(Page(plainTheme(), Facts{Checked: true, ProjectsLoaded: true, Projects: 2}, 140, Away)), "\n")
	for _, want := range []string{"✓ Nothing changed", "○ Nothing ran yet", "✓ All checks pass", "○ Nothing ready"} {
		if !strings.Contains(none, want) {
			t.Errorf("missing %q:\n%s", want, none)
		}
	}
	// Every box stays, so the grid keeps its shape.
	if got := len(boxLines(Page(plainTheme(), Facts{}, 80, Away))); got != 8 {
		t.Fatalf("the narrow grid has %d lines", got)
	}
}

func TestTheLastRunBoxTellsHowTheNewestRunEnded(t *testing.T) {
	f := runsFacts("run build", "run lint")
	f.Recent[0].OK, f.Recent[0].Failed, f.Recent[0].Passed, f.Recent[0].Cached = false, 1, 17, 9
	text := strings.Join(boxLines(Page(plainTheme(), f, 140, Away)), "\n")
	if !strings.Contains(text, "× 1 failed") || !strings.Contains(text, "17 passed · 9 cached") {
		t.Fatalf("failed:\n%s", text)
	}
	f.Recent[0].OK, f.Recent[0].Failed = true, 0
	if text := strings.Join(boxLines(Page(plainTheme(), f, 140, Away)), "\n"); !strings.Contains(text, "✓ 17 passed") || !strings.Contains(text, "9 cached · 1s") {
		t.Fatalf("passed:\n%s", text)
	}
	f.Recent[0].OK, f.Recent[0].Cancelled = false, true
	if text := strings.Join(boxLines(Page(plainTheme(), f, 140, Away)), "\n"); !strings.Contains(text, "⊘ Stopped") || !strings.Contains(text, "build") {
		t.Fatalf("stopped:\n%s", text)
	}
	f = runsFacts("run lint")
	f.Recent[0].OK = false
	if text := strings.Join(boxLines(Page(plainTheme(), f, 140, Away)), "\n"); !strings.Contains(text, "× Failed") {
		t.Fatalf("failed without tasks:\n%s", text)
	}
}

func TestAClickOnTheLastRunBoxOpensTheNewestRun(t *testing.T) {
	f := runsFacts("run build", "run lint")
	for _, w := range []int{140, 80} {
		r := Page(plainTheme(), f, w, Away)
		text := plainLines(r)
		rows := 0
		for _, tg := range r.Targets {
			if tg.Kind != TargetRun {
				continue
			}
			rows++
			if tg.Run != 0 || !strings.HasPrefix(textAt(text[tg.Row], tg.From, tg.To), "╭") && !strings.HasPrefix(textAt(text[tg.Row], tg.From, tg.To), "│") && !strings.HasPrefix(textAt(text[tg.Row], tg.From, tg.To), "╰") {
				t.Fatalf("width %d: the run target is not on the box: %+v %q", w, tg, text[tg.Row])
			}
			if got := textAt(text[tg.Row], tg.From, tg.To); tg.Row == lineOf(text, "Last run") && !strings.Contains(got, "Last run") {
				t.Fatalf("width %d: the target covers %q", w, got)
			}
		}
		if want := map[int]int{140: 5, 80: 4}[w]; rows != want {
			t.Fatalf("width %d: %d rows of the box answer a click, want %d", w, rows, want)
		}
	}
	for _, tg := range Page(plainTheme(), runsFacts(), 140, Away).Targets {
		if tg.Kind == TargetRun {
			t.Fatal("without a run the box opens nothing")
		}
	}
}

func TestTheHealthBoxTellsFailuresThenWarnings(t *testing.T) {
	for _, c := range []struct {
		f    Facts
		want []string
	}{
		{Facts{Checked: true, Failed: 2, Warnings: 1}, []string{"× 2 failed", "1 warning"}},
		{Facts{Checked: true, Warnings: 2}, []string{"▲ 2 warnings", "0 failed checks"}},
		{Facts{Checked: true, Passed: 12}, []string{"✓ All checks pass", "12 checks passed"}},
		{Facts{CheckErr: "boom"}, []string{"× Could not check"}},
	} {
		text := strings.Join(boxLines(Page(plainTheme(), c.f, 140, Away)), "\n")
		for _, want := range c.want {
			if !strings.Contains(text, want) {
				t.Errorf("%+v: missing %q in\n%s", c.f, want, text)
			}
		}
	}
}

func TestTheReleaseBoxTellsTheReadyPackagesAndTheirBumps(t *testing.T) {
	f := Facts{Checked: true, Release: &Ready{Packages: 3, Name: "kit", To: "1.4.0", Major: 1, Minor: 2}}
	text := strings.Join(boxLines(Page(plainTheme(), f, 140, Away)), "\n")
	if !strings.Contains(text, "✓ 3 ready") || !strings.Contains(text, "1 major · 2 minor") || strings.Contains(text, "patch") {
		t.Fatalf("%s", text)
	}
}
