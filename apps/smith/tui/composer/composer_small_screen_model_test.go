package composer

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// small is the screen of m placed on 100x16, between a 3-row header and a
// 2-row footer, as the shell places it.
func small(m Model) Model { return m.Within(3, 2).Place(100, 16) }

// fitsBetweenTheTabsAndTheFooter fails when the screen leaves the free rows or
// draws more rows than its box.
func fitsBetweenTheTabsAndTheFooter(t *testing.T, m Model) {
	t.Helper()
	r := m.Rect()
	if r.Y < 3 || r.Y+r.H > 14 {
		t.Fatalf("the screen covers the tabs or the footer: %+v", r)
	}
	if got := len(m.Lines(theme.Default())); got != r.H {
		t.Fatalf("the screen draws %d rows in a %d-row box", got, r.H)
	}
}

func TestOnASmallScreenBrowsingFitsAndClicksStillTick(t *testing.T) {
	m := small(screen())
	fitsBetweenTheTabsAndTheFooter(t, m)
	m, _ = press(m, "down")
	x, y := cellOf(t, m, "[ ]")
	m, _ = m.Click(x+1, y)
	if m.lang == nil && len(m.Ticked()) == 0 {
		t.Fatal("a click on a row of the short list must tick it")
	}
}

func TestOnASmallScreenTheListKeepsFiveRowsAndTheSummaryMovesToTheTitle(t *testing.T) {
	m := small(screen())
	lines := m.Lines(theme.Default())
	if !strings.Contains(pane.Strip(lines[0]), "nothing ticked yet") {
		t.Errorf("the summary must be in the title: %q", pane.Strip(lines[0]))
	}
	th := theme.Default()
	inner := m.rect.W - 4
	if got := m.listRoom(len(m.statusLines(th, inner)), len(m.keyLines(th, inner))); got < 5 {
		t.Fatalf("the list gets %d rows on 100x16, want at least 5:\n%s", got, text(m))
	}
	if !strings.Contains(text(m), "Esc back") {
		t.Errorf("the short key line must still say every key:\n%s", text(m))
	}
}

func TestOnASmallScreenTheQuestionScrollsAndFollowsTheCursor(t *testing.T) {
	m := typed(small(screen()), "serverless")
	m, _ = press(m, "down", " ")
	if m.stage != stQuestion {
		t.Fatalf("Azure Functions asks for a trigger first, stage %v", m.stage)
	}
	fitsBetweenTheTabsAndTheFooter(t, m)
	n := len(m.chk.Question.Options)
	for range n {
		m, _ = press(m, "down")
	}
	last := m.chk.Question.Options[n-1].Title
	x, y := cellOf(t, m, last)
	if !strings.Contains(pane.Strip(m.Lines(theme.Default())[m.rect.H-2]), "Esc") {
		t.Error("the keys must stay on the last row")
	}
	m, _ = m.Click(x+1, y)
	if got := m.Ticked(); len(got) != 2 {
		t.Fatalf("a click on the scrolled answer must pick it: %v", got)
	}
}

func TestOnASmallScreenThePlanScrollsWithTheKeysAndTheWheel(t *testing.T) {
	plan := compose.Plan{Command: "smith add app web --lang node --with react", Argv: []string{"smith", "add"}}
	for i := range 20 {
		plan.Files = append(plan.Files, compose.PlanLine{Action: "create", Path: "apps/web/file" + string(rune('a'+i))})
	}
	m := small(screen()).WithPlan(plan, nil)
	fitsBetweenTheTabsAndTheFooter(t, m)
	if strings.Contains(text(m), "apps/web/filet") {
		t.Fatal("the test needs a plan taller than the screen")
	}
	m, _ = press(m, "pgdown", "pgdown", "pgdown", "pgdown", "pgdown", "pgdown")
	if !strings.Contains(text(m), "apps/web/filet") {
		t.Fatalf("PgDn must reach the end of the plan:\n%s", text(m))
	}
	m = m.Wheel(-100)
	if !strings.Contains(text(m), "smith add app web") {
		t.Fatalf("the wheel must scroll back to the top:\n%s", text(m))
	}
}
