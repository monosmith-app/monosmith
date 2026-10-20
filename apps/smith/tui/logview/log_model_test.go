package logview

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func lines(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("kit:test │ line %d", i)
	}
	return out
}

func TestALogFollowsTheEndUntilYouScrollUp(t *testing.T) {
	var l Log
	if l.Top(100, 10) != 90 || !l.Following() {
		t.Fatal("a new log follows")
	}
	l = l.Scroll(-3, 100, 10)
	if l.Following() || l.Top(100, 10) != 87 {
		t.Fatalf("top %d following %v", l.Top(100, 10), l.Following())
	}
	if l.Top(150, 10) != 87 {
		t.Fatal("a paused log stays where it is while lines arrive")
	}
	l = l.Scroll(3, 150, 10)
	if l.Following() {
		t.Fatal("still 53 lines from the end")
	}
	l = l.End()
	if !l.Following() || l.Top(150, 10) != 140 {
		t.Fatal("End follows again")
	}
}

func TestScrollingDownToTheEndResumesFollowing(t *testing.T) {
	l := Log{}.Scroll(-5, 100, 10).Scroll(5, 100, 10)
	if !l.Following() {
		t.Fatal("back at the end")
	}
}

func TestToggleFreezesAndResumes(t *testing.T) {
	l := Log{}.Toggle(100, 10)
	if l.Following() || l.Top(120, 10) != 90 {
		t.Fatalf("frozen at %d", l.Top(120, 10))
	}
	if !l.Toggle(120, 10).Following() {
		t.Fatal("resumed")
	}
}

func TestPagesAndHome(t *testing.T) {
	l := Log{}.Page(-1, 100, 10)
	if l.Top(100, 10) != 81 {
		t.Fatalf("%d", l.Top(100, 10))
	}
	if (Log{}).Home().Top(100, 10) != 0 {
		t.Fatal("home")
	}
}

func TestAShortLogIsPaddedToTheHeight(t *testing.T) {
	out := Log{}.Render(theme.Default(), lines(2), 20, 5)
	if len(out) != 5 || !strings.Contains(out[1], "line 1") {
		t.Fatalf("%q", out)
	}
	for _, o := range out {
		if len([]rune(o)) != 20 {
			t.Fatalf("width %d: %q", len([]rune(o)), o)
		}
	}
}

func TestLinesAreDrawnFromTheTopOfThePausedView(t *testing.T) {
	out := Log{}.Home().Render(theme.Default(), lines(50), 30, 3)
	if !strings.Contains(out[0], "line 0") || !strings.Contains(out[2], "line 2") {
		t.Fatalf("%q", out)
	}
}
