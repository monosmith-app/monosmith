package runpage

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

var plain = theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode})

func TestATaskRowHasItsBranchMarkNameAndTimeAtTheRight(t *testing.T) {
	row := pane.Strip(TaskRow(plain, Row{Project: "web", TaskID: "web:test", Last: true}, TaskLine{Mark: "✓", Took: "3.2s", Cache: "hit"}, 40))
	if pane.Width(row) != 40 || !strings.HasPrefix(row, "  └─ ✓ test") || !strings.HasSuffix(strings.TrimRight(row, " "), "3.2s  hit") {
		t.Fatalf("row %q", row)
	}
	sel := pane.Strip(TaskRow(plain, Row{Project: "web", TaskID: "web:test"}, TaskLine{Mark: "×", Selected: true}, 40))
	if !strings.HasPrefix(sel, "› ├─ × test") {
		t.Fatalf("selected row %q", sel)
	}
}

func TestAHeadingNamesItsProjectAndATaskWithoutOneIsUnderOtherTasks(t *testing.T) {
	if got := pane.Strip(HeadingLine(plain, "web", 20)); strings.TrimSpace(got) != "web" {
		t.Fatalf("heading %q", got)
	}
	if got := pane.Strip(HeadingLine(plain, "", 20)); strings.TrimSpace(got) != "Other tasks" {
		t.Fatalf("heading %q", got)
	}
}

func TestARunOfTheSessionIsCutToItsWidthWithItsAgeAtTheRight(t *testing.T) {
	row := pane.Strip(SessionRow(plain, SessionLine{Mark: "×", Command: "smith run build --affected --all-the-things", Ago: "2m ago", Selected: true}, 30))
	if pane.Width(row) != 30 || !strings.HasPrefix(row, "› × smith") || !strings.Contains(row, "…") || !strings.HasSuffix(row, "2m ago") {
		t.Fatalf("row %q", row)
	}
}

func TestAllOutputCountsTheLines(t *testing.T) {
	if got := pane.Strip(AllOutputRow(plain, 1, false, 30)); !strings.Contains(got, "All output  1 line") {
		t.Fatalf("row %q", got)
	}
}
