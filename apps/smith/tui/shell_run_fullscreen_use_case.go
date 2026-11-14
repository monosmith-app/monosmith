package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// fullRun is true while the output of the run fills the whole terminal under a
// top bar: no header, tabs, footer or status bar.
func (s Shell) fullRun() bool { return s.runFull && s.runWanted() }

// toggleFullRun makes the run output fill the terminal, or gives the screen back.
// The button of the run panel, the button of the top bar, a double click on the
// output and the key all come here.
func (s *Shell) toggleFullRun() {
	if !s.runFull && !s.runWanted() {
		return
	}
	s.runFull = !s.runFull
	s.selection = s.selection.Clear()
	s.keepRunVisible()
}

// fullBarButton is the button of the top bar, and the columns it fills.
func (s Shell) fullBarButton() (text string, first, end int) {
	state := button.Rest
	if s.mouseY == 0 {
		if first, end = s.fullBarSpan(); s.mouseX >= first && s.mouseX < end {
			state = button.Hover
		}
	}
	text = button.Render(s.th, words.Say("tui.run_full.shrink"), "", state)
	first, end = s.fullBarSpan()
	return text, first, end
}

// fullBarSpan is the columns of the button of the top bar, at its right end.
func (s Shell) fullBarSpan() (first, end int) {
	width := button.Width(s.th, words.Say("tui.run_full.shrink"), "")
	end = max(s.w-1, 0)
	return max(end-width, 0), end
}

// fullBar is the top row of the full-screen run: what runs, and the button that
// gives the screen back. Esc and a double click on the output do the same.
func (s Shell) fullBar() string {
	title := commands.RunTitle(s.run.Command, s.list.Len())
	if title == "" {
		title = s.run.CommandLine()
	}
	btn, first, _ := s.fullBarButton()
	left := " " + s.th.Title(title) + "  " + s.th.Muted(words.Say("tui.run_full.hint"))
	line := pane.Fit(left, first) + btn + " "
	return s.th.Bar(pane.Fit(line, s.w))
}

// inRunPanel is true when the cell (x, y) is in the run output, whether it is
// beside the list, over it or on the whole terminal.
func (s Shell) inRunPanel(x, y int) bool {
	switch {
	case s.fullRun():
		return y >= 1
	case !s.runShown():
		return false
	case s.lay().HasDetail():
		return s.detailInner().Contains(x, y)
	}
	return s.mainInner().Contains(x, y)
}
