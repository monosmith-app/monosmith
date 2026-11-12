package dashboard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

func runsFacts(commands ...string) Facts {
	f := Facts{Checked: true, DepsChecked: true}
	for i, c := range commands {
		f.Recent = append(f.Recent, Run{Command: c, OK: true, Took: "1s", Index: i})
	}
	return f
}

func recentLine(r Rendered) (string, []Target) {
	var targets []Target
	row := -1
	for _, tg := range r.Targets {
		if tg.Command == "" {
			row = tg.Row
			targets = append(targets, tg)
		}
	}
	if row < 0 {
		return "", nil
	}
	return plainLines(r)[row], targets
}

func TestTheRecentRunsAreOneLineWithATargetEach(t *testing.T) {
	line, targets := recentLine(Page(plainTheme(), runsFacts("run build", "run lint", "run test", "run typecheck"), 100, Away))
	if !strings.Contains(line, "Recent runs") || strings.Contains(line, "typecheck") {
		t.Fatalf("only the last three runs are listed: %q", line)
	}
	if len(targets) != 3 {
		t.Fatalf("%d targets", len(targets))
	}
	for i, want := range []string{"✓ build 1s", "✓ lint 1s", "✓ test 1s"} {
		if got := textAt(line, targets[i].From, targets[i].To); got != want || targets[i].Run != i {
			t.Errorf("run %d covers %q (run %d), want %q", i, got, targets[i].Run, want)
		}
	}
}

func TestARunThatDoesNotFitIsLeftOutAndTheFirstIsCutShort(t *testing.T) {
	long := "run " + strings.Repeat("x", 60)
	line, targets := recentLine(Page(plainTheme(), runsFacts("run build", long, long), 50, Away))
	if pane.Width(line) != 50 || len(targets) != 1 {
		t.Fatalf("%d targets in %q", len(targets), line)
	}
	line, targets = recentLine(Page(plainTheme(), runsFacts(long), 40, Away))
	if len(targets) != 1 || !strings.Contains(line, "…") || targets[0].To > 39 {
		t.Fatalf("the first run is cut with an ellipsis: %q %+v", line, targets)
	}
}

func TestThereIsNoRecentLineWithoutRuns(t *testing.T) {
	if line, _ := recentLine(Page(plainTheme(), runsFacts(), 100, Away)); line != "" {
		t.Fatalf("%q", line)
	}
}

func TestAStoppedRunAndAFailedRunHaveTheirOwnMark(t *testing.T) {
	f := runsFacts("run build", "run lint")
	f.Recent[0].OK, f.Recent[0].Cancelled = false, true
	f.Recent[1].OK = false
	line, _ := recentLine(Page(plainTheme(), f, 100, Away))
	if !strings.Contains(line, "⊘ build") || !strings.Contains(line, "× lint") {
		t.Fatalf("%q", line)
	}
}

func TestTheSelectedRecentRunHasTheCursorMarkInTheGapAndTheTargetsStayPut(t *testing.T) {
	f := runsFacts("run build", "run lint", "run test")
	_, rest := recentLine(Page(plainTheme(), f, 100, Away))
	at := -1
	for i, tg := range Page(plainTheme(), f, 100, Away).Targets {
		if tg.Command == "" && tg.Run == 1 {
			at = i
		}
	}
	line, targets := recentLine(PageAt(plainTheme(), f, 100, Away, at))
	if strings.Count(line, "› ") != 1 || !strings.Contains(line, "› ✓ lint") {
		t.Fatalf("the cursor mark is not before the selected run: %q", line)
	}
	for i := range targets {
		if targets[i].From != rest[i].From || targets[i].To != rest[i].To {
			t.Fatalf("the selection moved run %d: %v, want %v", i, targets[i], rest[i])
		}
	}
}
