package detail

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func problemPanel(n int, expanded bool) Panel {
	var items []string
	for i := 0; i < n; i++ {
		items = append(items, "project-"+string(rune('a'+i))+":build writes packages/x/dist but its task caches dist/packages/x: a cache hit would leave dist empty")
	}
	return Panel{
		Sections: []Section{{Heading: "check", Rows: []Row{{"state", "failed"}}}, {Heading: "How to fix", Items: []string{"declare the outputs"}}},
		Actions:  []Action{{ID: "fix", Label: "Fix this"}},
		After:    []Section{{Heading: "Problem", Items: items, Bullets: true, Collapse: 5}},
		Expanded: expanded,
	}
}

func TestTheRemedyAndTheActionsComeBeforeALongProblem(t *testing.T) {
	lines, hits := problemPanel(20, false).View(theme.Default(), 40, "")
	text := strings.Join(lines, "\n")
	fix, act, prob := strings.Index(text, "How to fix"), strings.Index(text, "Actions"), strings.Index(text, "Problem")
	if fix < 0 || act < fix || prob < act {
		t.Fatalf("order must be remedy, actions, problem:\n%s", text)
	}
	if len(hits) != 1 {
		t.Fatalf("hits %v", hits)
	}
	for row := range hits {
		if row > 10 {
			t.Fatalf("the action must be near the top, it is on row %d", row)
		}
	}
}

func TestAProblemShowsFiveBulletsAndCountsTheRest(t *testing.T) {
	lines, _ := problemPanel(20, false).View(theme.Default(), 60, "")
	text := strings.Join(lines, "\n")
	if got := strings.Count(text, "•"); got != 5 {
		t.Fatalf("%d bullets:\n%s", got, text)
	}
	if !strings.Contains(text, "+15 more (e shows all)") {
		t.Fatalf("the count of the rest is missing:\n%s", text)
	}
}

func TestExpandingShowsEveryBullet(t *testing.T) {
	lines, _ := problemPanel(20, true).View(theme.Default(), 60, "")
	text := strings.Join(lines, "\n")
	if got := strings.Count(text, "•"); got != 20 {
		t.Fatalf("%d bullets", got)
	}
	if strings.Contains(text, "more (e shows all)") || !strings.Contains(text, "e shows fewer") {
		t.Fatalf("expanded panel should offer to fold:\n%s", text)
	}
}

func TestABulletWrapsUnderItsText(t *testing.T) {
	lines, _ := problemPanel(1, false).View(theme.Default(), 30, "")
	found := false
	for _, l := range lines {
		if strings.HasPrefix(l, "    ") && !strings.HasPrefix(l, "     •") && !strings.Contains(l, "•") && strings.TrimSpace(l) != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("a long bullet should wrap with a hanging indent:\n%s", strings.Join(lines, "\n"))
	}
}

func TestFewItemsAreNotCollapsed(t *testing.T) {
	lines, _ := problemPanel(3, false).View(theme.Default(), 60, "")
	text := strings.Join(lines, "\n")
	if strings.Contains(text, "more (") || strings.Count(text, "•") != 3 {
		t.Fatalf("%s", text)
	}
}
