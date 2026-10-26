package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

// overlayCases open each layer that is drawn over the screen, on a tab.
var overlayCases = map[string]struct {
	tab  string
	open func(l *lab)
}{
	"palette":       {commands.TabProjects, func(l *lab) { l.keys("ctrl+p") }},
	"palette typed": {commands.TabProjects, func(l *lab) { l.keys("ctrl+p").typeText("e") }},
	"palette home":  {commands.TabHome, func(l *lab) { l.keys("ctrl+p") }},
	"help":          {commands.TabProjects, func(l *lab) { l.keys("f1") }},
	"package form":  {commands.TabProjects, func(l *lab) { l.keys("+") }},
	"confirmation":  {commands.TabRelease, func(l *lab) { l.keys("x") }},
	"toast":         {commands.TabProjects, func(l *lab) { l.toast("Passed in 734ms", toast.Success) }},
	"long toast":    {commands.TabRun, func(l *lab) { l.toast(strings.Repeat("a very long message ", 12), toast.Error) }},
}

// toast raises a toast the way a finished run does.
func (l *lab) toast(text string, level toast.Level) {
	sh := l.sh()
	sh.notify(text, level)
	l.m = sh
}

// TestOverlaysStayInsideTheFrameAtEverySize checks, for each layer drawn over
// the screen, that the screen keeps its size and that the outer frame (its
// left and right cells) is not touched on the content rows; a toast must
// also leave the frame's bottom edge whole.
func TestOverlaysStayInsideTheFrameAtEverySize(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {140, 42}} {
		w, h := sz[0], sz[1]
		for name, c := range overlayCases {
			l := newLab(t, true, w, h)
			l.openTab(c.tab)
			base := l.lines()
			c.open(l)
			lines := l.lines()
			if len(lines) != h {
				t.Fatalf("%s at %dx%d: %d lines", name, w, h, len(lines))
			}
			lay := l.sh().lay()
			for y, line := range lines {
				if got := len([]rune(line)); got != w {
					t.Fatalf("%s at %dx%d: line %d is %d cells wide: %q", name, w, h, y, got, line)
				}
				if y < lay.Main.Y || y >= lay.Footer.Y {
					continue
				}
				now, was := []rune(line), []rune(base[y])
				if now[0] != was[0] || now[w-1] != was[w-1] {
					t.Errorf("%s at %dx%d: the frame is broken on line %d:\n%s\nwas\n%s", name, w, h, y, line, base[y])
				}
			}
			if strings.HasPrefix(name, "palette") {
				continue
			}
			if y := lay.Footer.Y - 1; strings.Contains(name, "toast") && lines[y] != base[y] {
				t.Errorf("%s at %dx%d: the frame's bottom edge changed:\n%s\nwas\n%s", name, w, h, lines[y], base[y])
			}
		}
	}
}

func TestPaletteFitsInsideTheFrameAt80x24(t *testing.T) {
	l := newLab(t, true, 80, 24)
	l.keys("ctrl+p")
	lay := l.sh().lay()
	box := l.sh().paletteBox()
	if box.X < 2 || box.X+box.W > 78 || box.Y < lay.Main.Y || box.Y+box.H > lay.Footer.Y {
		t.Fatalf("the palette %+v is not inside the frame %+v", box, lay.Main)
	}
}

func TestBehindAModalTheViewIsEmptiedSoNoHalfWordsShowAtItsEdges(t *testing.T) {
	l := newLab(t, true, 140, 42)
	l.openTab(commands.TabProjects)
	l.has("language", "kind")
	l.keys("f1")
	// The list and the detail are gone; the tabs and the frame stay.
	l.hasNot("language", "Run build")
	l.has("Projects", "Keyboard shortcuts")
	l.keys("esc")
	l.has("language", "kind")
}

func TestAToastSitsInsideTheContentAreaAndNeverOnTheFrame(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {140, 42}} {
		l := newLab(t, true, sz[0], sz[1])
		l.openTab(commands.TabProjects)
		l.toast("copied: orders", toast.Success)
		var rows []int
		for y, line := range l.lines() {
			if strings.Contains(line, "copied: orders") {
				rows = append(rows, y)
			}
		}
		lay := l.sh().lay()
		if len(rows) != 1 || rows[0] < lay.Main.Y+1 || rows[0] >= lay.Footer.Y-1 {
			t.Fatalf("%v: the toast is on rows %v, the content is %d..%d", sz, rows, lay.Main.Y+1, lay.Footer.Y-2)
		}
	}
}
