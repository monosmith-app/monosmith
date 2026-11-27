package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
)

// mainTop is the top border of the main frame, as plain text.
func mainTop(l *lab) string {
	lay := l.sh().lay()
	return string([]rune(l.lines()[lay.Main.Y])[lay.Main.X : lay.Main.X+lay.Main.W])
}

func TestOnANarrowScreenEnterOpensTheProjectAndEscGoesBackToTheList(t *testing.T) {
	l := newLab(t, true, 80, 30)
	name := l.sh().list.SelectedName()
	if top := mainTop(l); !strings.HasPrefix(top, "╭─ Projects ") {
		t.Fatalf("the list's frame is titled Projects: %q", top)
	}
	l.keys("enter")
	if !l.sh().detailOpen {
		t.Fatal("Enter did not open the detail")
	}
	if top := mainTop(l); !strings.HasPrefix(top, "╭─ ‹ Projects / "+name+" ") {
		t.Fatalf("the detail's frame names the way back: %q", top)
	}
	l.has("Last result")
	l.keys("esc")
	if l.sh().detailOpen {
		t.Fatal("Esc did not go back to the list")
	}
	if top := mainTop(l); !strings.HasPrefix(top, "╭─ Projects ") {
		t.Fatalf("back on the list, the frame is titled Projects again: %q", top)
	}
}

func TestOnANarrowScreenEnterOpensTheCheckOfHealthAndEscGoesBack(t *testing.T) {
	l := newLab(t, true, 80, 30)
	l.keys("4")
	if top := mainTop(l); !strings.HasPrefix(top, "╭─ Health ") {
		t.Fatalf("the list's frame is titled Health: %q", top)
	}
	sh := l.sh()
	f, ok := sh.doctor.Selected()
	if !ok {
		t.Fatal("no check is selected")
	}
	l.keys("enter")
	if !l.sh().detailFull() {
		t.Fatal("Enter did not open the check")
	}
	if top := mainTop(l); !strings.HasPrefix(top, "╭─ ‹ Health / "+f.Check) {
		t.Fatalf("the check's frame names the way back: %q", top)
	}
	l.keys("esc")
	if l.sh().detailFull() {
		t.Fatal("Esc did not go back to the checks")
	}
}

func TestOnAWideScreenTheDetailFrameIsTitledWithTheSelection(t *testing.T) {
	l := newLab(t, true, 150, 44)
	lay := l.sh().lay()
	top := string([]rune(l.lines()[lay.Detail.Y])[lay.Detail.X:])
	if want := "╭─ " + l.sh().list.SelectedName() + " "; !strings.HasPrefix(top, want) {
		t.Fatalf("the detail's frame is titled %q: %q", want, top)
	}
	l.keys(commands.ShortcutOf("view.doctor"))
	f, _ := l.sh().doctor.Selected()
	if top := string([]rune(l.lines()[lay.Detail.Y])[l.sh().lay().Detail.X:]); !strings.HasPrefix(top, "╭─ "+f.Check) {
		t.Fatalf("the detail of Health is titled with the check: %q", top)
	}
}

func TestTheFrameWithTheKeyboardHasTheAccentBorder(t *testing.T) {
	l := newLab(t, true, 150, 44)
	if !l.sh().mainFocused() {
		t.Fatal("the list has the keyboard first")
	}
	l.keys("right")
	if l.sh().mainFocused() {
		t.Fatal("Right gives the keyboard to the detail")
	}
}
