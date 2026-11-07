package tabs

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

func init() {
	// Draw escape codes even though the tests have no terminal, so the states can be told apart.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func sample() []Tab {
	return []Tab{
		{ID: "projects", Label: "Projects", Short: "Proj", Key: "2"},
		{ID: "run", Label: "Run", Key: "3", Reserve: 3},
		{ID: "doctor", Label: "Doctor", Key: "4", Reserve: 3},
		{ID: "release", Label: "Release", Short: "Rel", Key: "6"},
		{ID: "settings", Label: "Settings", Short: "Sett", Key: "7"},
	}
}

func plain(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = pane.Strip(l)
	}
	return out
}

func props(width int, active string) Props {
	return Props{Theme: theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}), Tabs: sample(), Active: active, Width: width,
		ShortenOrder: []string{"settings", "release", "projects"}}
}

func TestTheStripIsOneRowOfButtonsWithTheirKeysAndABlankBetweenThem(t *testing.T) {
	s := Render(props(90, "run"))
	if len(s.Lines) != 1 {
		t.Fatalf("rows: %d", len(s.Lines))
	}
	line := plain(s.Lines)[0]
	if pane.Width(line) != 90 || !strings.Contains(line, "[Projects (2)] [Run (3)] [Doctor (4)] [Release (6)] [Settings (7)]") {
		t.Fatalf("%q", line)
	}
}

func TestEverySpanFindsItsTab(t *testing.T) {
	s := Render(props(90, "doctor"))
	for _, sp := range s.Spans {
		if got := s.At(sp.X0); got != sp.ID {
			t.Errorf("column %d is %q, want %q", sp.X0, got, sp.ID)
		}
	}
	if s.At(89) != "" {
		t.Errorf("the far end belongs to no tab")
	}
}

func TestNamesAreShortenedBeforeTheStripScrolls(t *testing.T) {
	// Wide enough for the short names, not for the full ones.
	s := Render(props(68, "run"))
	text := strings.Join(plain(s.Lines), "\n")
	if strings.Contains(text, "‹") || strings.Contains(text, "›") {
		t.Fatalf("shortening is enough, no scrolling:\n%s", text)
	}
	if !strings.Contains(text, "Sett") || strings.Contains(text, "Settings") {
		t.Fatalf("Settings should be shortened first:\n%s", text)
	}
	if !strings.Contains(text, "Doctor") {
		t.Fatalf("Doctor is not shortenable:\n%s", text)
	}
}

func TestTheActiveTabIsAlwaysInViewWhenTheStripScrolls(t *testing.T) {
	for w := 34; w <= 110; w++ {
		for _, a := range sample() {
			s := Render(props(w, a.ID))
			found := false
			for _, sp := range s.Spans {
				found = found || sp.ID == a.ID
			}
			for _, l := range s.Lines {
				if pane.Width(l) != w {
					t.Fatalf("width %d: a row is %d cells", w, pane.Width(l))
				}
			}
			if !found {
				t.Fatalf("width %d: %s is not in view:\n%s", w, a.ID, strings.Join(plain(s.Lines), "\n"))
			}
		}
	}
}

func TestMarkersShowWhatIsHiddenAndAreClickable(t *testing.T) {
	s := Render(props(40, "doctor"))
	text := strings.Join(plain(s.Lines), "\n")
	if !strings.Contains(text, "‹") || !strings.Contains(text, "›") {
		t.Fatalf("tabs hidden on both sides need both markers:\n%s", text)
	}
	for _, sp := range s.Spans {
		if sp.ID == Prev && s.At(sp.X0) != Prev {
			t.Fatalf("the left marker is clickable: %q", s.At(sp.X0))
		}
	}
}

func TestABadgeSitsAfterTheButtonAndNeverChangesTheNamesSpelledOut(t *testing.T) {
	without := props(68, "run")
	with := without
	with.Tabs = sample()
	with.Tabs[2].Badge = "×2"
	a := strings.Join(plain(Render(without).Lines), "\n")
	b := strings.Join(plain(Render(with).Lines), "\n")
	if !strings.Contains(b, "Doctor (4)] ×2") {
		t.Fatalf("the badge follows the button:\n%s", b)
	}
	for _, name := range []string{"Sett", "Rel", "Proj"} {
		if strings.Contains(a, name) != strings.Contains(b, name) {
			t.Fatalf("%s was shortened differently with a badge", name)
		}
	}
}

func TestSpansCoverTheButtonAndItsBadge(t *testing.T) {
	p := props(90, "run")
	p.Tabs[2].Badge = "×2"
	s := Render(p)
	for _, sp := range s.Spans {
		if sp.ID == "doctor" && s.At(sp.X1-1) != "doctor" {
			t.Fatalf("the badge is part of the tab: %+v", sp)
		}
	}
}

func TestTheActiveAndTheHoveredTabLookDifferentFromTheOthers(t *testing.T) {
	p := props(90, "run")
	p.Theme = theme.Default()
	rest := Render(p).Lines[0]
	p.Hover = "doctor"
	if Render(p).Lines[0] == rest {
		t.Fatal("a hovered tab must be drawn differently from a resting one")
	}
	p.Hover, p.Active = "", "doctor"
	if Render(p).Lines[0] == rest {
		t.Fatal("the active tab must be drawn differently from a resting one")
	}
}

func TestTinyWidthsDoNotPanic(t *testing.T) {
	for w := 0; w < 12; w++ {
		Render(props(w, "run"))
	}
}
