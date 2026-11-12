package packageprompt

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// fits fails when the dialog leaves the rows between a 3-row header and a
// 2-row footer of a 16-row screen, or draws more rows than its box.
func fits(t *testing.T, p Prompt) {
	t.Helper()
	r := p.Rect()
	if r.Y < 3 || r.Y+r.H > 14 {
		t.Fatalf("the dialog covers the tabs or the footer: %+v", r)
	}
	if got := len(p.Lines(plain())); got != r.H {
		t.Fatalf("the dialog draws %d rows in a %d-row box", got, r.H)
	}
}

func TestOnASmallScreenInstallPackagesScrollsAboveItsButtons(t *testing.T) {
	p := NewInstall("web", []string{"web", "api"}).WithLanguages(map[string]string{"web": "node"}).WithLibraries(libs).Within(3, 2).Place(100, 16)
	fits(t, p)
	p = arrow(arrow(p, tea.KeyTab), tea.KeyTab)
	if p.focus != focusDev {
		t.Fatalf("Tab goes to the development box, focus %d", p.focus)
	}
	lines := p.Lines(plain())
	if !strings.Contains(screen(p), "[ ]") {
		t.Fatalf("the body must follow the focus to the box:\n%s", screen(p))
	}
	if !strings.Contains(pane.Strip(lines[len(lines)-2]), "Cancel") {
		t.Fatalf("the buttons must stay on the last row:\n%s", screen(p))
	}
	for y := p.Rect().Y; y < p.Rect().Y+p.Rect().H; y++ {
		if strings.Contains(pane.Strip(lines[y-p.Rect().Y]), "[ ]") {
			p, _ = p.Click(p.Rect().X+4, y)
			break
		}
	}
	if !p.Dev() {
		t.Fatal("a click on the scrolled box must tick it")
	}
}

func TestTheDialogSizesItselfAgainWhenItsBodyGrows(t *testing.T) {
	p := NewInstall("web", []string{"web"}).Within(3, 2).Place(100, 30)
	before := p.Rect().H
	p = typeText(p, "a-very-long-package-name another-long-package-name a-third-long-package-name yet-another-one")
	if p.Rect().H <= before {
		t.Fatalf("the command wraps on more lines, so the dialog must grow: %d then %d", before, p.Rect().H)
	}
	if !strings.Contains(strings.Join(strings.Fields(screen(p)), " "), "yet-another-one") {
		t.Fatalf("the whole command must show:\n%s", screen(p))
	}
}

// scrollbarOnTheBorder fails unless the lines draw a scrollbar and every cell
// of its track is the last one inside the right border, as in every pane.
func scrollbarOnTheBorder(t *testing.T, lines []string) {
	t.Helper()
	track := []rune(theme.Default().Icons.Track)[0]
	bars := 0
	for _, l := range lines {
		r := []rune(pane.Strip(l))
		for x, c := range r {
			if c != track {
				continue
			}
			if x != len(r)-2 {
				t.Fatalf("the scrollbar is in column %d, not next to the border in column %d: %q", x, len(r)-2, string(r))
			}
			bars++
		}
	}
	if bars == 0 {
		t.Fatalf("the test needs a body that scrolls:\n%s", pane.Strip(strings.Join(lines, "\n")))
	}
}

func TestTheInstallPackagesScrollbarSitsNextToTheBorder(t *testing.T) {
	p := NewInstall("web", []string{"web", "api"}).WithLanguages(map[string]string{"web": "node"}).WithLibraries(libs).Within(3, 2).Place(100, 16)
	scrollbarOnTheBorder(t, p.Lines(theme.Default()))
}
