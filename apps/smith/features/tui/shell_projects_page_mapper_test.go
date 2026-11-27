package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectpage"
)

// frameEnds are the cells a frame ends a row with.
const frameEnds = "│╮╯┤"

func TestNoSectionOfProjectsOverflowsItsFrameAt80120And150Columns(t *testing.T) {
	for _, w := range []int{80, 120, 150} {
		h := 30
		if w == 150 {
			h = 44
		}
		for _, sec := range projectpage.Sections {
			l := newLab(t, true, w, h)
			taskRun(l, "orders", "build", "✓ 106ms") // a run of orders, so the Runs section has a row
			taskRun(l, "orders", "test", "× exited with code 1")
			l.openTab(commands.TabProjects)
			l.keys("enter") // the keyboard goes to the detail, or the detail opens over the list
			sh := l.sh()
			sh.projSection = sec
			l.m = sh
			lines := l.lines()
			if len(lines) != h {
				t.Fatalf("%d columns, section %d: %d rows", w, sec, len(lines))
			}
			lay := l.sh().lay()
			frame := lay.Main
			if lay.HasDetail() {
				frame = lay.Detail
			}
			for y, line := range lines {
				if got := pane.Width(line); got != w {
					t.Fatalf("%d columns, section %d: row %d is %d cells: %q", w, sec, y, got, line)
				}
				if y <= frame.Y || y >= frame.Y+frame.H-1 {
					continue
				}
				row := []rune(line)
				if end := string(row[frame.X+frame.W-1]); !strings.Contains(frameEnds, end) {
					t.Fatalf("%d columns, section %d: row %d does not end on the frame (%q):\n%s", w, sec, y, end, l.view())
				}
			}
		}
	}
}

func TestTheListOfProjectsFitsItsFrameAtEveryWidth(t *testing.T) {
	for _, w := range []int{80, 120, 150} {
		l := newLab(t, true, w, 30)
		in := l.sh().mainInner()
		for _, line := range l.inside(in) {
			if pane.Width(line) > in.W {
				t.Fatalf("%d columns: a row of the list is wider than it: %q", w, line)
			}
		}
		lay := l.sh().lay()
		for y := lay.Main.Y + 1; y < lay.Main.Y+lay.Main.H-1; y++ {
			if end := string([]rune(l.lines()[y])[lay.Main.X+lay.Main.W-1]); !strings.Contains(frameEnds, end) {
				t.Fatalf("%d columns: row %d of the list does not end on the frame:\n%s", w, y, l.view())
			}
		}
	}
}

func TestBesideTheListTheDetailOfAProjectTakesWhatTheListLeaves(t *testing.T) {
	for _, c := range []struct{ w, list int }{{150, 44}, {120, 41}, {100, 36}} {
		l := newLab(t, true, c.w, 30)
		if got := l.sh().lay().Main.W; got != c.list {
			t.Fatalf("%d columns: the list is %d wide, want %d", c.w, got, c.list)
		}
	}
}
