package depsview

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func sample() Model {
	var rows []Row
	for i := 0; i < 12; i++ {
		rows = append(rows, Row{Name: fmt.Sprintf("@scope/some-long-package-name-%d", i), Current: "1.2.3", Latest: "1.2.9", Level: "patch", Projects: []string{"alpha", "beta", "gamma", "delta", "epsilon"}})
	}
	for i := 0; i < 6; i++ {
		rows = append(rows, Row{Name: fmt.Sprintf("tool-%d", i), Current: "2.0.0", Latest: "2.4.0", Level: "minor", Projects: []string{"alpha"}})
	}
	for i := 0; i < 4; i++ {
		rows = append(rows, Row{Name: fmt.Sprintf("framework-%d", i), Current: "3.0.0", Latest: "4.0.0", Level: "major", Projects: []string{"alpha"}})
	}
	// Found in an unhelpful order: the model sorts them.
	rows[0], rows[len(rows)-1] = rows[len(rows)-1], rows[0]
	return FromResult(Result{Rows: rows})
}

func render(m Model, w, h int) []string {
	lines, _ := m.View(theme.Default(), w, h)
	out := make([]string, len(lines))
	for i, l := range lines {
		if pane.Width(l) > w {
			panic("line wider than the pane: " + pane.Strip(l))
		}
		out[i] = pane.Strip(l)
	}
	return out
}

func TestSafeAndMajorReleasesAreSeparatedAndExplained(t *testing.T) {
	text := strings.Join(render(sample(), 90, 40), "\n")
	for _, want := range []string{"safe to take together", "Patch and minor releases: safe to take together", "Major releases: may break your code"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if strings.Index(text, "Patch and minor") > strings.Index(text, "Major releases") {
		t.Fatalf("safe group should come first:\n%s", text)
	}
}

func TestMarkingTheSafeOnesLeavesTheMajorsUnmarked(t *testing.T) {
	m := sample().Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}, 10)
	if len(m.Marked) != 18 {
		t.Fatalf("marked %d, want the 18 patch and minor rows", len(m.Marked))
	}
	if strings.Contains(strings.Join(render(m, 90, 40), "\n"), "major releases: they can break") {
		t.Fatal("no major is chosen, so no warning")
	}
	m = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}, 10)
	if !strings.Contains(strings.Join(render(m, 90, 40), "\n"), "4 of the chosen packages are major releases") {
		t.Fatalf("marking everything must warn about the majors:\n%s", strings.Join(render(m, 90, 40), "\n"))
	}
}

func TestColumnsStayInsideTheirWidth(t *testing.T) {
	for _, w := range []int{50, 72, 110} {
		render(sample(), w, 24) // panics on an overlong line
	}
}
