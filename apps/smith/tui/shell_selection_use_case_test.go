package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// copying makes the lab write to a buffer the test can read: what a drag copied.
func copying(l *lab) *bytes.Buffer {
	var out bytes.Buffer
	sh := l.sh()
	sh.clip = &out
	l.m = sh
	return &out
}

func (l *lab) copied(t *testing.T, out *bytes.Buffer) string {
	t.Helper()
	if out.Len() == 0 {
		return ""
	}
	return string(mustDecode(t, out.String()))
}

func TestDraggingAcrossListTextCopiesItAndKeepsTheHighlight(t *testing.T) {
	l := doctorLab(t, 140, 30)
	out := copying(l)
	x, y := l.find("second check")
	l.send(click(x, y)).send(drag(x+3, y)).send(drag(x+5, y))
	if !l.sh().selection.Active || out.Len() != 0 {
		t.Fatal("nothing is copied before the release")
	}
	l.send(letGo(x+5, y))
	if got := l.copied(t, out); got != "second" {
		t.Fatalf("copied %q", got)
	}
	l.has("copied: second")
	if !l.sh().selection.Held {
		t.Fatal("the highlight stays after the release")
	}
	if raw := strings.Split(l.sh().View(), "\n")[y]; !strings.Contains(raw, "\x1b[") || raw == l.lines()[y] {
		t.Fatalf("the selected cells are drawn: %q", raw)
	}
}

func TestAClickWithoutMovementSelectsAndCopiesNothing(t *testing.T) {
	l := doctorLab(t, 140, 30)
	out := copying(l)
	x, y := l.find("second check")
	l.send(click(x, y)).send(letGo(x, y))
	if out.Len() != 0 || l.sh().selection.Selected() {
		t.Fatal("a plain click selects nothing")
	}
}

func TestADragThatStartsOnAButtonOrADividerSelectsNothing(t *testing.T) {
	l := doctorLab(t, 140, 30)
	out := copying(l)
	x, y := l.find("[Check again (r)]")
	l.send(click(x+2, y)).send(drag(x+8, y)).send(letGo(x+8, y))
	if out.Len() != 0 || l.sh().selection.Active {
		t.Fatal("a button starts no selection")
	}

	d := l.sh().spots()[0].Hit
	l.send(click(d.X, d.Y+3))
	if l.sh().split.Dragging == "" {
		t.Fatal("the divider is grabbed")
	}
	l.send(drag(d.X-4, d.Y+3)).send(letGo(d.X-4, d.Y+3))
	if out.Len() != 0 || l.sh().selection.Selected() {
		t.Fatal("a divider starts no selection")
	}
}

func TestAPressOnTheTabsOrTheFrameStartsNothing(t *testing.T) {
	l := doctorLab(t, 140, 30)
	out := copying(l)
	for _, at := range [][2]int{{3, 0}, {3, 2}, {0, 6}, {3, 29}} {
		l.send(click(at[0], at[1])).send(drag(at[0]+5, at[1]+1)).send(letGo(at[0]+5, at[1]+1))
		if out.Len() != 0 || l.sh().selection.Active || l.sh().selection.Selected() {
			t.Fatalf("a press at %v starts a selection", at)
		}
	}
}

func TestEscOnlyClearsTheHighlightAndOtherKeysClearItToo(t *testing.T) {
	l := doctorLab(t, 140, 30)
	copying(l)
	x, y := l.find("second check")
	hold := func() {
		l.send(click(x, y)).send(drag(x+5, y)).send(letGo(x+5, y))
		if !l.sh().selection.Held {
			t.Fatal("a highlight is held")
		}
	}
	hold()
	cursor := l.sh().doctor.Cursor
	l.keys("esc")
	if l.sh().selection.Held || l.sh().doctor.Cursor != cursor {
		t.Fatal("Esc clears the highlight and does nothing else")
	}
	hold()
	l.keys("up")
	if l.sh().selection.Held || l.sh().doctor.Cursor == cursor {
		t.Fatal("another key clears the highlight and still does its work")
	}
}

func TestANewPressTheWheelAndATabChangeClearTheHighlight(t *testing.T) {
	l := doctorLab(t, 140, 30)
	copying(l)
	x, y := l.find("second check")
	for name, act := range map[string]func(){
		"press": func() { l.send(click(x, y-1)).send(letGo(x, y-1)) },
		"wheel": func() { l.send(wheel(x, y, 1)) },
		"tab":   func() { l.keys("2") },
	} {
		l.send(click(x, y)).send(drag(x+5, y)).send(letGo(x+5, y))
		if !l.sh().selection.Held {
			t.Fatalf("%s: a highlight is held", name)
		}
		act()
		if l.sh().selection.Held || l.sh().selection.Active {
			t.Fatalf("%s clears the highlight", name)
		}
		l.keys("3")
	}
}

func TestAResizeClearsTheHighlight(t *testing.T) {
	l := doctorLab(t, 140, 30)
	copying(l)
	x, y := l.find("second check")
	l.send(click(x, y)).send(drag(x+5, y)).send(letGo(x+5, y))
	sh := l.sh()
	sh.resize(120, 30)
	if sh.selection.Held {
		t.Fatal("a resize clears the highlight")
	}
}

func TestADragCannotLeaveThePanelItStartedIn(t *testing.T) {
	l := doctorLab(t, 140, 30)
	out := copying(l)
	x, y := l.find("second check")
	l.send(click(x, y)).send(drag(120, y+8)).send(letGo(120, y+8))
	got := l.copied(t, out)
	if !strings.HasPrefix(got, "second check") || strings.Contains(got, "│") || strings.Contains(got, "What is wrong") || strings.Contains(got, "pkg0") {
		t.Fatalf("only the list part is copied: %q", got)
	}
}

func TestInSettingsADragFromTheSectionsNeverReachesTheDetail(t *testing.T) {
	l, _ := newSettingsLab(t, 140, 30)
	l.keys("6")
	out := copying(l)
	x, y := l.find("Recipes")
	l.send(click(x, y)).send(drag(130, y+1)).send(letGo(130, y+1))
	got := l.copied(t, out)
	if !strings.Contains(got, "Recipes") || !strings.Contains(got, "Forms") {
		t.Fatalf("the sections are copied: %q", got)
	}
	for _, other := range []string{"│", "┃", "[x]", "Enter", "smith plugin"} {
		if strings.Contains(got, other) {
			t.Fatalf("%q is outside the sections list: %q", other, got)
		}
	}
}

func TestInSettingsADragInTheDetailStaysAboveTheSeparator(t *testing.T) {
	l, _ := newSettingsLab(t, 140, 30)
	l.keys("6")
	out := copying(l)
	x, y := l.find("[x] go")
	l.send(click(x, y)).send(drag(x+5, 27)).send(letGo(x+5, 27))
	got := l.copied(t, out)
	if strings.Contains(got, "support for a language") || strings.Contains(got, "Sections") {
		t.Fatalf("the rows above the separator and the text below it are separate regions: %q", got)
	}
}

func TestThePlainThemeStillShowsTheHighlightInReverseVideo(t *testing.T) {
	l := doctorLab(t, 140, 30)
	copying(l)
	sh := l.sh()
	sh.th = theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	l.m = sh
	x, y := l.find("second check")
	before := strings.Split(l.sh().View(), "\n")[y]
	if strings.Contains(before, "\x1b[7m") {
		t.Fatalf("an unselected row has no reverse video: %q", before)
	}
	l.send(click(x, y)).send(drag(x+5, y)).send(letGo(x+5, y))
	after := strings.Split(l.sh().View(), "\n")[y]
	if !strings.Contains(after, "\x1b[7m") {
		t.Fatalf("the highlight is reverse video without colour: %q", after)
	}
}

func TestNoSelectionStartsUnderADialogOrWithTheMouseOff(t *testing.T) {
	l := doctorLab(t, 140, 30)
	out := copying(l)
	x, y := l.find("second check")
	l.keys("ctrl+p")
	l.send(click(x, y)).send(drag(x+5, y)).send(letGo(x+5, y))
	if out.Len() != 0 || l.sh().selection.Selected() {
		t.Fatal("the palette covers the screen: nothing is selected")
	}
}
