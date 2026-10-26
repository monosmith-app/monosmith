package logview

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
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
	out := Log{}.Render(Rows(theme.Default(), lines(2), 20), 20, 5)
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
	out := Log{}.Home().Render(Rows(theme.Default(), lines(50), 30), 30, 3)
	if !strings.Contains(out[0], "line 0") || !strings.Contains(out[2], "line 2") {
		t.Fatalf("%q", out)
	}
}

func TestALongLineWrapsOntoIndentedRowsAndIsNeverCut(t *testing.T) {
	line := "web:build │ src/confluence/confluence.client.ts(284,39): error TS2322: Type 'Buffer<ArrayBufferLike>' is not assignable to type 'BlobPart'."
	rows := Rows(theme.Default(), []string{line}, 40)
	if len(rows) < 3 {
		t.Fatalf("a %d-cell line in 40 cells took %d rows: %q", len(line), len(rows), rows)
	}
	var plain []string
	for _, r := range rows {
		if w := pane.Width(r); w > 40 {
			t.Fatalf("a row is %d cells: %q", w, r)
		}
		plain = append(plain, strings.TrimSpace(pane.Strip(r)))
	}
	joined := strings.Join(strings.Fields(strings.Join(plain, " ")), " ")
	if !strings.Contains(joined, "is not assignable to type 'BlobPart'.") {
		t.Fatalf("the end of the line is gone: %q", joined)
	}
	if !strings.HasPrefix(rows[1], "  ") && !strings.HasPrefix(pane.Strip(rows[1]), "  ") {
		t.Fatalf("a continuation row is not indented: %q", rows[1])
	}
}

func TestTheStatusGlyphsOfTheOutputAreTheThemesOwn(t *testing.T) {
	th := theme.Default()
	rows := Rows(th, []string{"  ✖ web:build exited with code 1", "✔ 3/4 tasks succeeded in 1s"}, 60)
	text := pane.Strip(strings.Join(rows, " "))
	if strings.Contains(text, "✖") || strings.Contains(text, "✔") || !strings.Contains(text, th.Icons.Failed) {
		t.Fatalf("%q", text)
	}
}
