package runhistory

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// NavPart is one part of the run header: its text, the action a click on it
// stands for ("" for plain text) and whether that action can be done now.
type NavPart struct {
	Text, Action string
	On           bool
}

// NavParts are the parts of the header of a run view: "‹ Start · run 2 of 5 ·
// ‹ older  newer ›". idle is false while a run goes, at is the position of the
// shown run (from 0) and total the runs kept. A part that cannot be used now is
// drawn faint; a click on it still says why nothing happened.
func NavParts(idle bool, at, total int) []NavPart {
	return []NavPart{
		{words.Say("tui.run_nav.start"), "start", idle},
		{" · ", "", false},
		{words.Say("tui.run_nav.run_of", "value", at+1, "len", total), "", false},
		{" · ", "", false},
		{words.Say("tui.run_nav.older"), "older", idle && at > 0},
		{"  ", "", false},
		{words.Say("tui.run_nav.newer"), "newer", idle && at < total-1},
	}
}

// ActionAt is the action of the header part under column x ("start", "older" or
// "newer"), or "" when x is on plain text. The header starts at column 1.
func ActionAt(parts []NavPart, x int) string {
	at := 1
	for _, p := range parts {
		w := pane.Width(p.Text)
		if x >= at && x < at+w {
			return p.Action
		}
		at += w
	}
	return ""
}
