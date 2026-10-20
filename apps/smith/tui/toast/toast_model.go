package toast

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Level is how serious the message is.
type Level int

// The levels.
const (
	Info Level = iota
	Success
	Warning
	Error
)

// Lifetime is how many ticks a toast stays (ticks are 100ms apart).
const Lifetime = 40

// Toast is a message with a lifetime. The zero value is not shown.
type Toast struct {
	Text  string
	Level Level
	left  int
}

// New shows text for Lifetime ticks.
func New(text string, level Level) Toast { return Toast{Text: text, Level: level, left: Lifetime} }

// Shown reports whether the toast is visible.
func (t Toast) Shown() bool { return t.left > 0 && t.Text != "" }

// Tick ages the toast by one tick.
func (t Toast) Tick() Toast {
	if t.left > 0 {
		t.left--
	}
	return t
}

// Dismiss hides the toast.
func (t Toast) Dismiss() Toast { t.left = 0; return t }

// Lines draws the toast as a small rounded box no wider than maxW cells.
func (t Toast) Lines(th theme.Theme, maxW int) []string {
	if !t.Shown() || maxW < 8 {
		return nil
	}
	icon, paint := th.Icons.Running, th.Info
	switch t.Level {
	case Success:
		icon, paint = th.Icons.Passed, th.Success
	case Warning:
		icon, paint = th.Icons.Warning, th.Warning
	case Error:
		icon, paint = th.Icons.Failed, th.Failure
	}
	text := " " + paint(icon) + " " + t.Text + " "
	w := min(pane.Width(text), maxW-2)
	text = pane.Fit(text, w)
	return []string{
		th.Border("╭"+strings.Repeat("─", w)+"╮", false),
		th.Border("│", false) + text + th.Border("│", false),
		th.Border("╰"+strings.Repeat("─", w)+"╯", false),
	}
}
