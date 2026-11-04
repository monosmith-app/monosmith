package settingsview

import (
	"strings"
	"testing"
	"time"
)

// A change shows its working line at once, before the change ends.
func TestAChangeSaysWhatItDoesAtOnce(t *testing.T) {
	m := opened(t, sample())
	m, _ = m.Key(keyMsg("down"))
	m, res := m.Key(keyMsg(" "))
	if res.Cmd == nil || !m.Busy() {
		t.Fatalf("the change must start and the tab must be busy: busy=%v", m.Busy())
	}
	out := text(m.Pulse(1, 0), 100, 28)
	if !strings.Contains(out, "Turning node off…") {
		t.Fatalf("the working line is missing:\n%s", out)
	}
	if out := text(m.Pulse(1, 5*time.Second), 100, 28); !strings.Contains(out, "Turning node off… 5s") {
		t.Fatalf("the elapsed time shows after 2 s:\n%s", out)
	}
	m, _ = settle(m, res)
	if m.Busy() {
		t.Fatal("the line must go when the change ends")
	}
	if out := text(m, 100, 28); strings.Contains(out, "Turning node off") {
		t.Fatalf("the working line stays:\n%s", out)
	}
}

// While a write runs, a second change waits and Esc says it cannot stop it.
func TestAWriteInProgressCannotBeStartedTwiceOrStopped(t *testing.T) {
	f := sample()
	m := opened(t, f)
	m, _ = m.Key(keyMsg("down"))
	m, res := m.Key(keyMsg(" "))
	m, again := m.Key(keyMsg(" "))
	if again.Cmd != nil || again.Note == "" {
		t.Fatalf("a second change must wait and say so: %+v", again)
	}
	m, esc := m.Key(keyMsg("esc"))
	if !m.Busy() || !strings.Contains(esc.Note, "cannot be stopped") {
		t.Fatalf("Esc must not drop a write: busy=%v note=%q", m.Busy(), esc.Note)
	}
	settle(m, res)
	if got := strings.Join(f.calls, ";"); got != "plugin node false" {
		t.Fatalf("calls: %q", got)
	}
}

// Esc stops the review of a plugin, and the review that arrives later opens nothing.
func TestEscStopsAReviewAndItsLateAnswerIsDropped(t *testing.T) {
	m := opened(t, sample())
	// Rows: built-in go, node, python, then the external widget.
	m, _ = m.Key(keyMsg("down"))
	m, _ = m.Key(keyMsg("down"))
	m, _ = m.Key(keyMsg("down"))
	m, res := m.Key(keyMsg("t"))
	if res.Cmd == nil || !m.Busy() {
		t.Fatalf("the review must start: %+v", res)
	}
	m, esc := m.Key(keyMsg("esc"))
	if m.Busy() || !strings.Contains(esc.Note, "Cancelled") {
		t.Fatalf("Esc must stop the review: busy=%v note=%q", m.Busy(), esc.Note)
	}
	m, _ = settle(m, res)
	if m.ed != nil {
		t.Fatal("a stopped review must not open its editor")
	}
}
